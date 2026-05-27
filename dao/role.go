package dao

import (
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/landc-go/mvc/pkg/di"
)

// RoleDAO defines the role data access interface.
type RoleDAO interface {
	Create(role *model.Role) error
	Update(role *model.Role) error
	Delete(id uint) error
	GetByID(id uint) (*model.Role, error)
	GetByCode(code string) (*model.Role, error)
	List(page, pageSize int, keyword string) ([]model.Role, int64, error)
	GetByIDs(ids []uint) ([]model.Role, error)
}

func RegisterRoleDAO(impl RoleDAO) {
	di.Provide[RoleDAO]("role.dao", impl)
}

func GetRoleDAO() RoleDAO {
	return di.Require[RoleDAO]("role.dao")
}
