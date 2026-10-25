package dao

import (
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// RolePermissionDAO defines the role-permission relationship data access interface.
type RolePermissionDAO interface {
	AssignPermissions(roleID uint, permissionIDs []uint) error
	RevokePermissions(roleID uint, permissionIDs []uint) error
	GetPermissionIDsByRoleID(roleID uint) ([]uint, error)
	GetPermissionIDsByRoleIDs(roleIDs []uint) ([]uint, error)
}

func RegisterRolePermissionDAO(impl RolePermissionDAO) {
	di.Provide[RolePermissionDAO]("role_permission.dao", impl)
}

func GetRolePermissionDAO() RolePermissionDAO {
	return di.Require[RolePermissionDAO]("role_permission.dao")
}

// UserRoleDAO defines the user-role relationship data access interface.
type UserRoleDAO interface {
	AssignRoles(userID uint, roleIDs []uint) error
	RevokeRoles(userID uint, roleIDs []uint) error
	GetRoleIDsByUserID(userID uint) ([]uint, error)
	GetUsersByRoleID(roleID uint) ([]model.UserRole, error)
}

func RegisterUserRoleDAO(impl UserRoleDAO) {
	di.Provide[UserRoleDAO]("user_role.dao", impl)
}

func GetUserRoleDAO() UserRoleDAO {
	return di.Require[UserRoleDAO]("user_role.dao")
}

// UserPermissionDAO defines the direct user-permission grant data access interface.
// 直接权限独立于角色体系，授予与撤销不影响用户已绑定的角色。
type UserPermissionDAO interface {
	// Grant 授予直接权限。同一 (user_id, permission_id) 已存在时更新过期时间与原因。
	Grant(userID uint, grants []model.UserPermission) error
	// Revoke 撤销指定权限的直接授权。
	Revoke(userID uint, permissionIDs []uint) error
	// GetByUserID 返回用户的全部直接授权记录（含已过期，供管理端查看）。
	GetByUserID(userID uint) ([]model.UserPermission, error)
	// GetActivePermissionIDsByUserID 返回用户当前有效的直接权限 ID（过滤已过期）。
	GetActivePermissionIDsByUserID(userID uint) ([]uint, error)
	// DeleteExpired 删除已过期的授权记录，返回删除行数。用于周期性回收，非必需。
	DeleteExpired() (int64, error)
}

func RegisterUserPermissionDAO(impl UserPermissionDAO) {
	di.Provide[UserPermissionDAO]("user_permission.dao", impl)
}

func GetUserPermissionDAO() UserPermissionDAO {
	return di.Require[UserPermissionDAO]("user_permission.dao")
}
