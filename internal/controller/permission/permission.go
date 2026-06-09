package permission

import (
	"context"

	permApi "github.com/LandcLi/LandcPermission/api/permission"
	v1 "github.com/LandcLi/LandcPermission/api/permission/v1"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/service"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

func init() {
	di.Provide[permApi.PermissionController]("permission.controller", &permissionControllerImpl{})
}

type permissionControllerImpl struct{}

func (c *permissionControllerImpl) CreatePermission(ctx context.Context, req *v1.CreatePermissionRequest) (*v1.CreatePermissionResponse, error) {
	perm, err := service.GetPermissionService().CreatePermission(req.Name, req.Code, req.Resource, req.Action, req.Description)
	if err != nil {
		return nil, err
	}
	return &v1.CreatePermissionResponse{
		ID:   perm.ID,
		Name: perm.Name,
		Code: perm.Code,
	}, nil
}

func (c *permissionControllerImpl) UpdatePermission(ctx context.Context, req *v1.UpdatePermissionRequest) (*v1.UpdatePermissionResponse, error) {
	err := service.GetPermissionService().UpdatePermission(req.ID, req.Name, req.Code, req.Resource, req.Action, req.Description, req.Status)
	if err != nil {
		return nil, err
	}
	return &v1.UpdatePermissionResponse{Success: true}, nil
}

func (c *permissionControllerImpl) DeletePermission(ctx context.Context, req *v1.DeletePermissionRequest) (*v1.DeletePermissionResponse, error) {
	err := service.GetPermissionService().DeletePermission(req.ID)
	if err != nil {
		return nil, err
	}
	return &v1.DeletePermissionResponse{Success: true}, nil
}

func (c *permissionControllerImpl) GetPermission(ctx context.Context, req *v1.GetPermissionRequest) (*v1.GetPermissionResponse, error) {
	perm, err := service.GetPermissionService().GetPermission(req.ID)
	if err != nil {
		return nil, err
	}
	return &v1.GetPermissionResponse{
		ID:          perm.ID,
		Name:        perm.Name,
		Code:        perm.Code,
		Resource:    perm.Resource,
		Action:      perm.Action,
		Description: perm.Description,
		Status:      perm.Status,
	}, nil
}

func (c *permissionControllerImpl) ListPermissions(ctx context.Context, req *v1.ListPermissionsRequest) (*v1.ListPermissionsResponse, error) {
	perms, total, err := service.GetPermissionService().ListPermissions(req.Page, req.PageSize, req.Resource, req.Keyword)
	if err != nil {
		return nil, err
	}

	list := make([]v1.GetPermissionResponse, 0, len(perms))
	for _, p := range perms {
		list = append(list, v1.GetPermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			Code:        p.Code,
			Resource:    p.Resource,
			Action:      p.Action,
			Description: p.Description,
			Status:      p.Status,
		})
	}

	return &v1.ListPermissionsResponse{
		Total: total,
		List:  list,
	}, nil
}

func (c *permissionControllerImpl) AssignPermissions(ctx context.Context, req *v1.AssignPermissionsRequest) (*v1.AssignPermissionsResponse, error) {
	err := service.GetPermissionService().AssignPermissionsToRole(req.RoleID, req.PermissionIDs)
	if err != nil {
		return nil, err
	}
	return &v1.AssignPermissionsResponse{Success: true}, nil
}

func (c *permissionControllerImpl) RevokePermissions(ctx context.Context, req *v1.RevokePermissionsRequest) (*v1.RevokePermissionsResponse, error) {
	err := service.GetPermissionService().RevokePermissionsFromRole(req.RoleID, req.PermissionIDs)
	if err != nil {
		return nil, err
	}
	return &v1.RevokePermissionsResponse{Success: true}, nil
}

