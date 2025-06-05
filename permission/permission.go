package permission

import (
	"context"
	"embed"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:embed migrations
var migrationFs embed.FS

type PermissionServiceOption func(*PermissionService)

type PermissionItem struct {
	Name     string `json:"name" yaml:"name"`
	Title    string `json:"title" yaml:"title"`
	Domain   string `json:"domain" yaml:"domain"`
	Resource string `json:"resource" yaml:"resource"`
	Action   string `json:"action" yaml:"action"`
}

type PermissionGroupItem struct {
	Name             string                 `json:"name" yaml:"name"`
	Domain           string                 `json:"domain,omitempty" yaml:"domain,omitempty"`
	Title            string                 `json:"title" yaml:"title"`
	Permissions      []string               `json:"permissions,omitempty" yaml:"permissions,omitempty"`
	PermissionGroups []*PermissionGroupItem `json:"permission_groups,omitempty" yaml:"permission_groups,omitempty"`
}

type RolePermissionGroupItem struct {
	RoleableType     string   `json:"roleable_type" yaml:"roleable_type"`
	Name             string   `json:"name" yaml:"name"`
	Title            string   `json:"title" yaml:"title"`
	Description      string   `json:"description" yaml:"description"`
	PermissionGroups []string `json:"permission_groups" yaml:"permission_groups"`
}

type PermissionMetadata struct {
	Permissions      []*PermissionItem          `json:"permissions" yaml:"permissions"`
	PermissionGroups []*PermissionGroupItem     `json:"permission_groups" yaml:"permission_groups"`
	Roles            []*RolePermissionGroupItem `json:"roles" yaml:"roles"`
}

type PermissionService struct {
	db       *gorm.DB
	metadata *PermissionMetadata

	cachedTableNames struct {
		permissionTableName                string
		permissionGroupTableName           string
		permissionGroupPermissionTableName string
		roleTableName                      string
		rolePermissionGroupTableName       string
		userRoleTableName                  string
	} // 缓存表名用于构造自定义查询语句，应对表名规则调整的情况
}

func New(db *gorm.DB, metadata *PermissionMetadata, opts ...PermissionServiceOption) *PermissionService {
	s := &PermissionService{
		db:       db,
		metadata: metadata,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.cacheTableNames()
	return s
}

func (s *PermissionService) cacheTableNames() {
	s.cachedTableNames.permissionTableName = s.db.Config.NamingStrategy.TableName("Permission")
	s.cachedTableNames.permissionGroupTableName = s.db.Config.NamingStrategy.TableName("PermissionGroup")
	s.cachedTableNames.permissionGroupPermissionTableName = s.db.Config.NamingStrategy.TableName("PermissionGroupPermission")
	s.cachedTableNames.roleTableName = s.db.Config.NamingStrategy.TableName("Role")
	s.cachedTableNames.rolePermissionGroupTableName = s.db.Config.NamingStrategy.TableName("RolePermissionGroup")
	s.cachedTableNames.userRoleTableName = s.db.Config.NamingStrategy.TableName("UserRole")
}

// 数据库表结构迁移
func (s *PermissionService) Migrate() error {
	return s.db.AutoMigrate(&Permission{}, &PermissionGroup{}, &PermissionGroupPermission{}, &Role{}, &RolePermissionGroup{}, &UserRole{})
}

// 输出数据库表结构迁移语句
func (s *PermissionService) GetMigrateStatements() (string, error) {
	dialectorName := s.db.Dialector.Name()
	migrationFileName := fmt.Sprintf("migrations/%s.sql", dialectorName)
	switch dialectorName {
	case "postgres", "mysql", "sqlite":
		content, err := migrationFs.ReadFile(migrationFileName)
		if err != nil {
			return "", err
		}
		return string(content), nil
	default:
		return "", fmt.Errorf("unsupported dialector name %s", dialectorName)
	}
}

// 同步权限元数据
func (s *PermissionService) SyncPermissionMetadata(ctx context.Context) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.syncPermissions(tx); err != nil {
			return err
		}
		if err := s.syncPermissionGroups(tx); err != nil {
			return err
		}
		return nil
	})
}

