package role

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
	dao.RegisterRoleDAO(&roleDAOImpl{
		redisClient: db.GetRedis(),
	})
}

type roleDAOImpl struct {
	redisClient *redis.Client
}

func (d *roleDAOImpl) Create(role *model.Role) error {
	return db.GetDB().Create(role).Error
}

func (d *roleDAOImpl) Update(role *model.Role) error {
	err := db.GetDB().Save(role).Error
	if err == nil {
		d.invalidateCache(role.ID, role.Code)
	}
	return err
}

func (d *roleDAOImpl) Delete(id uint) error {
	if d.redisClient != nil {
		if role, err := d.GetByID(id); err == nil {
			defer d.invalidateCache(role.ID, role.Code)
		}
	}
	return db.GetDB().Delete(&model.Role{}, id).Error
}

func (d *roleDAOImpl) GetByID(id uint) (*model.Role, error) {
	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("role:id:%d", id)
		data, err := d.redisClient.Get(context.Background(), cacheKey).Bytes()
		if err == nil {
			var role model.Role
			if err := json.Unmarshal(data, &role); err == nil {
				return &role, nil
			}
		}
	}

	var role model.Role
	err := db.GetDB().First(&role, id).Error
	if err != nil {
		return nil, err
	}

	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("role:id:%d", id)
		data, _ := json.Marshal(role)
		d.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
	}

	return &role, nil
}

func (d *roleDAOImpl) GetByCode(code string) (*model.Role, error) {
	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("role:code:%s", code)
		data, err := d.redisClient.Get(context.Background(), cacheKey).Bytes()
		if err == nil {
			var role model.Role
			if err := json.Unmarshal(data, &role); err == nil {
				return &role, nil
			}
		}
	}

	var role model.Role
	err := db.GetDB().Where("code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}

	if d.redisClient != nil {
		cacheKey := fmt.Sprintf("role:code:%s", code)
		data, _ := json.Marshal(role)
		d.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
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

func (d *roleDAOImpl) invalidateCache(id uint, code string) {
	if d.redisClient == nil {
		return
	}
	ctx := context.Background()
	d.redisClient.Del(ctx, fmt.Sprintf("role:id:%d", id))
	if code != "" {
		d.redisClient.Del(ctx, fmt.Sprintf("role:code:%s", code))
	}
}
