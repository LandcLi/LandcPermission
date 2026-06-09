package service

import (
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// PermissionService defines the permission business logic interface.
type PermissionService interface {
	// Permission CRUD
	CreatePermission(name, code, resource, action, description string) (*model.Permission, error)
	UpdatePermission(id uint, name, code, resource, action, description string, status *int) error
	DeletePermission(id uint) error
	GetPermission(id uint) (*model.Permission, error)
	ListPermissions(page, pageSize int, resource, keyword string) ([]model.Permission, int64, error)

	// Role-Permission binding
	AssignPermissionsToRole(roleID uint, permissionIDs []uint) error
	RevokePermissionsFromRole(roleID uint, permissionIDs []uint) error
	GetRolePermissions(roleID uint) ([]model.Permission, error)

	// User-Role binding
	AssignRolesToUser(userID uint, roleIDs []uint) error
	RevokeRolesFromUser(userID uint, roleIDs []uint) error
	GetUserRoles(userID uint) ([]model.Role, error)
	GetUserPermissions(userID uint) ([]model.Permission, error)

	// Permission check (RBAC 基础方法)
	CheckPermission(userID uint, resource, action string) (bool, error)

	// Permission check with context (支持多种权限模型扩展)
	CheckPermissionWithContext(ctx *model.PermissionContext) (bool, error)
}

func RegisterPermissionService(impl PermissionService) {
	di.Provide[PermissionService]("permission.service", impl)
}

func GetPermissionService() PermissionService {
	return di.Require[PermissionService]("permission.service")
}