// 同步基础权限
func (s *PermissionService) syncPermissions(tx *gorm.DB) error {
	permissionResourceActionKeysMap := make(map[string]struct{}, len(s.metadata.Permissions))
	permissionKeysMap := make(map[string]struct{}, len(s.metadata.Permissions))
	permissions := make([]*Permission, 0, len(s.metadata.Permissions))
	for _, p := range s.metadata.Permissions {
		resourceActionKey := fmt.Sprintf("%s_%s_%s", p.Domain, p.Resource, p.Action)
		if _, ok := permissionResourceActionKeysMap[resourceActionKey]; ok {
			return fmt.Errorf("permission domain:%s + resource:%s + action:%s reduplicated", p.Domain, p.Resource, p.Action)
		} else {
			permissionResourceActionKeysMap[resourceActionKey] = struct{}{}
		}
		permissionKey := p.Name
		if _, ok := permissionKeysMap[permissionKey]; ok {
			return fmt.Errorf("permission name:%s reduplicated", p.Name)
		} else {
			permissionKeysMap[permissionKey] = struct{}{}
		}
		permissions = append(permissions, &Permission{
			Name:     p.Name,
			Title:    p.Title,
			Domain:   p.Domain,
			Resource: p.Resource,
			Action:   p.Action,
		})
	}

	var existedPermissionKeys []string
	if err := tx.Model(&Permission{}).Pluck("name", &existedPermissionKeys).Error; err != nil {
		return err
	}

	// 比较已存在权限和最新读取的权限列表，删除不存在最新列表的权限
	var needDeleteKeys []string
	for _, key := range existedPermissionKeys {
		if _, ok := permissionKeysMap[key]; !ok {
			needDeleteKeys = append(needDeleteKeys, key)
		}
	}

	if len(needDeleteKeys) > 0 {
		if err := tx.Where("name IN ?", needDeleteKeys).Delete(&Permission{}).Error; err != nil {
			return err
		}
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "domain", "resource", "action"}),
	}).Create(permissions).Error; err != nil {
		return err
	}

	return nil
}

// 同步权限组中间状态
type syncPermissionGroupIntermediateState struct {
	existedPermissionsMap                map[string]*Permission
	existedPermissionGroupPermissionsMap map[string][]*PermissionGroupPermission

	permissionGroupKeys        []string
	permissionGroups           []*PermissionGroup
	permissionGroupPermissions []*PermissionGroupPermission
}

// 同步权限组
func (s *PermissionService) syncPermissionGroups(tx *gorm.DB) error {
	var intermediateState syncPermissionGroupIntermediateState
	var permissions []*Permission
	if err := tx.Find(&permissions).Error; err != nil {
		return err
	}
	intermediateState.existedPermissionsMap = make(map[string]*Permission, len(permissions))
	for _, p := range permissions {
		intermediateState.existedPermissionsMap[p.Name] = p
	}

	var permissionGroupPermissions []*PermissionGroupPermission
	if err := tx.Find(&permissionGroupPermissions).Error; err != nil {
		return err
	}
	intermediateState.existedPermissionGroupPermissionsMap = make(map[string][]*PermissionGroupPermission, len(permissions))
	for _, p := range permissionGroupPermissions {
		intermediateState.existedPermissionGroupPermissionsMap[p.PermissionGroupName] = append(intermediateState.existedPermissionGroupPermissionsMap[p.PermissionGroupName], p)
	}

	// 递归生成权限组
	for i, g := range s.metadata.PermissionGroups {
		if err := s.createPermissionGroup(tx, g, i, "", &intermediateState); err != nil {
			return err
		}
	}

	var existedPermissionGroupKeys []string
	if err := tx.Model(&PermissionGroup{}).Pluck("name", &existedPermissionGroupKeys).Error; err != nil {
		return err
	}

	// 比较已存在权限和最新读取的权限列表，删除不存在最新列表的权限
	permissionGroupKeysMap := make(map[string]struct{}, len(intermediateState.permissionGroupKeys))
	for _, key := range intermediateState.permissionGroupKeys {
		permissionGroupKeysMap[key] = struct{}{}
	}
	var needDeletePermissionGroupKeys []string
	for _, key := range existedPermissionGroupKeys {
		if _, ok := permissionGroupKeysMap[key]; !ok {
			needDeletePermissionGroupKeys = append(needDeletePermissionGroupKeys, key)
		}
	}

	if len(needDeletePermissionGroupKeys) > 0 {
		if err := tx.Where("name IN ?", needDeletePermissionGroupKeys).Delete(&PermissionGroup{}).Error; err != nil {
			return err
		}
		if err := tx.Where("permission_group_name IN ?", needDeletePermissionGroupKeys).Delete(&PermissionGroupPermission{}).Error; err != nil {
			return err
		}
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "domain", "group_index", "parent_name"}),
	}).Create(intermediateState.permissionGroups).Error; err != nil {
		return err
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "permission_group_name"}, {Name: "permission_name"}},
		DoNothing: true,
	}).Create(intermediateState.permissionGroupPermissions).Error; err != nil {
		return err
	}
	return nil
}

