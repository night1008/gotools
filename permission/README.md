# 权限模块

> 基于 gorm 实现 RBAC 权限模块
> 支持三种数据库 postgres | mysql | sqlite


### 权限模型关系图

```
                                    ┌──────────┐
                                    │          │
                                    │   App    │
                                    │          │
                                    └──────────┘
                                          ▲
                                          │
                                          │
                                          │
         ┌──────────┐               ┌─────┴────┐                   ┌──────────────────────┐                ┌───────────────┐
         │          │               │          │                   │                      │                │               │
         │   User   │               │   Role   │             ┌─────┤    PermissionGroup   ├──────┐   ┌─────┤   Permission  │
         │          │               │          │             │     │                      │      │   │     │               │
         └─────┬────┘               └────┬─┬───┘             │     └──────────────────────┘      │   │     └───────────────┘
               │                         │ │                 │                                   │   │
               │                         │ │                 │                                   │   │
               │                         │ │                 │                                   │   │
               │                         │ │                 ▼                                   ▼   ▼
               │     ┌────────────┐      │ │      ┌──────────────────────┐            ┌────────────────────────────┐
               │     │            │      │ │      │                      │            │                            │
               └────►│  UserRole  │◄─────┘ └─────►│  RolePermissionGroup │            │  PermissionGroupPermission │
                     │            │               │                      │            │                            │
                     └────────────┘               └──────────────────────┘            └────────────────────────────┘
```

### 表名介绍
> 该包会创建以下几张表
> 1. 该包不维护以上关系图中的 User(用户)，只需保证 `user_id` 为 `int64`
> 2. 该包不维护以上关系图中的 App(应用)，实际在 `roles` 表使用了 gorm polymorphic，表字段为 `roleable_type` 和 `roleable_id`，也就是可以对团队或组织之类的对象构建权限组
> 3. 没有使用 `key` 作为权限的主键字段，因为是 `mysql` 的关键字

| 表名    | 标题 |
| -------- | ------- |
| Permission  | 基础权限    |
| PermissionGroup | 权限组     |
| PermissionGroupPermission    | 权限组和基础权限的关联关系    |
| Role  | 角色    |
| RolePermissionGroup | 角色下的权限组     |
| UserRole    | 用户和角色的关联关系    |

---

### 注意事项
1. 每个系统共用所有权限，暂时不可以单个应用下实现权限差异化
2. 需要自行保证 `permissions` 和 `permission_groups` 的 `name` 的唯一性
3. `permissions` 的 `domain` 可用于指定不同系统，比如 fa 或 fb
4. `permission_groups` 的 `domain` 可用于指定不同对象，比如团队或应用

### 代码示例

元数据组织格式可参考文件 [examples/metadata.yaml](./examples/metadata.yaml)

代码示例在目录 [examples](./examples) 下

```go
dsn := "host=localhost user=postgres password=secret dbname=permission port=5432 sslmode=disable TimeZone=Asia/Shanghai"
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
if err != nil {
  panic(err)
}

svc := gopermission.New(db, &metadata)
// 使用 gorm 的迁移机制，如果使用其他迁移机制，可以不执行以下方法
if err := svc.Migrate(); err != nil {
  panic(err)
}

ctx := context.Background()

// 同步权限元数据
if err := svc.SyncPermissionMetadata(ctx); err != nil {
  panic(err)
}
// 同步应用预置角色
roleableType := "app"
if err := db.Transaction(func(tx *gorm.DB) error {
  for i := 1; i <= 3; i++ {
    if err := svc.SyncPresetRoles(tx, int64(i), roleableType); err != nil {
      return err
    }
  }
  return nil
}); err != nil {
  panic(err)
}

// 判断是否有应用权限
ok, err := svc.HasPermission(ctx, HasPermissionParam{
  UserID:       1,
  RoleableType: roleableType,
  RoleableID:   1,
  Resource:     "/api/v1/apps/:id/access-keys",
  Action:       "GET",
})
if err != nil {
  panic(err)
}
fmt.Println("===> HasPermission ", ok)
```