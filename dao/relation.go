package dao

import (
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/landc-go/mvc/pkg/di"
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
