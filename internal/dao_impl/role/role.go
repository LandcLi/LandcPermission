package role

import (
	"github.com/LandcLi/LandcPermission/dao"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/utility/db"
)

func init() {
	dao.RegisterRoleDAO(&roleDAOImpl{})
}

type roleDAOImpl struct{}

func (d *roleDAOImpl) Create(role *model.Role) error {
	return db.GetDB().Create(role).Error
}

func (d *roleDAOImpl) Update(role *model.Role) error {
	return db.GetDB().Save(role).Error
}

func (d *roleDAOImpl) Delete(id uint) error {
	return db.GetDB().Delete(&model.Role{}, id).Error
}

func (d *roleDAOImpl) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	err := db.GetDB().First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (d *roleDAOImpl) GetByCode(code string) (*model.Role, error) {
	var role model.Role
	err := db.GetDB().Where("code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (d *roleDAOImpl) List(page, pageSize int, keyword string) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	query := db.GetDB().Model(&model.Role{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}

	err := query.Find(&roles).Error
	return roles, total, err
}

func (d *roleDAOImpl) GetByIDs(ids []uint) ([]model.Role, error) {
	var roles []model.Role
	err := db.GetDB().Where("id IN ?", ids).Find(&roles).Error
	return roles, err
}
