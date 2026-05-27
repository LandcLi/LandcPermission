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
