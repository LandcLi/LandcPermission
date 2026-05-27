package permission

import (
	"github.com/LandcLi/LandcPermission/dao"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/utility/db"
)

func init() {
	dao.RegisterPermissionDAO(&permissionDAOImpl{})
}

type permissionDAOImpl struct{}

func (d *permissionDAOImpl) Create(perm *model.Permission) error {
	return db.GetDB().Create(perm).Error
}

func (d *permissionDAOImpl) Update(perm *model.Permission) error {
	return db.GetDB().Save(perm).Error
}

func (d *permissionDAOImpl) Delete(id uint) error {
	return db.GetDB().Delete(&model.Permission{}, id).Error
}

func (d *permissionDAOImpl) GetByID(id uint) (*model.Permission, error) {
	var perm model.Permission
	err := db.GetDB().First(&perm, id).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (d *permissionDAOImpl) GetByCode(code string) (*model.Permission, error) {
	var perm model.Permission
	err := db.GetDB().Where("code = ?", code).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (d *permissionDAOImpl) List(page, pageSize int, resource, keyword string) ([]model.Permission, int64, error) {
	var perms []model.Permission
	var total int64

	query := db.GetDB().Model(&model.Permission{})
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}

	err := query.Find(&perms).Error
	return perms, total, err
}

func (d *permissionDAOImpl) GetByIDs(ids []uint) ([]model.Permission, error) {
	var perms []model.Permission
	err := db.GetDB().Where("id IN ?", ids).Find(&perms).Error
	return perms, err
}
