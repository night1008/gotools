package permission

import (
	"context"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _permissionSvc *PermissionService

func TestMain(m *testing.M) {
	content, err := os.ReadFile("./examples/metadata.yaml")
	if err != nil {
		panic(err)
	}
	var metadata PermissionMetadata
	if err := yaml.Unmarshal(content, &metadata); err != nil {
		panic(err)
	}

	db, err := gorm.Open(sqlite.Open("permission.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	_permissionSvc = New(db, &metadata)
	if err := _permissionSvc.Migrate(); err != nil {
		panic(err)
	}
	ctx := context.Background()
	if err := _permissionSvc.SyncPermissionMetadata(ctx); err != nil {
		panic(err)
	}

	code := m.Run()

	_ = os.Remove("permission.db")

	os.Exit(code)
}

func TestPermissionService_HasPermission(t *testing.T) {
	type args struct {
		ctx   context.Context
		param HasPermissionParam
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "apps-meta-events-get",
			args: args{
				ctx: context.Background(),
				param: HasPermissionParam{
					UserID:       1,
					RoleableType: "app",
					RoleableID:   1,
					Domain:       "",
					Resource:     "/api/v1/apps/:id/meta/events",
					Action:       "GET",
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := _permissionSvc.HasPermission(tt.args.ctx, tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("PermissionService.HasPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PermissionService.HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}