func (s *PermissionService) createPermissionGroup(tx *gorm.DB, g *PermissionGroupItem, groupIndex int, parentName string, intermediateState *syncPermissionGroupIntermediateState) error {
	permissionGroup := &PermissionGroup{
		Name:       g.Name,
		Domain:     g.Domain,
		Title:      g.Title,
		GroupIndex: groupIndex,
		ParentName: parentName,
	}
	intermediateState.permissionGroupKeys = append(intermediateState.permissionGroupKeys, g.Name)
	intermediateState.permissionGroups = append(intermediateState.permissionGroups, permissionGroup)

	if len(g.Permissions) > 0 {
		notExistedPermissionNames := make([]string, 0, len(g.Permissions))
		permissionKeysMap := make(map[string]struct{}, len(g.Permissions))
		for _, permissionKey := range g.Permissions {
			permissionKeysMap[permissionKey] = struct{}{}
			if _, ok := intermediateState.existedPermissionsMap[permissionKey]; !ok {
				notExistedPermissionNames = append(notExistedPermissionNames, permissionKey)
			}
		}
		if len(notExistedPermissionNames) > 0 {
			return fmt.Errorf("permission group num error: permission_group_name:%s notExistedPermissionNames: %v", g.Name, notExistedPermissionNames)
		}

		var needDeletePermissionNames []string
		for _, p := range intermediateState.existedPermissionGroupPermissionsMap[g.Name] {
			if _, ok := permissionKeysMap[p.PermissionName]; !ok {
				needDeletePermissionNames = append(needDeletePermissionNames, p.PermissionName)
			}
		}
		// 最新权限组中的权限不包含已存在权限，删除已存在权限
		if len(needDeletePermissionNames) > 0 {
			if err := tx.Where("permission_group_name = ?", g.Name).Where("permission_name IN ?", needDeletePermissionNames).Delete(&PermissionGroupPermission{}).Error; err != nil {
				return err
			}
		}

		groupPermissions := make([]*PermissionGroupPermission, len(g.Permissions))
		for i, permissionKey := range g.Permissions {
			groupPermissions[i] = &PermissionGroupPermission{
				PermissionGroupName: permissionGroup.Name,
				PermissionName:      permissionKey,
			}
		}
		intermediateState.permissionGroupPermissions = append(intermediateState.permissionGroupPermissions, groupPermissions...)
	}

	// 递归生成权限组
	for i, g2 := range g.PermissionGroups {
		if err := s.createPermissionGroup(tx, g2, i, permissionGroup.Name, intermediateState); err != nil {
			return err
		}
	}
	return nil
}

