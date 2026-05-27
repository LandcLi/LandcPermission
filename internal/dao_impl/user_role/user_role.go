package user_role

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/LandcLi/LandcPermission/dao"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/utility/db"
	"github.com/go-redis/redis/v8"
)

func init() {
	impl := &userRoleDAOImpl{redisClient: db.GetRedis()}
	dao.RegisterRolePermissionDAO(&rolePermissionDAOImpl{redisClient: db.GetRedis()})
	dao.RegisterUserRoleDAO(impl)
}

// --- RolePermission DAO ---

type rolePermissionDAOImpl struct {
	redisClient *redis.Client
}

func (d *rolePermissionDAOImpl) AssignPermissions(roleID uint, permissionIDs []uint) error {
	tx := db.GetDB().Begin()
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
	// Try cache
	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("role_perms:%d", roleID)
		data, err := d.redisClient.Get(context.Background(), cacheKey).Bytes()
		if err == nil {
			var ids []uint
			if err := json.Unmarshal(data, &ids); err == nil {
				return ids, nil
			}
		}
	}

	var ids []uint
	err := db.GetDB().Model(&model.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permission_id", &ids).Error
	if err != nil {
		return nil, err
	}

	// Store in cache
	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("role_perms:%d", roleID)
		data, _ := json.Marshal(ids)
		d.redisClient.Set(context.Background(), cacheKey, data, 30*time.Minute)
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
	if d.redisClient == nil {
		return
	}
	d.redisClient.Del(context.Background(), fmt.Sprintf("role_perms:%d", roleID))
}

// --- UserRole DAO ---

type userRoleDAOImpl struct {
	redisClient *redis.Client
}

func (d *userRoleDAOImpl) AssignRoles(userID uint, roleIDs []uint) error {
	tx := db.GetDB().Begin()
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
	// Try cache
	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("user_roles:%d", userID)
		data, err := d.redisClient.Get(context.Background(), cacheKey).Bytes()
		if err == nil {
			var ids []uint
			if err := json.Unmarshal(data, &ids); err == nil {
				return ids, nil
			}
		}
	}

	var ids []uint
	err := db.GetDB().Model(&model.UserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &ids).Error
	if err != nil {
		return nil, err
	}

	// Store in cache
	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("user_roles:%d", userID)
		data, _ := json.Marshal(ids)
		d.redisClient.Set(context.Background(), cacheKey, data, 30*time.Minute)
	}

	return ids, nil
}

func (d *userRoleDAOImpl) GetUsersByRoleID(roleID uint) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := db.GetDB().Where("role_id = ?", roleID).Find(&userRoles).Error
	return userRoles, err
}

func (d *userRoleDAOImpl) invalidateUserRoleCache(userID uint) {
	if d.redisClient == nil {
		return
	}
	d.redisClient.Del(context.Background(), fmt.Sprintf("user_roles:%d", userID))
}