func (c *permissionControllerImpl) GetRolePermissions(ctx context.Context, req *v1.GetRolePermissionsRequest) (*v1.GetRolePermissionsResponse, error) {
	perms, err := service.GetPermissionService().GetRolePermissions(req.RoleID)
	if err != nil {
		return nil, err
	}

	list := make([]v1.GetPermissionResponse, 0, len(perms))
	for _, p := range perms {
		list = append(list, v1.GetPermissionResponse{
			ID:       p.ID,
			Name:     p.Name,
			Code:     p.Code,
			Resource: p.Resource,
			Action:   p.Action,
			Status:   p.Status,
		})
	}

	return &v1.GetRolePermissionsResponse{Permissions: list}, nil
}

func (c *permissionControllerImpl) AssignRoles(ctx context.Context, req *v1.AssignRolesRequest) (*v1.AssignRolesResponse, error) {
	err := service.GetPermissionService().AssignRolesToUser(req.UserID, req.RoleIDs)
	if err != nil {
		return nil, err
	}
	return &v1.AssignRolesResponse{Success: true}, nil
}

func (c *permissionControllerImpl) RevokeRoles(ctx context.Context, req *v1.RevokeRolesRequest) (*v1.RevokeRolesResponse, error) {
	err := service.GetPermissionService().RevokeRolesFromUser(req.UserID, req.RoleIDs)
	if err != nil {
		return nil, err
	}
	return &v1.RevokeRolesResponse{Success: true}, nil
}

func (c *permissionControllerImpl) GetUserRoles(ctx context.Context, req *v1.GetUserRolesRequest) (*v1.GetUserRolesResponse, error) {
	roles, err := service.GetPermissionService().GetUserRoles(req.UserID)
	if err != nil {
		return nil, err
	}

	list := make([]v1.RoleInfo, 0, len(roles))
	for _, r := range roles {
		list = append(list, v1.RoleInfo{
			ID:          r.ID,
			Name:        r.Name,
			Code:        r.Code,
			Description: r.Description,
			Status:      r.Status,
		})
	}

	return &v1.GetUserRolesResponse{Roles: list}, nil
}

func (c *permissionControllerImpl) GetUserPermissions(ctx context.Context, req *v1.GetUserPermissionsRequest) (*v1.GetUserPermissionsResponse, error) {
	perms, err := service.GetPermissionService().GetUserPermissions(req.UserID)
	if err != nil {
		return nil, err
	}

	list := make([]v1.GetPermissionResponse, 0, len(perms))
	for _, p := range perms {
		list = append(list, v1.GetPermissionResponse{
			ID:       p.ID,
			Name:     p.Name,
			Code:     p.Code,
			Resource: p.Resource,
			Action:   p.Action,
			Status:   p.Status,
		})
	}

	return &v1.GetUserPermissionsResponse{Permissions: list}, nil
}

func (c *permissionControllerImpl) CheckPermission(ctx context.Context, req *v1.CheckPermissionRequest) (*v1.CheckPermissionResponse, error) {
	allowed, err := service.GetPermissionService().CheckPermission(req.UserID, req.Resource, req.Action)
	if err != nil {
		return nil, err
	}
	return &v1.CheckPermissionResponse{Allowed: allowed}, nil
}

func (c *permissionControllerImpl) CheckPermissionWithContext(ctx context.Context, req *v1.CheckPermissionWithContextRequest) (*v1.CheckPermissionWithContextResponse, error) {
	// 构建权限上下文
	permCtx := model.NewPermissionContext(ctx, req.UserID, req.Resource, req.Action).
		WithTenantID(req.TenantID).
		WithUserAttributes(req.Attrs.UserAttributes).
		WithResourceAttributes(req.Attrs.ResourceAttributes).
		WithEnvironmentAttrs(req.Attrs.EnvironmentAttrs)

	allowed, err := service.GetPermissionService().CheckPermissionWithContext(permCtx)
	if err != nil {
		return &v1.CheckPermissionWithContextResponse{
			Allowed: false,
			Reason:  err.Error(),
		}, nil
	}

	reason := ""
	if !allowed {
		reason = "permission denied"
	}

	return &v1.CheckPermissionWithContextResponse{
		Allowed: allowed,
		Reason:  reason,
	}, nil
}
