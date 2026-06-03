package role

import (
	"context"

	roleApi "github.com/LandcLi/LandcPermission/api/role"
	v1 "github.com/LandcLi/LandcPermission/api/role/v1"
	"github.com/LandcLi/LandcPermission/service"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

func init() {
	di.Provide[roleApi.RoleController]("role.controller", &roleControllerImpl{})
}

type roleControllerImpl struct{}

func (c *roleControllerImpl) CreateRole(ctx context.Context, req *v1.CreateRoleRequest) (*v1.CreateRoleResponse, error) {
	role, err := service.GetRoleService().CreateRole(req.Name, req.Code, req.Description)
	if err != nil {
		return nil, err
	}
	return &v1.CreateRoleResponse{
		ID:   role.ID,
		Name: role.Name,
		Code: role.Code,
	}, nil
}

func (c *roleControllerImpl) UpdateRole(ctx context.Context, req *v1.UpdateRoleRequest) (*v1.UpdateRoleResponse, error) {
	err := service.GetRoleService().UpdateRole(req.ID, req.Name, req.Code, req.Description, req.Status)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRoleResponse{Success: true}, nil
}

func (c *roleControllerImpl) DeleteRole(ctx context.Context, req *v1.DeleteRoleRequest) (*v1.DeleteRoleResponse, error) {
	err := service.GetRoleService().DeleteRole(req.ID)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteRoleResponse{Success: true}, nil
}

func (c *roleControllerImpl) GetRole(ctx context.Context, req *v1.GetRoleRequest) (*v1.GetRoleResponse, error) {
	role, err := service.GetRoleService().GetRole(req.ID)
	if err != nil {
		return nil, err
	}
	return &v1.GetRoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Status:      role.Status,
	}, nil
}

func (c *roleControllerImpl) ListRoles(ctx context.Context, req *v1.ListRolesRequest) (*v1.ListRolesResponse, error) {
	roles, total, err := service.GetRoleService().ListRoles(req.Page, req.PageSize, req.Keyword)
	if err != nil {
		return nil, err
	}

	list := make([]v1.GetRoleResponse, 0, len(roles))
	for _, r := range roles {
		list = append(list, v1.GetRoleResponse{
			ID:          r.ID,
			Name:        r.Name,
			Code:        r.Code,
			Description: r.Description,
			Status:      r.Status,
		})
	}

	return &v1.ListRolesResponse{
		Total: total,
		List:  list,
	}, nil
}
