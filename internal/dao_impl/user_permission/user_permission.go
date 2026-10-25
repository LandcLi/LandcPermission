package user_permission

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LandcLi/LandcPermission/dao"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/utility/db"
	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
	"gorm.io/gorm"
)

const (
	cacheKeyPrefix = "user_direct_perms:"
	cacheTTL       = 30 * time.Minute
)

func init() {
	// 注意：这里刻意不在 init 阶段读取缓存实例。
	// DAO 包通常在包 init() 中被构造，而 Redis / DB 要等到 main 之后才就绪，
	// 在 init 阶段抓取会把 nil 永久固化到字段里（缓存静默失效）。
	// 因此缓存实例一律在使用时惰性获取。
	dao.RegisterUserPermissionDAO(&userPermissionDAOImpl{})
}

type userPermissionDAOImpl struct{}

// cache 惰性获取缓存实例；未配置 Redis 时返回 nil（调用方需判空）。
func (d *userPermissionDAOImpl) cache() framecache.Cache {
	c := db.GetRedis()
	if c == nil {
		return nil
	}
	cache, ok := c.(framecache.Cache)
	if !ok {
		return nil
	}
	return cache
}

func (d *userPermissionDAOImpl) invalidate(userID uint) {
	cache := d.cache()
	if cache == nil {
		return
	}
	_ = cache.Delete(context.Background(), fmt.Sprintf("%s%d", cacheKeyPrefix, userID))
}

// Grant 授予直接权限。已存在相同 (user_id, permission_id) 的记录时，
// 更新其过期时间与授权原因，而不是报唯一键冲突——重复授予应当是幂等的。
func (d *userPermissionDAOImpl) Grant(userID uint, grants []model.UserPermission) error {
	if len(grants) == 0 {
		return nil
	}

	tx := db.GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, g := range grants {
		record := &model.UserPermission{
			UserID:       userID,
			PermissionID: g.PermissionID,
			ExpireAt:     g.ExpireAt,
			Reason:       g.Reason,
		}

		var existing model.UserPermission
		err := tx.Where("user_id = ? AND permission_id = ?", userID, g.PermissionID).
			First(&existing).Error

		if err == nil {
			// 已存在：刷新过期时间与原因
			updates := map[string]interface{}{
				"expire_at": g.ExpireAt,
				"reason":    g.Reason,
			}
			if err := tx.Model(&existing).Updates(updates).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("update user permission grant: %w", err)
			}
			continue
		}

		// 只有「确实不存在」才新建；真实 DB 错误必须原样上报，不能吞掉
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return fmt.Errorf("query user permission grant: %w", err)
		}
		if err := tx.Create(record).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("create user permission grant: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	d.invalidate(userID)
	return nil
}

// Revoke 撤销指定权限的直接授权。
func (d *userPermissionDAOImpl) Revoke(userID uint, permissionIDs []uint) error {
	if len(permissionIDs) == 0 {
		return nil
	}
	err := db.GetDB().
		Where("user_id = ? AND permission_id IN ?", userID, permissionIDs).
		Delete(&model.UserPermission{}).Error
	if err != nil {
		return err
	}
	d.invalidate(userID)
	return nil
}

// GetByUserID 返回用户全部直接授权（含已过期，管理端可见完整授权历史）。
func (d *userPermissionDAOImpl) GetByUserID(userID uint) ([]model.UserPermission, error) {
	var records []model.UserPermission
	err := db.GetDB().
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetActivePermissionIDsByUserID 返回用户当前有效的直接权限 ID。
// 过期判断下推到 SQL，避免把整表捞到内存里再过滤。
func (d *userPermissionDAOImpl) GetActivePermissionIDsByUserID(userID uint) ([]uint, error) {
	cacheKey := fmt.Sprintf("%s%d", cacheKeyPrefix, userID)
	if cache := d.cache(); cache != nil {
		var ids []uint
		if err := cache.GetObject(context.Background(), cacheKey, &ids); err == nil {
			return ids, nil
		}
	}

	now := time.Now()
	var ids []uint
	// 括号必须显式写出：GORM 把多个 Where 用 AND 串起来时不会为条件内部的 OR
	// 自动加括号，`user_id = ? AND expire_at IS NULL OR expire_at > ?` 会被 SQL
	// 按优先级解析成 `(user_id = ? AND expire_at IS NULL) OR (expire_at > ?)`，
	// 从而把其他用户的未过期授权也捞进来。
	err := db.GetDB().Model(&model.UserPermission{}).
		Where("user_id = ?", userID).
		Where("(expire_at IS NULL OR expire_at > ?)", now).
		Pluck("permission_id", &ids).Error
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []uint{}
	}

	if cache := d.cache(); cache != nil {
		_ = cache.SetObject(context.Background(), cacheKey, ids, cacheTTL)
	}
	return ids, nil
}

// DeleteExpired 删除已过期的授权记录。仅用于回收空间，正确性不依赖它。
func (d *userPermissionDAOImpl) DeleteExpired() (int64, error) {
	now := time.Now()
	// 先取出 user_id 用于精准失效缓存（过期记录通常很少，代价可控）
	var affectedUsers []uint
	if err := db.GetDB().Model(&model.UserPermission{}).
		Where("expire_at IS NOT NULL AND expire_at <= ?", now).
		Distinct("user_id").
		Pluck("user_id", &affectedUsers).Error; err != nil {
		return 0, err
	}

	result := db.GetDB().
		Where("expire_at IS NOT NULL AND expire_at <= ?", now).
		Delete(&model.UserPermission{})
	if result.Error != nil {
		return 0, result.Error
	}

	for _, uid := range affectedUsers {
		d.invalidate(uid)
	}
	return result.RowsAffected, nil
}
