package role

import (
	"context"

	v1 "github.com/LandcLi/LandcPermission/api/role/v1"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// RoleController defines the role management interface.
type RoleController interface {
	CreateRole(ctx context.Context, req *v1.CreateRoleRequest) (*v1.CreateRoleResponse, error)
	UpdateRole(ctx context.Context, req *v1.UpdateRoleRequest) (*v1.UpdateRoleResponse, error)
	DeleteRole(ctx context.Context, req *v1.DeleteRoleRequest) (*v1.DeleteRoleResponse, error)
	GetRole(ctx context.Context, req *v1.GetRoleRequest) (*v1.GetRoleResponse, error)
	ListRoles(ctx context.Context, req *v1.ListRolesRequest) (*v1.ListRolesResponse, error)
}

// RoleGateway is the service gateway for role management.
var RoleGateway = di.NewGateway[RoleController]("role.controller")

// GetRoleController returns the registered RoleController implementation.
func GetRoleController() RoleController {
	return RoleGateway.Get()
}