// 同步某个应用下的预置角色
func (s *PermissionService) SyncPresetRoles(tx *gorm.DB, roleableID int64, roleableType string) error {
	for _, roleGroups := range s.metadata.Roles {
		if roleGroups.RoleableType != roleableType {
			continue
		}

		role := &Role{
			RoleableType: roleableType,
			RoleableID:   roleableID,
			Name:         roleGroups.Name,
			Title:        roleGroups.Title,
			Description:  roleGroups.Description,
		}
		if err := tx.FirstOrCreate(role, &Role{
			RoleableType: roleableType,
			RoleableID:   roleableID,
			Name:         roleGroups.Name,
		}).Error; err != nil {
			return err
		}
		if len(roleGroups.PermissionGroups) > 0 {
			rolePermissionGroups := make([]*RolePermissionGroup, 0, len(roleGroups.PermissionGroups))
			for _, groupKey := range roleGroups.PermissionGroups {
				rolePermissionGroups = append(rolePermissionGroups, &RolePermissionGroup{
					RoleID:              role.ID,
					PermissionGroupName: groupKey,
				})
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "role_id"}, {Name: "permission_group_name"}},
				DoNothing: true,
			}).Create(rolePermissionGroups).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

type CreateRoleParam struct {
	RoleableType     string   `json:"roleable_type" yaml:"roleable_type"`
	RoleableID       int64    `json:"roleable_id" yaml:"roleable_id"`
	Name             string   `json:"name" yaml:"name"`
	Title            string   `json:"title" yaml:"title"`
	Description      string   `json:"description" yaml:"description"`
	PermissionGroups []string `json:"permission_groups" yaml:"permission_groups"`
	CreatorUserID    int64    `json:"creator_user_id" yaml:"creator_user_id"`
}

// 创建角色
func (s *PermissionService) CreateRole(ctx context.Context, param CreateRoleParam) (*Role, error) {
	role := Role{
		Name:          param.Name,
		RoleableType:  param.RoleableType,
		Title:         param.Title,
		Description:   param.Description,
		RoleableID:    param.RoleableID,
		CreatorUserID: param.CreatorUserID,
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.FirstOrCreate(&role, Role{
			RoleableType: param.RoleableType,
			RoleableID:   param.RoleableID,
			Name:         param.Name,
		}).Error; err != nil {
			return err
		}
		if err := s.assignPermissionGroupsToRole(tx, role.ID, param.PermissionGroups); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &role, nil
}

type UpdateRoleParam struct {
	ID               int64    `json:"id" yaml:"id"`
	Title            string   `json:"title" yaml:"title"`
	Description      string   `json:"description" yaml:"description"`
	PermissionGroups []string `json:"permission_groups" yaml:"permission_groups"`
}

// 更新角色
func (s *PermissionService) UpdateRole(ctx context.Context, param UpdateRoleParam) (*Role, error) {
	var role Role
	if err := s.db.WithContext(ctx).Model(&Role{}).Where("id = ?", param.ID).First(&role).Error; err != nil {
		return nil, err
	}

	role.Title = param.Title
	role.Description = param.Description
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&role).Error; err != nil {
			return err
		}
		if err := s.assignPermissionGroupsToRole(tx, role.ID, param.PermissionGroups); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &role, nil
}

// 删除角色
func (s *PermissionService) DeleteRole(ctx context.Context, roleID int64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&RolePermissionGroup{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", roleID).Delete(&UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", roleID).Delete(&Role{}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// 为角色分配权限组
func (s *PermissionService) assignPermissionGroupsToRole(tx *gorm.DB, roleID int64, permissionGroupNames []string) error {
	var permissionGroups []*PermissionGroup
	if err := tx.Where("name IN ?", permissionGroupNames).Find(&permissionGroups).Error; err != nil {
		return err
	}

	if len(permissionGroups) != len(permissionGroupNames) {
		return fmt.Errorf("some permission groups not found")
	}

	rolePermissionGroups := make([]*RolePermissionGroup, 0, len(permissionGroups))
	for _, key := range permissionGroupNames {
		rolePermissionGroups = append(rolePermissionGroups, &RolePermissionGroup{
			RoleID:              roleID,
			PermissionGroupName: key,
		})
	}
	if len(permissionGroupNames) == 0 {
		return fmt.Errorf("role must have at least one permission groups")
	}

	if err := tx.Where("role_id = ?", roleID).Delete(&RolePermissionGroup{}).Error; err != nil {
		return err
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "role_id"}, {Name: "permission_group_name"}},
		DoNothing: true,
	}).Create(&rolePermissionGroups).Error; err != nil {
		return err
	}
	return nil
}

