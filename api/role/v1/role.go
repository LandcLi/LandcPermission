package v1

// --- Role API ---

// CreateRoleRequest creates a new role.
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

// CreateRoleResponse returns the created role.
type CreateRoleResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// UpdateRoleRequest updates an existing role.
type UpdateRoleRequest struct {
	ID          uint   `json:"id" binding:"required"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      *int   `json:"status"`
}

// UpdateRoleResponse returns the updated role info.
type UpdateRoleResponse struct {
	Success bool `json:"success"`
}

// DeleteRoleRequest deletes a role by ID.
type DeleteRoleRequest struct {
	ID uint `json:"id" binding:"required"`
}

// DeleteRoleResponse returns deletion result.
type DeleteRoleResponse struct {
	Success bool `json:"success"`
}

// GetRoleRequest gets a role by ID.
type GetRoleRequest struct {
	ID uint `json:"id" binding:"required"`
}

// GetRoleResponse returns role details.
type GetRoleResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

// ListRolesRequest lists roles with pagination.
type ListRolesRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Keyword  string `json:"keyword"`
}

// ListRolesResponse returns paginated role list.
type ListRolesResponse struct {
	Total int64            `json:"total"`
	List  []GetRoleResponse `json:"list"`
}
