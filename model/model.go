package model

import "time"

// Role represents a role entity in the RBAC model.
type Role struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Code        string    `json:"code" gorm:"uniqueIndex;size:100;not null"`
	Description string    `json:"description" gorm:"size:500"`
	Status      int       `json:"status" gorm:"default:1;comment:1=active,0=disabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a permission entity.
type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:200;not null"`
	Code        string    `json:"code" gorm:"uniqueIndex;size:200;not null"`
	Resource    string    `json:"resource" gorm:"size:200;not null;comment:resource identifier"`
	Action      string    `json:"action" gorm:"size:50;not null;comment:read/write/delete/admin"`
	Description string    `json:"description" gorm:"size:500"`
	Status      int       `json:"status" gorm:"default:1;comment:1=active,0=disabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RolePermission represents the many-to-many relationship between roles and permissions.
type RolePermission struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	RoleID       uint      `json:"role_id" gorm:"index;not null"`
	PermissionID uint      `json:"permission_id" gorm:"index;not null"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserRole represents the many-to-many relationship between users and roles.
type UserRole struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	RoleID    uint      `json:"role_id" gorm:"index;not null"`
	CreatedAt time.Time `json:"created_at"`
}

// UserPermission 用户直接权限（绕过角色的定向授权）。
//
// 与 RolePermission 的区别：RolePermission 描述「角色拥有什么权限」，
// UserPermission 直接把权限授予某个用户，不经过角色。典型场景：
//   - 应急授权：线上故障时临时给运维开某个写权限，事后自动失效
//   - 外部协作：外包/顾问只做一件事，不值得为它建角色
//   - 例外处理：某人比其角色多一项权限，但不应影响同角色的其他人
//
// 生效规则：最终有效权限 = 角色权限 ∪ 未过期的直接权限。
// ExpireAt 为 nil 表示永久有效；非 nil 且已过期时在校验阶段被过滤，
// 不依赖后台清理任务（清理仅用于回收空间）。
type UserPermission struct {
	ID           uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       uint       `json:"user_id" gorm:"uniqueIndex:uk_user_permission;not null"`
	PermissionID uint       `json:"permission_id" gorm:"uniqueIndex:uk_user_permission;not null"`
	ExpireAt     *time.Time `json:"expire_at" gorm:"index"` // nil = 永久
	Reason       string     `json:"reason" gorm:"size:500"` // 授权原因，供审计追溯
	CreatedAt    time.Time  `json:"created_at"`
}

// IsExpired 判断该授权是否已过期（ExpireAt 为 nil 视为永久有效）。
func (up *UserPermission) IsExpired(now time.Time) bool {
	return up.ExpireAt != nil && now.After(*up.ExpireAt)
}

// UserPermissionDetail 直接授权记录 + 权限基本信息，供管理端一次取全。
// 由 service 层聚合，避免调用方为每条授权再单查一次权限。
type UserPermissionDetail struct {
	UserPermission
	PermissionCode string `json:"permission_code"`
	PermissionName string `json:"permission_name"`
	Expired        bool   `json:"expired"`
}
