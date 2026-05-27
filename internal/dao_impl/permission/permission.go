package permission

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
	dao.RegisterPermissionDAO(&permissionDAOImpl{
		redisClient: db.GetRedis(),
	})
}

type permissionDAOImpl struct {
	redisClient *redis.Client
}

func (d *permissionDAOImpl) Create(perm *model.Permission) error {
	return db.GetDB().Create(perm).Error
}

func (d *permissionDAOImpl) Update(perm *model.Permission) error {
	err := db.GetDB().Save(perm).Error
	if err == nil {
		d.invalidateCache(perm.ID, perm.Code)
	}
	return err
}

func (d *permissionDAOImpl) Delete(id uint) error {
	// Get permission info before delete for cache invalidation
	if d.redisClient != nil {
		if perm, err := d.GetByID(id); err == nil {
			defer d.invalidateCache(perm.ID, perm.Code)
		}
	}
	return db.GetDB().Delete(&model.Permission{}, id).Error
}

func (d *permissionDAOImpl) GetByID(id uint) (*model.Permission, error) {
	// Try cache first
	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("perm:id:%d", id)
		data, err := d.redisClient.Get(context.Background(), cacheKey).Bytes()
		if err == nil {
			var perm model.Permission
			if err := json.Unmarshal(data, &perm); err == nil {
				return &perm, nil
			}
		}
	}

	var perm model.Permission
	err := db.GetDB().First(&perm, id).Error
	if err != nil {
		return nil, err
	}

	// Store in cache
	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("perm:id:%d", id)
		data, _ := json.Marshal(perm)
		d.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
	}

	return &perm, nil
}

func (d *permissionDAOImpl) GetByCode(code string) (*model.Permission, error) {
	// Try cache first
	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("perm:code:%s", code)
		data, err := d.redisClient.Get(context.Background(), cacheKey).Bytes()
		if err == nil {
			var perm model.Permission
			if err := json.Unmarshal(data, &perm); err == nil {
				return &perm, nil
			}
		}
	}

	var perm model.Permission
	err := db.GetDB().Where("code = ?", code).First(&perm).Error
	if err != nil {
		return nil, err
	}

	// Store in cache
	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("perm:code:%s", code)
		data, _ := json.Marshal(perm)
		d.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
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

// invalidateCache removes cached permission data.
func (d *permissionDAOImpl) invalidateCache(id uint, code string) {
	if d.redisClient == nil {
		return
	}
	ctx := context.Background()
	d.redisClient.Del(ctx, fmt.Sprintf("perm:id:%d", id))
	if code != "" {
		d.redisClient.Del(ctx, fmt.Sprintf("perm:code:%s", code))
	}
}