type AssignRolesToUserParam struct {
	UserID       int64   `json:"user_id" yaml:"user_id"`
	RoleableType string  `json:"roleable_type" yaml:"roleable_type"`
	RoleableID   int64   `json:"roleable_id" yaml:"roleable_id"`
	RoleIDs      []int64 `json:"role_ids" yaml:"role_ids"`
}

// 为用户分配角色
func (s *PermissionService) AssignRolesToUser(ctx context.Context, param AssignRolesToUserParam) error {
	var roles []*Role
	if err := s.db.WithContext(ctx).Where("roleable_type = ?", param.RoleableType).
		Where("roleable_id = ?", param.RoleableID).
		Find(&roles).Error; err != nil {
		return err
	}
	rolesMap := make(map[int64]*Role, len(roles))
	roleIDs := make([]int64, 0, len(roles))
	for _, r := range roles {
		rolesMap[r.ID] = r
		roleIDs = append(roleIDs, r.ID)
	}

	userRoles := make([]*UserRole, 0, len(param.RoleIDs))
	for _, roleID := range param.RoleIDs {
		if _, ok := rolesMap[roleID]; !ok {
			return fmt.Errorf("role id %d not found in %s:%d", roleID, param.RoleableType, param.RoleableID)
		}
		userRoles = append(userRoles, &UserRole{
			UserID: param.UserID,
			RoleID: roleID,
		})
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(roleIDs) > 0 {
			if err := tx.Where("user_id = ?", param.UserID).
				Where("role_id IN (?)", tx.Model(&Role{}).Select("id").Where("roleable_type = ?", param.RoleableType).Where("roleable_id = ?", param.RoleableID)).
				Delete(&UserRole{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "role_id"}},
			DoNothing: true,
		}).Create(userRoles).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

type HasPermissionParam struct {
	UserID       int64  `json:"user_id" yaml:"user_id"`
	RoleableType string `json:"roleable_type" yaml:"roleable_type"`
	RoleableID   int64  `json:"roleable_id" yaml:"roleable_id"`
	Domain       string `json:"domain" yaml:"domain"`
	Resource     string `json:"resource" yaml:"resource"`
	Action       string `json:"action" yaml:"action"`
}

// 检查用户是否有特定权限
func (s *PermissionService) HasPermission(ctx context.Context, param HasPermissionParam) (bool, error) {
	sql := fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE name IN (
		SELECT permission_name FROM %s WHERE permission_group_name IN (
			SELECT permission_group_name FROM %s WHERE role_id IN (
				SELECT id FROM %s WHERE roleable_type = ? AND roleable_id = ? AND id IN (
					SELECT role_id FROM %s WHERE user_id = ?
				)
			)
		)
	) AND domain = ? AND resource = ? AND action = ?`,
		s.cachedTableNames.permissionTableName,
		s.cachedTableNames.permissionGroupPermissionTableName,
		s.cachedTableNames.rolePermissionGroupTableName,
		s.cachedTableNames.roleTableName,
		s.cachedTableNames.userRoleTableName)

	var count int64
	if err := s.db.WithContext(ctx).Raw(sql, param.RoleableType, param.RoleableID, param.UserID, param.Domain, param.Resource, param.Action).Scan(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

type HasPermissionGroupParam struct {
	UserID              int64  `json:"user_id" yaml:"user_id"`
	RoleableType        string `json:"roleable_type" yaml:"roleable_type"`
	RoleableID          int64  `json:"roleable_id" yaml:"roleable_id"`
	PermissionGroupName string `json:"permission_group_name" yaml:"permission_group_name"`
}

// 检查用户在某个对象下是否拥有某个权限组
func (s *PermissionService) HasPermissionGroup(ctx context.Context, param HasPermissionGroupParam) (bool, error) {
	sql := fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE role_id IN (
		SELECT id FROM %s WHERE roleable_type = ? AND roleable_id = ? AND id IN (
			SELECT role_id FROM %s WHERE user_id = ?
		)
	) AND permission_group_name = ?`,
		s.cachedTableNames.rolePermissionGroupTableName,
		s.cachedTableNames.roleTableName,
		s.cachedTableNames.userRoleTableName)

	var count int64
	if err := s.db.WithContext(ctx).Raw(sql, param.RoleableType, param.RoleableID, param.UserID, param.PermissionGroupName).Scan(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

type HasPermissionGroupsParam struct {
	UserID               int64    `json:"user_id" yaml:"user_id"`
	RoleableType         string   `json:"roleable_type" yaml:"roleable_type"`
	RoleableID           int64    `json:"roleable_id" yaml:"roleable_id"`
	PermissionGroupNames []string `json:"permission_group_names" yaml:"permission_group_names"`
}

// 检查用户在某个对象下权限组列表拥有情况
func (s *PermissionService) HasPermissionGroups(ctx context.Context, param HasPermissionGroupsParam) (map[string]bool, error) {
	if len(param.PermissionGroupNames) == 0 {
		return nil, nil
	}

	sql := fmt.Sprintf(`SELECT DISTINCT permission_group_name FROM %s WHERE role_id IN (
		SELECT id FROM %s WHERE roleable_type = ? AND roleable_id = ? AND id IN (
			SELECT role_id FROM %s WHERE user_id = ?
		)
	) AND permission_group_name IN ?`,
		s.cachedTableNames.rolePermissionGroupTableName,
		s.cachedTableNames.roleTableName,
		s.cachedTableNames.userRoleTableName)

	var existedPermissionGroupKeys []string
	if err := s.db.WithContext(ctx).Raw(sql, param.RoleableType, param.RoleableID, param.UserID, param.PermissionGroupNames).Scan(&existedPermissionGroupKeys).Error; err != nil {
		return nil, err
	}
	existedPermissionGroupKeysMap := make(map[string]struct{}, len(existedPermissionGroupKeys))
	for _, key := range existedPermissionGroupKeys {
		existedPermissionGroupKeysMap[key] = struct{}{}
	}
	permissionGroupKeysMap := make(map[string]bool, len(param.PermissionGroupNames))
	for _, key := range param.PermissionGroupNames {
		_, ok := existedPermissionGroupKeysMap[key]
		permissionGroupKeysMap[key] = ok
	}
	return permissionGroupKeysMap, nil
}

// 应用下是否有任意角色
func (s *PermissionService) HasAnyRole(ctx context.Context, userID, roleableID int64, roleableType string) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&UserRole{}).
		Where("user_id = ?", userID).
		Where("role_id IN (?)", s.db.Model(&Role{}).Select("id").Where("roleable_type = ?", roleableType).Where("roleable_id = ?", roleableID)).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// 获取角色列表
func (s *PermissionService) GetRoles(ctx context.Context, roleableID int64, roleableType string) ([]*Role, error) {
	var roles []*Role
	if err := s.db.WithContext(ctx).Model(&Role{}).
		Where("roleable_type = ?", roleableType).
		Where("roleable_id = ?", roleableID).
		Order("id").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// 获取角色权限组 name 列表
func (s *PermissionService) GetRolePermissionGroupNames(ctx context.Context, roleID int64) ([]string, error) {
	var permissionGroupNames []string
	if err := s.db.WithContext(ctx).Model(&RolePermissionGroup{}).
		Select("permission_group_name").Where("role_id = ?", roleID).
		Pluck("permission_group_name", &permissionGroupNames).Error; err != nil {
		return nil, err
	}

	return permissionGroupNames, nil
}

// 获取角色权限组
func (s *PermissionService) GetRolePermissionGroups(ctx context.Context, roleID int64) ([]*PermissionGroup, error) {
	var permissionGroups []*PermissionGroup
	if err := s.db.WithContext(ctx).Model(&PermissionGroup{}).
		Where("name IN (?)", s.db.Model(&RolePermissionGroup{}).Select("permission_group_name").Where("role_id = ?", roleID)).
		Order("group_index").Find(&permissionGroups).Error; err != nil {
		return nil, err
	}

	return permissionGroups, nil
}

// 获取角色列表
func (s *PermissionService) GetUserRoles(ctx context.Context, userID, roleableID int64, roleableType string) ([]*Role, error) {
	var roles []*Role
	if err := s.db.WithContext(ctx).Model(&Role{}).
		Where("roleable_type = ?", roleableType).
		Where("roleable_id = ?", roleableID).
		Where("id IN (?)", s.db.Model(&UserRole{}).Select("role_id").Where("user_id = ?", userID)).
		Order("id").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// 获取用户应用ID列表
func (s *PermissionService) GetUserRoleableIDs(ctx context.Context, userID int64, roleableType string) ([]int64, error) {
	var roleableIDs []int64
	if err := s.db.WithContext(ctx).Model(&Role{}).
		Distinct("roleable_id").
		Where("roleable_type = ?", roleableType).
		Where("id IN (?)", s.db.Model(&UserRole{}).Select("role_id").Where("user_id = ?", userID)).
		Pluck("roleable_id", &roleableIDs).Error; err != nil {
		return nil, err
	}
	return roleableIDs, nil
}

// 获取用户应用ID列表组合
func (s *PermissionService) GetUserRoleableIDsMap(ctx context.Context, userID int64, roleableTypes ...string) (map[string][]int64, error) {
	if len(roleableTypes) == 0 {
		return nil, nil
	}

	var roles []*Role
	if err := s.db.WithContext(ctx).Model(&Role{}).
		Where("roleable_type IN ?", roleableTypes).
		Where("id IN (?)", s.db.Model(&UserRole{}).Select("role_id").Where("user_id = ?", userID)).
		Find(&roles).Error; err != nil {
		return nil, err
	}
	roleableIDsMap := make(map[string][]int64, len(roleableTypes))
	for _, role := range roles {
		var hasRoleableID bool
		for _, roleableID := range roleableIDsMap[role.RoleableType] {
			if roleableID == role.RoleableID {
				hasRoleableID = true
				break
			}
		}
		if !hasRoleableID {
			roleableIDsMap[role.RoleableType] = append(roleableIDsMap[role.RoleableType], role.RoleableID)
		}
	}
	return roleableIDsMap, nil
}

// 根据某个 domain 下所有权限组构造完整的权限树
func (s *PermissionService) BuildFullPermissionGroupTree(ctx context.Context, domain string) ([]*PermissionGroupItem, error) {
	var permissionGroups []*PermissionGroup
	if err := s.db.WithContext(ctx).Model(&PermissionGroup{}).
		Where("domain = ?", domain).
		Order("group_index").
		Find(&permissionGroups).Error; err != nil {
		return nil, err
	}

	tree := s.BuildPermissionGroupTree(permissionGroups, "")
	return tree, nil
}

// 根据权限组构造权限树
func (s *PermissionService) BuildPermissionGroupTree(permissionGroups []*PermissionGroup, parentName string) []*PermissionGroupItem {
	var tree []*PermissionGroupItem
	for _, group := range permissionGroups {
		item := &PermissionGroupItem{
			Name:   group.Name,
			Domain: group.Domain,
			Title:  group.Title,
		}
		if group.ParentName == parentName {
			item.PermissionGroups = s.BuildPermissionGroupTree(permissionGroups, group.Name)
			tree = append(tree, item)
		}
	}
	return tree
}
