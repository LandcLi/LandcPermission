package sdk

import (
	"context"

	permV1 "github.com/LandcLi/LandcPermission/api/permission/v1"
	roleV1 "github.com/LandcLi/LandcPermission/api/role/v1"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// RemoteRoleController implements role.RoleController via HTTP calls.
type RemoteRoleController struct {
	client *di.ProxyClient
}

// NewRemoteRoleController creates a remote role controller.
func NewRemoteRoleController(baseURL string, opts ...di.RemoteOption) *RemoteRoleController {
	return &RemoteRoleController{
		client: di.NewProxyClient(baseURL, opts...),
	}
}

func (r *RemoteRoleController) CreateRole(ctx context.Context, req *roleV1.CreateRoleRequest) (*roleV1.CreateRoleResponse, error) {
	return di.Call[roleV1.CreateRoleResponse](r.client, ctx, "CreateRole", req)
}

func (r *RemoteRoleController) UpdateRole(ctx context.Context, req *roleV1.UpdateRoleRequest) (*roleV1.UpdateRoleResponse, error) {
	return di.Call[roleV1.UpdateRoleResponse](r.client, ctx, "UpdateRole", req)
}

func (r *RemoteRoleController) DeleteRole(ctx context.Context, req *roleV1.DeleteRoleRequest) (*roleV1.DeleteRoleResponse, error) {
	return di.Call[roleV1.DeleteRoleResponse](r.client, ctx, "DeleteRole", req)
}

func (r *RemoteRoleController) GetRole(ctx context.Context, req *roleV1.GetRoleRequest) (*roleV1.GetRoleResponse, error) {
	return di.Call[roleV1.GetRoleResponse](r.client, ctx, "GetRole", req)
}

func (r *RemoteRoleController) ListRoles(ctx context.Context, req *roleV1.ListRolesRequest) (*roleV1.ListRolesResponse, error) {
	return di.Call[roleV1.ListRolesResponse](r.client, ctx, "ListRoles", req)
}

// RemotePermissionController implements permission.PermissionController via HTTP calls.
type RemotePermissionController struct {
	client *di.ProxyClient
}

// NewRemotePermissionController creates a remote permission controller.
func NewRemotePermissionController(baseURL string, opts ...di.RemoteOption) *RemotePermissionController {
	return &RemotePermissionController{
		client: di.NewProxyClient(baseURL, opts...),
	}
}

func (r *RemotePermissionController) CreatePermission(ctx context.Context, req *permV1.CreatePermissionRequest) (*permV1.CreatePermissionResponse, error) {
	return di.Call[permV1.CreatePermissionResponse](r.client, ctx, "CreatePermission", req)
}

func (r *RemotePermissionController) UpdatePermission(ctx context.Context, req *permV1.UpdatePermissionRequest) (*permV1.UpdatePermissionResponse, error) {
	return di.Call[permV1.UpdatePermissionResponse](r.client, ctx, "UpdatePermission", req)
}

func (r *RemotePermissionController) DeletePermission(ctx context.Context, req *permV1.DeletePermissionRequest) (*permV1.DeletePermissionResponse, error) {
	return di.Call[permV1.DeletePermissionResponse](r.client, ctx, "DeletePermission", req)
}

func (r *RemotePermissionController) GetPermission(ctx context.Context, req *permV1.GetPermissionRequest) (*permV1.GetPermissionResponse, error) {
	return di.Call[permV1.GetPermissionResponse](r.client, ctx, "GetPermission", req)
}

func (r *RemotePermissionController) ListPermissions(ctx context.Context, req *permV1.ListPermissionsRequest) (*permV1.ListPermissionsResponse, error) {
	return di.Call[permV1.ListPermissionsResponse](r.client, ctx, "ListPermissions", req)
}

func (r *RemotePermissionController) AssignPermissions(ctx context.Context, req *permV1.AssignPermissionsRequest) (*permV1.AssignPermissionsResponse, error) {
	return di.Call[permV1.AssignPermissionsResponse](r.client, ctx, "AssignPermissions", req)
}

func (r *RemotePermissionController) RevokePermissions(ctx context.Context, req *permV1.RevokePermissionsRequest) (*permV1.RevokePermissionsResponse, error) {
	return di.Call[permV1.RevokePermissionsResponse](r.client, ctx, "RevokePermissions", req)
}

func (r *RemotePermissionController) GetRolePermissions(ctx context.Context, req *permV1.GetRolePermissionsRequest) (*permV1.GetRolePermissionsResponse, error) {
	return di.Call[permV1.GetRolePermissionsResponse](r.client, ctx, "GetRolePermissions", req)
}

func (r *RemotePermissionController) AssignRoles(ctx context.Context, req *permV1.AssignRolesRequest) (*permV1.AssignRolesResponse, error) {
	return di.Call[permV1.AssignRolesResponse](r.client, ctx, "AssignRoles", req)
}

func (r *RemotePermissionController) RevokeRoles(ctx context.Context, req *permV1.RevokeRolesRequest) (*permV1.RevokeRolesResponse, error) {
	return di.Call[permV1.RevokeRolesResponse](r.client, ctx, "RevokeRoles", req)
}

func (r *RemotePermissionController) GetUserRoles(ctx context.Context, req *permV1.GetUserRolesRequest) (*permV1.GetUserRolesResponse, error) {
	return di.Call[permV1.GetUserRolesResponse](r.client, ctx, "GetUserRoles", req)
}

func (r *RemotePermissionController) GetUserPermissions(ctx context.Context, req *permV1.GetUserPermissionsRequest) (*permV1.GetUserPermissionsResponse, error) {
	return di.Call[permV1.GetUserPermissionsResponse](r.client, ctx, "GetUserPermissions", req)
}

func (r *RemotePermissionController) CheckPermission(ctx context.Context, req *permV1.CheckPermissionRequest) (*permV1.CheckPermissionResponse, error) {
	return di.Call[permV1.CheckPermissionResponse](r.client, ctx, "CheckPermission", req)
}

// CheckPermissions 批量判定，一次问清多个权限，适合渲染菜单/按钮。
func (r *RemotePermissionController) CheckPermissions(ctx context.Context, req *permV1.CheckPermissionsRequest) (*permV1.CheckPermissionsResponse, error) {
	return di.Call[permV1.CheckPermissionsResponse](r.client, ctx, "CheckPermissions", req)
}

func (r *RemotePermissionController) CheckPermissionWithContext(ctx context.Context, req *permV1.CheckPermissionWithContextRequest) (*permV1.CheckPermissionWithContextResponse, error) {
	return di.Call[permV1.CheckPermissionWithContextResponse](r.client, ctx, "CheckPermissionWithContext", req)
}
