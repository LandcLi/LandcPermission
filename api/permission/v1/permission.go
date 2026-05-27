package v1

// --- Permission API ---

// CreatePermissionRequest creates a new permission.
type CreatePermissionRequest struct {
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
	ID uint `json:"id" binding:"required"`
}

// DeletePermissionResponse returns deletion result.
type DeletePermissionResponse struct {
	Success bool `json:"success"`
}

// GetPermissionRequest gets a permission by ID.
type GetPermissionRequest struct {
	ID uint `json:"id" binding:"required"`
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
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Resource string `json:"resource"`
	Keyword  string `json:"keyword"`
}

// ListPermissionsResponse returns paginated permission list.
type ListPermissionsResponse struct {
	Total int64                    `json:"total"`
	List  []GetPermissionResponse `json:"list"`
}

// --- Role-Permission Binding ---

// AssignPermissionsRequest assigns permissions to a role.
type AssignPermissionsRequest struct {
	RoleID        uint   `json:"role_id" binding:"required"`
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

// AssignPermissionsResponse returns assignment result.
type AssignPermissionsResponse struct {
	Success bool `json:"success"`
}

// RevokePermissionsRequest revokes permissions from a role.
type RevokePermissionsRequest struct {
	RoleID        uint   `json:"role_id" binding:"required"`
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

// RevokePermissionsResponse returns revocation result.
type RevokePermissionsResponse struct {
	Success bool `json:"success"`
}

// GetRolePermissionsRequest gets all permissions of a role.
type GetRolePermissionsRequest struct {
	RoleID uint `json:"role_id" binding:"required"`
}

// GetRolePermissionsResponse returns role's permissions.
type GetRolePermissionsResponse struct {
	Permissions []GetPermissionResponse `json:"permissions"`
}

// --- User-Role Binding ---

// AssignRolesRequest assigns roles to a user.
type AssignRolesRequest struct {
	UserID  uint   `json:"user_id" binding:"required"`
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

// AssignRolesResponse returns assignment result.
type AssignRolesResponse struct {
	Success bool `json:"success"`
}

// RevokeRolesRequest revokes roles from a user.
type RevokeRolesRequest struct {
	UserID  uint   `json:"user_id" binding:"required"`
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

// RevokeRolesResponse returns revocation result.
type RevokeRolesResponse struct {
	Success bool `json:"success"`
}

// GetUserRolesRequest gets all roles of a user.
type GetUserRolesRequest struct {
	UserID uint `json:"user_id" binding:"required"`
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
	UserID uint `json:"user_id" binding:"required"`
}

// GetUserPermissionsResponse returns user's effective permissions.
type GetUserPermissionsResponse struct {
	Permissions []GetPermissionResponse `json:"permissions"`
}

// --- Permission Check ---

// CheckPermissionRequest checks if a user has a specific permission.
type CheckPermissionRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

// CheckPermissionResponse returns permission check result.
type CheckPermissionResponse struct {
	Allowed bool `json:"allowed"`
}
