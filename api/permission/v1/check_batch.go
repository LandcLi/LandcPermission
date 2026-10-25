package v1

import "github.com/LandcLi/landc-go/frame/pkg/meta"

// CheckPermissionItem 是批量判定中的一项。
type CheckPermissionItem struct {
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

// CheckPermissionsRequest 批量权限判定请求。
type CheckPermissionsRequest struct {
	meta.Meta `path:"/api/permission/check-batch" method:"POST"`
	UserID    uint                  `json:"user_id" binding:"required"`
	Checks    []CheckPermissionItem `json:"checks" binding:"required,min=1,max=200"`
}

// CheckPermissionResult 单条判定结果。
type CheckPermissionResult struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Allowed  bool   `json:"allowed"`
}

// CheckPermissionsResponse 批量判定响应，results 与请求 checks 顺序一一对应。
type CheckPermissionsResponse struct {
	Results []CheckPermissionResult `json:"results"`
}
