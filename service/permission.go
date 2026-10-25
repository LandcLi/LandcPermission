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

	// User-Permission 直接授权（绕过角色）
	// GrantUserPermissions 授予直接权限；相同权限重复授予为幂等操作（刷新过期时间与原因）。
	GrantUserPermissions(userID uint, grants []model.UserPermission) error
	// RevokeUserPermissions 撤销直接权限。
	RevokeUserPermissions(userID uint, permissionIDs []uint) error
	// GetUserDirectPermissions 返回用户的直接授权记录（含已过期，供管理端审计查看）。
	GetUserDirectPermissions(userID uint) ([]model.UserPermission, error)
	// GetUserDirectPermissionDetails 返回直接授权记录并补齐权限编码/名称与过期状态。
	GetUserDirectPermissionDetails(userID uint) ([]model.UserPermissionDetail, error)
	// GetUserEffectivePermissions 返回用户最终有效权限 = 角色权限 ∪ 未过期直接权限。
	GetUserEffectivePermissions(userID uint) ([]model.Permission, error)

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
