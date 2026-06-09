package permission

import (
	"context"

	v1 "github.com/LandcLi/LandcPermission/api/permission/v1"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// PermissionController defines the permission management interface.
type PermissionController interface {
	// Permission CRUD
	CreatePermission(ctx context.Context, req *v1.CreatePermissionRequest) (*v1.CreatePermissionResponse, error)
	UpdatePermission(ctx context.Context, req *v1.UpdatePermissionRequest) (*v1.UpdatePermissionResponse, error)
	DeletePermission(ctx context.Context, req *v1.DeletePermissionRequest) (*v1.DeletePermissionResponse, error)
	GetPermission(ctx context.Context, req *v1.GetPermissionRequest) (*v1.GetPermissionResponse, error)
	ListPermissions(ctx context.Context, req *v1.ListPermissionsRequest) (*v1.ListPermissionsResponse, error)

	// Role-Permission binding
	AssignPermissions(ctx context.Context, req *v1.AssignPermissionsRequest) (*v1.AssignPermissionsResponse, error)
	RevokePermissions(ctx context.Context, req *v1.RevokePermissionsRequest) (*v1.RevokePermissionsResponse, error)
	GetRolePermissions(ctx context.Context, req *v1.GetRolePermissionsRequest) (*v1.GetRolePermissionsResponse, error)

	// User-Role binding
	AssignRoles(ctx context.Context, req *v1.AssignRolesRequest) (*v1.AssignRolesResponse, error)
	RevokeRoles(ctx context.Context, req *v1.RevokeRolesRequest) (*v1.RevokeRolesResponse, error)
	GetUserRoles(ctx context.Context, req *v1.GetUserRolesRequest) (*v1.GetUserRolesResponse, error)
	GetUserPermissions(ctx context.Context, req *v1.GetUserPermissionsRequest) (*v1.GetUserPermissionsResponse, error)

	// Permission check
	CheckPermission(ctx context.Context, req *v1.CheckPermissionRequest) (*v1.CheckPermissionResponse, error)

	// Permission check with context (支持多种权限模型)
	CheckPermissionWithContext(ctx context.Context, req *v1.CheckPermissionWithContextRequest) (*v1.CheckPermissionWithContextResponse, error)
}

// PermissionGateway is the service gateway for permission management.
var PermissionGateway = di.NewGateway[PermissionController]("permission.controller")

// GetPermissionController returns the registered PermissionController implementation.
func GetPermissionController() PermissionController {
	return PermissionGateway.Get()
}
