package user_role

import (
	"context"
	"fmt"
	"time"

	"github.com/LandcLi/LandcPermission/dao"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/utility/db"
	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
)

func init() {
	// 刻意不在 init 阶段读取缓存实例：DAO 在包 init() 中构造，而 Redis / DB
	// 要等 main 之后才就绪。在 init 阶段抓取会把 nil 永久固化到字段里，
	// 表现为「Redis 明明配了却从不命中缓存」。缓存实例一律使用时惰性获取。
	dao.RegisterRolePermissionDAO(&rolePermissionDAOImpl{})
	dao.RegisterUserRoleDAO(&userRoleDAOImpl{})
}

// cache 惰性获取缓存实例；未配置 Redis 时返回 nil（调用方需判空）。
func cache() framecache.Cache {
	c := db.GetRedis()
	if c == nil {
		return nil
	}
	fc, ok := c.(framecache.Cache)
	if !ok {
		return nil
	}
	return fc
}

// --- RolePermission DAO ---

type rolePermissionDAOImpl struct{}

func (d *rolePermissionDAOImpl) AssignPermissions(roleID uint, permissionIDs []uint) error {
	tx := db.GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for _, permID := range permissionIDs {
		rp := &model.RolePermission{RoleID: roleID, PermissionID: permID}
		if err := tx.FirstOrCreate(rp, "role_id = ? AND permission_id = ?", roleID, permID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	err := tx.Commit().Error
	if err == nil {
		d.invalidateRolePermCache(roleID)
	}
	return err
}

func (d *rolePermissionDAOImpl) RevokePermissions(roleID uint, permissionIDs []uint) error {
	err := db.GetDB().Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
		Delete(&model.RolePermission{}).Error
	if err == nil {
		d.invalidateRolePermCache(roleID)
	}
	return err
}

func (d *rolePermissionDAOImpl) GetPermissionIDsByRoleID(roleID uint) ([]uint, error) {
	cacheKey := fmt.Sprintf("role_perms:%d", roleID)
	if c := cache(); c != nil {
		var ids []uint
		if err := c.GetObject(context.Background(), cacheKey, &ids); err == nil {
			return ids, nil
		}
	}

	var ids []uint
	err := db.GetDB().Model(&model.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permission_id", &ids).Error
	if err != nil {
		return nil, err
	}

	if c := cache(); c != nil {
		_ = c.SetObject(context.Background(), cacheKey, ids, 30*time.Minute)
	}

	return ids, nil
}

func (d *rolePermissionDAOImpl) GetPermissionIDsByRoleIDs(roleIDs []uint) ([]uint, error) {
	var ids []uint
	err := db.GetDB().Model(&model.RolePermission{}).
		Where("role_id IN ?", roleIDs).
		Distinct("permission_id").
		Pluck("permission_id", &ids).Error
	return ids, err
}

func (d *rolePermissionDAOImpl) invalidateRolePermCache(roleID uint) {
	c := cache()
	if c == nil {
		return
	}
	_ = c.Delete(context.Background(), fmt.Sprintf("role_perms:%d", roleID))
}

// --- UserRole DAO ---

type userRoleDAOImpl struct{}

func (d *userRoleDAOImpl) AssignRoles(userID uint, roleIDs []uint) error {
	tx := db.GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for _, roleID := range roleIDs {
		ur := &model.UserRole{UserID: userID, RoleID: roleID}
		if err := tx.FirstOrCreate(ur, "user_id = ? AND role_id = ?", userID, roleID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	err := tx.Commit().Error
	if err == nil {
		d.invalidateUserRoleCache(userID)
	}
	return err
}

func (d *userRoleDAOImpl) RevokeRoles(userID uint, roleIDs []uint) error {
	err := db.GetDB().Where("user_id = ? AND role_id IN ?", userID, roleIDs).
		Delete(&model.UserRole{}).Error
	if err == nil {
		d.invalidateUserRoleCache(userID)
	}
	return err
}

func (d *userRoleDAOImpl) GetRoleIDsByUserID(userID uint) ([]uint, error) {
	cacheKey := fmt.Sprintf("user_roles:%d", userID)
	if c := cache(); c != nil {
		var ids []uint
		if err := c.GetObject(context.Background(), cacheKey, &ids); err == nil {
			return ids, nil
		}
	}

	var ids []uint
	err := db.GetDB().Model(&model.UserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &ids).Error
	if err != nil {
		return nil, err
	}

	if c := cache(); c != nil {
		_ = c.SetObject(context.Background(), cacheKey, ids, 30*time.Minute)
	}

	return ids, nil
}

func (d *userRoleDAOImpl) GetUsersByRoleID(roleID uint) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := db.GetDB().Where("role_id = ?", roleID).Find(&userRoles).Error
	return userRoles, err
}

func (d *userRoleDAOImpl) invalidateUserRoleCache(userID uint) {
	c := cache()
	if c == nil {
		return
	}
	_ = c.Delete(context.Background(), fmt.Sprintf("user_roles:%d", userID))
}
