package user_role

import (
	"github.com/LandcLi/LandcPermission/dao"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/utility/db"
)

func init() {
	dao.RegisterRolePermissionDAO(&rolePermissionDAOImpl{})
	dao.RegisterUserRoleDAO(&userRoleDAOImpl{})
}

// --- RolePermission DAO ---

type rolePermissionDAOImpl struct{}

func (d *rolePermissionDAOImpl) AssignPermissions(roleID uint, permissionIDs []uint) error {
	tx := db.GetDB().Begin()
	for _, permID := range permissionIDs {
		rp := &model.RolePermission{RoleID: roleID, PermissionID: permID}
		if err := tx.FirstOrCreate(rp, "role_id = ? AND permission_id = ?", roleID, permID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (d *rolePermissionDAOImpl) RevokePermissions(roleID uint, permissionIDs []uint) error {
	return db.GetDB().Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
		Delete(&model.RolePermission{}).Error
}

func (d *rolePermissionDAOImpl) GetPermissionIDsByRoleID(roleID uint) ([]uint, error) {
	var ids []uint
	err := db.GetDB().Model(&model.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permission_id", &ids).Error
	return ids, err
}

func (d *rolePermissionDAOImpl) GetPermissionIDsByRoleIDs(roleIDs []uint) ([]uint, error) {
	var ids []uint
	err := db.GetDB().Model(&model.RolePermission{}).
		Where("role_id IN ?", roleIDs).
		Distinct("permission_id").
		Pluck("permission_id", &ids).Error
	return ids, err
}

// --- UserRole DAO ---

type userRoleDAOImpl struct{}

func (d *userRoleDAOImpl) AssignRoles(userID uint, roleIDs []uint) error {
	tx := db.GetDB().Begin()
	for _, roleID := range roleIDs {
		ur := &model.UserRole{UserID: userID, RoleID: roleID}
		if err := tx.FirstOrCreate(ur, "user_id = ? AND role_id = ?", userID, roleID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (d *userRoleDAOImpl) RevokeRoles(userID uint, roleIDs []uint) error {
	return db.GetDB().Where("user_id = ? AND role_id IN ?", userID, roleIDs).
		Delete(&model.UserRole{}).Error
}

func (d *userRoleDAOImpl) GetRoleIDsByUserID(userID uint) ([]uint, error) {
	var ids []uint
	err := db.GetDB().Model(&model.UserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &ids).Error
	return ids, err
}

func (d *userRoleDAOImpl) GetUsersByRoleID(roleID uint) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := db.GetDB().Where("role_id = ?", roleID).Find(&userRoles).Error
	return userRoles, err
}
