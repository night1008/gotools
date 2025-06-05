package main

import (
	"context"
	"log"
	"os"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"git.sofunny.io/data-analysis/gotools/permission"
)

func main() {
	content, err := os.ReadFile("../metadata.yaml")
	if err != nil {
		panic(err)
	}
	var metadata permission.PermissionMetadata
	if err := yaml.Unmarshal(content, &metadata); err != nil {
		panic(err)
	}

	log.Printf("===> PermissionMetadata len(Permissions):%d, len(PermissionGroups):%d\n", len(metadata.Permissions), len(metadata.PermissionGroups))

	// CREATE DATABASE permission CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
	dsn := "root:secret@tcp(127.0.0.1:3306)/permission?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	svc := permission.New(db, &metadata)
	if err := svc.Migrate(); err != nil {
		panic(err)
	}
	ctx := context.Background()
	if err := svc.SyncPermissionMetadata(ctx); err != nil {
		panic(err)
	}

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

	var role permission.Role
	if err := db.First(&role).Error; err != nil {
		panic(err)
	}
	if err := svc.AssignRolesToUser(ctx, permission.AssignRolesToUserParam{
		UserID:       1,
		RoleableType: roleableType,
		RoleableID:   1,
		RoleIDs:      []int64{role.ID},
	}); err != nil {
		panic(err)
	}

	ok, err := svc.HasPermission(ctx, permission.HasPermissionParam{
		UserID:       1,
		RoleableType: roleableType,
		RoleableID:   1,
		Resource:     "/api/v1/apps/:id/access-keys",
		Action:       "GET",
	})
	if err != nil {
		panic(err)
	}
	log.Println("===> HasPermission ", ok)

	tree, err := svc.BuildFullPermissionGroupTree(ctx, "")
	if err != nil {
		panic(err)
	}
	log.Println("===> tree ", len(tree))
}
