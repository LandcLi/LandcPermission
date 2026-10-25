package model

import "context"

// PermissionContext 权限校验上下文，用于传递额外的属性信息。
// 支持多种权限模型（RBAC、ABAC、ACL等）扩展。
type PermissionContext struct {
	context.Context

	// 基础字段
	UserID   uint
	Resource string
	Action   string

	// ABAC 扩展字段（基于属性的访问控制）
	UserAttributes     map[string]interface{} // 用户属性（部门、职级、角色等）
	ResourceAttributes map[string]interface{} // 资源属性（所有者、敏感级别等）
	EnvironmentAttrs   map[string]interface{} // 环境属性（时间、IP、设备等）

	// 租户隔离（SAAS 多租户支持）
	TenantID uint
}

// NewPermissionContext 创建权限上下文
func NewPermissionContext(ctx context.Context, userID uint, resource, action string) *PermissionContext {
	return &PermissionContext{
		Context:  ctx,
		UserID:   userID,
		Resource: resource,
		Action:   action,
	}
}

// WithUserAttributes 设置用户属性
func (pc *PermissionContext) WithUserAttributes(attrs map[string]interface{}) *PermissionContext {
	pc.UserAttributes = attrs
	return pc
}

// WithResourceAttributes 设置资源属性
func (pc *PermissionContext) WithResourceAttributes(attrs map[string]interface{}) *PermissionContext {
	pc.ResourceAttributes = attrs
	return pc
}

// WithEnvironmentAttrs 设置环境属性
func (pc *PermissionContext) WithEnvironmentAttrs(attrs map[string]interface{}) *PermissionContext {
	pc.EnvironmentAttrs = attrs
	return pc
}

// WithTenantID 设置租户ID
func (pc *PermissionContext) WithTenantID(tenantID uint) *PermissionContext {
	pc.TenantID = tenantID
	return pc
}

// GetUserID 获取用户ID
func (pc *PermissionContext) GetUserID() uint {
	return pc.UserID
}

// GetResource 获取资源
func (pc *PermissionContext) GetResource() string {
	return pc.Resource
}

// GetAction 获取操作
func (pc *PermissionContext) GetAction() string {
	return pc.Action
}

// GetUserAttribute 获取用户属性
func (pc *PermissionContext) GetUserAttribute(key string) (interface{}, bool) {
	if pc.UserAttributes == nil {
		return nil, false
	}
	v, ok := pc.UserAttributes[key]
	return v, ok
}

// GetResourceAttribute 获取资源属性
func (pc *PermissionContext) GetResourceAttribute(key string) (interface{}, bool) {
	if pc.ResourceAttributes == nil {
		return nil, false
	}
	v, ok := pc.ResourceAttributes[key]
	return v, ok
}

// GetEnvironmentAttr 获取环境属性
func (pc *PermissionContext) GetEnvironmentAttr(key string) (interface{}, bool) {
	if pc.EnvironmentAttrs == nil {
		return nil, false
	}
	v, ok := pc.EnvironmentAttrs[key]
	return v, ok
}

// PermissionCheck 是一次「是否拥有某权限」的询问。
type PermissionCheck struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// PermissionCheckResult 是批量判定的单条结果。
// 回带 resource/action 是为了让调用方按位置对齐时也能自校验，
// 避免前端把结果数组错位当成别的权限。
type PermissionCheckResult struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Allowed  bool   `json:"allowed"`
}
