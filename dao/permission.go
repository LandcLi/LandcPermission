package dao

import (
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/landc-go/mvc/pkg/di"
)

// PermissionDAO defines the permission data access interface.
type PermissionDAO interface {
	Create(perm *model.Permission) error
	Update(perm *model.Permission) error
	Delete(id uint) error
	GetByID(id uint) (*model.Permission, error)
	GetByCode(code string) (*model.Permission, error)
	List(page, pageSize int, resource, keyword string) ([]model.Permission, int64, error)
	GetByIDs(ids []uint) ([]model.Permission, error)
}

func RegisterPermissionDAO(impl PermissionDAO) {
	di.Provide[PermissionDAO]("permission.dao", impl)
}

func GetPermissionDAO() PermissionDAO {
	return di.Require[PermissionDAO]("permission.dao")
}
