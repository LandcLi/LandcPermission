package service

import (
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/landc-go/mvc/pkg/di"
)

// RoleService defines the role business logic interface.
type RoleService interface {
	CreateRole(name, code, description string) (*model.Role, error)
	UpdateRole(id uint, name, code, description string, status *int) error
	DeleteRole(id uint) error
	GetRole(id uint) (*model.Role, error)
	ListRoles(page, pageSize int, keyword string) ([]model.Role, int64, error)
}

func RegisterRoleService(impl RoleService) {
	di.Provide[RoleService]("role.service", impl)
}

func GetRoleService() RoleService {
	return di.Require[RoleService]("role.service")
}
