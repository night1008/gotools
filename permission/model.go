package permission

// 基础权限
type Permission struct {
	Name     string `json:"name" yaml:"name" gorm:"primarykey;autoIncrement:false;size:256;"`               // 英文唯一标识
	Title    string `json:"title" yaml:"title"`                                                             // 中文标题
	Domain   string `json:"domain" yaml:"domain" gorm:"uniqueIndex:idx_permissions_resource;size:128;"`     // 可用于系统识别，比如 fa, fb
	Resource string `json:"resource" yaml:"resource" gorm:"uniqueIndex:idx_permissions_resource;size:256;"` // 比如 api/v1/posts
	Action   string `json:"action" yaml:"action" gorm:"uniqueIndex:idx_permissions_resource;size:64;"`      // 比如 get, post, put, delete 等

	CreatedAt int64 `gorm:"autoCreateTime:milli"`
}

// 权限组
type PermissionGroup struct {
	Name       string `json:"name" yaml:"name" gorm:"primaryKey;autoIncrement:false;size:256;"` // 英文唯一标识
	Domain     string `json:"domain" yaml:"domain"`                                             // 可用于菜单范围识别，比如团队，应用，空间
	Title      string `json:"title" yaml:"title"`                                               // 中文标题
	GroupIndex int    `json:"group_index" yaml:"group_index"`                                   // 用于菜单排序
	ParentName string `json:"parent_name" yaml:"parent_name"`                                   // 为空代表顶级菜单

	CreatedAt int64 `gorm:"autoCreateTime:milli"`
}

// 权限组和权限的关系
type PermissionGroupPermission struct {
	PermissionGroupName string `json:"permission_group_name" yaml:"permission_group_name" gorm:"primaryKey;autoIncrement:false;size:256;"`
	PermissionName      string `json:"permission_name" yaml:"permission_name" gorm:"primaryKey;autoIncrement:false;size:256;"`

	CreatedAt int64 `gorm:"autoCreateTime:milli"`
}

// 角色，使用 gorm polymorphic 机制
type Role struct {
	ID           int64  `json:"id" yaml:"id" gorm:"primarykey"`
	RoleableType string `json:"roleable_type" yaml:"roleable_type" gorm:"uniqueIndex:idx_roles_name;size:128;"` // 角色类型
	RoleableID   int64  `json:"roleable_id" yaml:"roleable_id" gorm:"uniqueIndex:idx_roles_name;"`              // 角色
	Name         string `json:"name" yaml:"name" gorm:"uniqueIndex:idx_roles_name;size:256;"`                   // 英文唯一标识

	Title         string `json:"title" yaml:"title"`                     // 中文标题
	Description   string `json:"description" yaml:"description"`         // 描述
	CreatorUserID int64  `json:"creator_user_id" yaml:"creator_user_id"` // 创建者ID

	CreatedAt int64 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli"`
}

// 角色拥有的权限组
type RolePermissionGroup struct {
	RoleID              int64  `json:"role_id" yaml:"role_id" gorm:"primaryKey;autoIncrement:false;"`
	PermissionGroupName string `json:"permission_group_name" yaml:"permission_group_name" gorm:"primaryKey;autoIncrement:false;size:256;"`

	CreatedAt int64 `gorm:"autoCreateTime:milli"`
}

// 用户和角色的关系，该包不进行用户创建，需要保证用户ID字段类型
type UserRole struct {
	UserID int64 `json:"user_id" yaml:"user_id" gorm:"primaryKey;autoIncrement:false;"`
	RoleID int64 `json:"role_id" yaml:"role_id" gorm:"primaryKey;autoIncrement:false;"`

	CreatedAt int64 `gorm:"autoCreateTime:milli"`
}
