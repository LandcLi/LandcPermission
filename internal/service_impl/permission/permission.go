package permission

import (
	"fmt"

	"github.com/LandcLi/LandcPermission/dao"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/service"
)

func init() {
	service.RegisterPermissionService(&permissionServiceImpl{})
}

type permissionServiceImpl struct{}

func (s *permissionServiceImpl) CreatePermission(name, code, resource, action, description string) (*model.Permission, error) {
	existing, _ := dao.GetPermissionDAO().GetByCode(code)
	if existing != nil {
		return nil, fmt.Errorf("permission code '%s' already exists", code)
	}

	perm := &model.Permission{
		Name:        name,
		Code:        code,
		Resource:    resource,
		Action:      action,
		Description: description,
		Status:      1,
	}
	if err := dao.GetPermissionDAO().Create(perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *permissionServiceImpl) UpdatePermission(id uint, name, code, resource, action, description string, status *int) error {
	perm, err := dao.GetPermissionDAO().GetByID(id)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	if name != "" {
		perm.Name = name
	}
	if code != "" {
		existing, _ := dao.GetPermissionDAO().GetByCode(code)
		if existing != nil && existing.ID != id {
			return fmt.Errorf("permission code '%s' already exists", code)
		}
		perm.Code = code
	}
	if resource != "" {
		perm.Resource = resource
	}
	if action != "" {
		perm.Action = action
	}
	if description != "" {
		perm.Description = description
	}
	if status != nil {
		perm.Status = *status
	}

	return dao.GetPermissionDAO().Update(perm)
}

func (s *permissionServiceImpl) DeletePermission(id uint) error {
	return dao.GetPermissionDAO().Delete(id)
}

func (s *permissionServiceImpl) GetPermission(id uint) (*model.Permission, error) {
	return dao.GetPermissionDAO().GetByID(id)
}

func (s *permissionServiceImpl) ListPermissions(page, pageSize int, resource, keyword string) ([]model.Permission, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return dao.GetPermissionDAO().List(page, pageSize, resource, keyword)
}

func (s *permissionServiceImpl) AssignPermissionsToRole(roleID uint, permissionIDs []uint) error {
	// Verify role exists
	if _, err := dao.GetRoleDAO().GetByID(roleID); err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	return dao.GetRolePermissionDAO().AssignPermissions(roleID, permissionIDs)
}

func (s *permissionServiceImpl) RevokePermissionsFromRole(roleID uint, permissionIDs []uint) error {
	return dao.GetRolePermissionDAO().RevokePermissions(roleID, permissionIDs)
}

func (s *permissionServiceImpl) GetRolePermissions(roleID uint) ([]model.Permission, error) {
	permIDs, err := dao.GetRolePermissionDAO().GetPermissionIDsByRoleID(roleID)
	if err != nil {
		return nil, err
	}
	if len(permIDs) == 0 {
		return []model.Permission{}, nil
	}
	return dao.GetPermissionDAO().GetByIDs(permIDs)
}

func (s *permissionServiceImpl) AssignRolesToUser(userID uint, roleIDs []uint) error {
	return dao.GetUserRoleDAO().AssignRoles(userID, roleIDs)
}

func (s *permissionServiceImpl) RevokeRolesFromUser(userID uint, roleIDs []uint) error {
	return dao.GetUserRoleDAO().RevokeRoles(userID, roleIDs)
}

func (s *permissionServiceImpl) GetUserRoles(userID uint) ([]model.Role, error) {
	allRoleIDs, err := s.getAllUserRoleIDs(userID)
	if err != nil {
		return nil, err
	}
	if len(allRoleIDs) == 0 {
		return []model.Role{}, nil
	}
	return dao.GetRoleDAO().GetByIDs(allRoleIDs)
}

func (s *permissionServiceImpl) GetUserPermissions(userID uint) ([]model.Permission, error) {
	allRoleIDs, err := s.getAllUserRoleIDs(userID)
	if err != nil {
		return nil, err
	}
	if len(allRoleIDs) == 0 {
		return []model.Permission{}, nil
	}

	permIDs, err := dao.GetRolePermissionDAO().GetPermissionIDsByRoleIDs(allRoleIDs)
	if err != nil {
		return nil, err
	}
	if len(permIDs) == 0 {
		return []model.Permission{}, nil
	}

	return dao.GetPermissionDAO().GetByIDs(permIDs)
}

func (s *permissionServiceImpl) CheckPermission(userID uint, resource, action string) (bool, error) {
	perms, err := s.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, p := range perms {
		if p.Status != 1 {
			continue
		}
		if p.Resource == resource && (p.Action == action || p.Action == "admin") {
			return true, nil
		}
		// Wildcard resource match: "user.*" matches "user.profile"
		if p.Resource == "*" || (len(p.Resource) > 1 && p.Resource[len(p.Resource)-1] == '*' &&
			len(resource) >= len(p.Resource)-1 &&
			resource[:len(p.Resource)-1] == p.Resource[:len(p.Resource)-1]) {
			if p.Action == action || p.Action == "admin" || p.Action == "*" {
				return true, nil
			}
		}
	}

	return false, nil
}

// CheckPermissionWithContext 支持多种权限模型的权限校验
// 默认实现为 RBAC，可通过 DI 替换为 ABAC/ACL 等其他模型
func (s *permissionServiceImpl) CheckPermissionWithContext(ctx *model.PermissionContext) (bool, error) {
	// 尝试使用注册的 PermissionChecker（支持多种权限模型）
	checker := service.GetPermissionChecker()
	if checker != nil {
		return checker.CheckPermission(ctx)
	}

	// 默认回退到 RBAC 实现
	return s.CheckPermission(ctx.UserID, ctx.Resource, ctx.Action)
}

// getAllUserRoleIDs aggregates directly assigned roles + organization-inherited roles.
func (s *permissionServiceImpl) getAllUserRoleIDs(userID uint) ([]uint, error) {
	// Get directly assigned roles
	directRoleIDs, err := dao.GetUserRoleDAO().GetRoleIDsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Get organization-inherited roles (if OrgRoleProvider is registered)
	provider := service.GetOrgRoleProvider()
	if provider == nil {
		return directRoleIDs, nil
	}

	inheritedRoleIDs, err := provider.GetUserInheritedRoleIDs(userID)
	if err != nil {
		// If org service is unavailable, fall back to direct roles only
		return directRoleIDs, nil
	}

	// Merge and deduplicate
	roleSet := make(map[uint]struct{})
	for _, id := range directRoleIDs {
		roleSet[id] = struct{}{}
	}
	for _, id := range inheritedRoleIDs {
		roleSet[id] = struct{}{}
	}

	result := make([]uint, 0, len(roleSet))
	for id := range roleSet {
		result = append(result, id)
	}
	return result, nil
}
