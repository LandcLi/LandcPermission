package v1

import (
	"time"

	"github.com/LandcLi/landc-go/frame/pkg/meta"
)

// --- Permission API ---

// CreatePermissionRequest creates a new permission.
type CreatePermissionRequest struct {
	meta.Meta   `path:"/api/permission/create" method:"POST"`
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

// CreatePermissionResponse returns the created permission.
type CreatePermissionResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// UpdatePermissionRequest updates an existing permission.
type UpdatePermissionRequest struct {
	meta.Meta   `path:"/api/permission/update" method:"POST"`
	ID          uint   `json:"id" binding:"required"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Status      *int   `json:"status"`
}

// UpdatePermissionResponse returns update result.
type UpdatePermissionResponse struct {
	Success bool `json:"success"`
}

// DeletePermissionRequest deletes a permission by ID.
type DeletePermissionRequest struct {
	meta.Meta `path:"/api/permission/delete" method:"POST"`
	ID        uint `json:"id" binding:"required"`
}

// DeletePermissionResponse returns deletion result.
type DeletePermissionResponse struct {
	Success bool `json:"success"`
}

// GetPermissionRequest gets a permission by ID.
type GetPermissionRequest struct {
	meta.Meta `path:"/api/permission/get" method:"POST"`
	ID        uint `json:"id" binding:"required"`
}

// GetPermissionResponse returns permission details.
type GetPermissionResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

// ListPermissionsRequest lists permissions with pagination.
type ListPermissionsRequest struct {
	meta.Meta `path:"/api/permission/list" method:"POST"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	Resource  string `json:"resource"`
	Keyword   string `json:"keyword"`
}

// ListPermissionsResponse returns paginated permission list.
type ListPermissionsResponse struct {
	Total int64                   `json:"total"`
	List  []GetPermissionResponse `json:"list"`
}

// --- Role-Permission Binding ---

// AssignPermissionsRequest assigns permissions to a role.
type AssignPermissionsRequest struct {
	meta.Meta     `path:"/api/permission/assign-to-role" method:"POST"`
	RoleID        uint   `json:"role_id" binding:"required"`
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

// AssignPermissionsResponse returns assignment result.
type AssignPermissionsResponse struct {
	Success bool `json:"success"`
}

// RevokePermissionsRequest revokes permissions from a role.
type RevokePermissionsRequest struct {
	meta.Meta     `path:"/api/permission/revoke-from-role" method:"POST"`
	RoleID        uint   `json:"role_id" binding:"required"`
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

// RevokePermissionsResponse returns revocation result.
type RevokePermissionsResponse struct {
	Success bool `json:"success"`
}

// GetRolePermissionsRequest gets all permissions of a role.
type GetRolePermissionsRequest struct {
	meta.Meta `path:"/api/permission/get-role-permissions" method:"POST"`
	RoleID    uint `json:"role_id" binding:"required"`
}

// GetRolePermissionsResponse returns role's permissions.
type GetRolePermissionsResponse struct {
	Permissions []GetPermissionResponse `json:"permissions"`
}

// --- User-Role Binding ---

// AssignRolesRequest assigns roles to a user.
type AssignRolesRequest struct {
	meta.Meta `path:"/api/permission/assign-roles" method:"POST"`
	UserID    uint   `json:"user_id" binding:"required"`
	RoleIDs   []uint `json:"role_ids" binding:"required"`
}

// AssignRolesResponse returns assignment result.
type AssignRolesResponse struct {
	Success bool `json:"success"`
}

// RevokeRolesRequest revokes roles from a user.
type RevokeRolesRequest struct {
	meta.Meta `path:"/api/permission/revoke-roles" method:"POST"`
	UserID    uint   `json:"user_id" binding:"required"`
	RoleIDs   []uint `json:"role_ids" binding:"required"`
}

// RevokeRolesResponse returns revocation result.
type RevokeRolesResponse struct {
	Success bool `json:"success"`
}

// GetUserRolesRequest gets all roles of a user.
type GetUserRolesRequest struct {
	meta.Meta `path:"/api/permission/get-user-roles" method:"POST"`
	UserID    uint `json:"user_id" binding:"required"`
}

// RoleInfo represents role information in permission context.
type RoleInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

// GetUserRolesResponse returns user's roles.
type GetUserRolesResponse struct {
	Roles []RoleInfo `json:"roles"`
}

// GetUserPermissionsRequest gets all permissions of a user (through roles).
type GetUserPermissionsRequest struct {
	meta.Meta `path:"/api/permission/get-user-permissions" method:"POST"`
	UserID    uint `json:"user_id" binding:"required"`
}

// GetUserPermissionsResponse returns user's effective permissions.
type GetUserPermissionsResponse struct {
	Permissions []GetPermissionResponse `json:"permissions"`
}

// --- Permission Check ---

// CheckPermissionRequest checks if a user has a specific permission.
type CheckPermissionRequest struct {
	meta.Meta `path:"/api/permission/check" method:"POST"`
	UserID    uint   `json:"user_id" binding:"required"`
	Resource  string `json:"resource" binding:"required"`
	Action    string `json:"action" binding:"required"`
}

// CheckPermissionResponse returns permission check result.
type CheckPermissionResponse struct {
	Allowed bool `json:"allowed"`
}

// CheckPermissionWithContextRequest 支持多种权限模型的权限校验请求
type CheckPermissionWithContextRequest struct {
	meta.Meta `path:"/api/permission/check-with-context" method:"POST"`
	UserID    uint                   `json:"user_id" binding:"required"`
	Resource  string                 `json:"resource" binding:"required"`
	Action    string                 `json:"action" binding:"required"`
	TenantID  uint                   `json:"tenant_id"` // 租户ID（SAAS 多租户支持）
	Attrs     PermissionContextAttrs `json:"attrs"`     // 扩展属性
}

// PermissionContextAttrs 权限上下文属性
type PermissionContextAttrs struct {
	UserAttributes     map[string]interface{} `json:"user_attributes"`     // 用户属性
	ResourceAttributes map[string]interface{} `json:"resource_attributes"` // 资源属性
	EnvironmentAttrs   map[string]interface{} `json:"environment_attrs"`   // 环境属性
}

// CheckPermissionWithContextResponse 返回权限校验结果
type CheckPermissionWithContextResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"` // 拒绝原因
}

// --- User-Permission 直接授权（绕过角色）---

// DirectPermissionGrant 单条直接授权项。
type DirectPermissionGrant struct {
	PermissionID uint       `json:"permission_id" binding:"required"`
	ExpireAt     *time.Time `json:"expire_at"` // 不传或 null 表示永久有效
	Reason       string     `json:"reason"`    // 授权原因，供审计追溯
}

// GrantUserPermissionsRequest 授予用户直接权限。
type GrantUserPermissionsRequest struct {
	meta.Meta `path:"/api/permission/grant-direct" method:"POST"`
	UserID    uint                    `json:"user_id" binding:"required"`
	Grants    []DirectPermissionGrant `json:"grants" binding:"required"`
}

// GrantUserPermissionsResponse 返回授权结果。
type GrantUserPermissionsResponse struct {
	Success bool `json:"success"`
}

// RevokeUserPermissionsRequest 撤销用户直接权限。
type RevokeUserPermissionsRequest struct {
	meta.Meta     `path:"/api/permission/revoke-direct" method:"POST"`
	UserID        uint   `json:"user_id" binding:"required"`
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

// RevokeUserPermissionsResponse 返回撤销结果。
type RevokeUserPermissionsResponse struct {
	Success bool `json:"success"`
}

// GetUserDirectPermissionsRequest 查询用户直接授权记录。
type GetUserDirectPermissionsRequest struct {
	meta.Meta `path:"/api/permission/get-direct" method:"POST"`
	UserID    uint `json:"user_id" binding:"required"`
}

// DirectPermissionInfo 直接授权记录（含过期状态与权限基本信息）。
type DirectPermissionInfo struct {
	ID           uint       `json:"id"`
	PermissionID uint       `json:"permission_id"`
	Code         string     `json:"code,omitempty"`
	Name         string     `json:"name,omitempty"`
	ExpireAt     *time.Time `json:"expire_at"`
	Expired      bool       `json:"expired"`
	Reason       string     `json:"reason"`
	CreatedAt    time.Time  `json:"created_at"`
}

// GetUserDirectPermissionsResponse 返回用户直接授权列表。
type GetUserDirectPermissionsResponse struct {
	Grants []DirectPermissionInfo `json:"grants"`
}

// GetUserEffectivePermissionsRequest 查询用户最终有效权限。
type GetUserEffectivePermissionsRequest struct {
	meta.Meta `path:"/api/permission/get-effective" method:"POST"`
	UserID    uint `json:"user_id" binding:"required"`
}

// GetUserEffectivePermissionsResponse 返回角色权限与直接权限的并集。
type GetUserEffectivePermissionsResponse struct {
	Permissions []GetPermissionResponse `json:"permissions"`
}
