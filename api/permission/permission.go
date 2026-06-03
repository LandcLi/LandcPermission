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
}

// PermissionRoutes maps controller methods to HTTP endpoints.
var PermissionRoutes = di.Routes{
	"CreatePermission": {Method: "POST", Path: "/api/permission/create"},
	"UpdatePermission": {Method: "POST", Path: "/api/permission/update"},
	"DeletePermission": {Method: "POST", Path: "/api/permission/delete"},
	"GetPermission":    {Method: "POST", Path: "/api/permission/get"},
	"ListPermissions":  {Method: "POST", Path: "/api/permission/list"},

	"AssignPermissions": {Method: "POST", Path: "/api/permission/assign-to-role"},
	"RevokePermissions": {Method: "POST", Path: "/api/permission/revoke-from-role"},
	"GetRolePermissions": {Method: "POST", Path: "/api/permission/get-role-permissions"},

	"AssignRoles":      {Method: "POST", Path: "/api/permission/assign-roles"},
	"RevokeRoles":      {Method: "POST", Path: "/api/permission/revoke-roles"},
	"GetUserRoles":     {Method: "POST", Path: "/api/permission/get-user-roles"},
	"GetUserPermissions": {Method: "POST", Path: "/api/permission/get-user-permissions"},

	"CheckPermission": {Method: "POST", Path: "/api/permission/check"},
}

// PermissionGateway is the service gateway for permission management.
var PermissionGateway = di.NewGateway[PermissionController]("permission.controller", PermissionRoutes)

// GetPermissionController returns the registered PermissionController implementation.
func GetPermissionController() PermissionController {
	return PermissionGateway.Get()
}
