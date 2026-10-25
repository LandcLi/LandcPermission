package service

import (
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// PermissionModelType 权限模型类型
type PermissionModelType string

const (
	PermissionModelRBAC PermissionModelType = "rbac" // 基于角色的访问控制
	PermissionModelABAC PermissionModelType = "abac" // 基于属性的访问控制
	PermissionModelACL  PermissionModelType = "acl"  // 访问控制列表
)

// PermissionChecker 权限校验器接口（支持多种权限模型）
type PermissionChecker interface {
	// CheckPermission 校验权限
	CheckPermission(ctx *model.PermissionContext) (bool, error)

	// GetModelType 获取权限模型类型
	GetModelType() PermissionModelType
}

// RBACPermissionChecker RBAC 权限校验器接口
type RBACPermissionChecker interface {
	PermissionChecker
	// GetUserRoles 获取用户角色
	GetUserRoles(userID uint) ([]model.Role, error)
	// GetUserPermissions 获取用户权限
	GetUserPermissions(userID uint) ([]model.Permission, error)
}

// ABACPermissionChecker ABAC 权限校验器接口
type ABACPermissionChecker interface {
	PermissionChecker
	// EvaluatePolicy 评估策略
	EvaluatePolicy(ctx *model.PermissionContext, policy *ABACPolicy) (bool, error)
	// GetPolicies 获取适用策略
	GetPolicies(ctx *model.PermissionContext) ([]ABACPolicy, error)
}

// ABACPolicy ABAC 策略定义
type ABACPolicy struct {
	ID          uint
	Name        string
	Description string
	Conditions  []PolicyCondition
	Effect      PolicyEffect // allow 或 deny
	Priority    int          // 策略优先级，数值越大优先级越高
}

// PolicyCondition 策略条件
type PolicyCondition struct {
	AttributeType AttributeType     // 属性类型：user/resource/environment
	AttributeKey  string            // 属性键
	Operator      ConditionOperator // 条件操作符
	Value         interface{}       // 比较值
}

// AttributeType 属性类型
type AttributeType string

const (
	AttributeTypeUser        AttributeType = "user"
	AttributeTypeResource    AttributeType = "resource"
	AttributeTypeEnvironment AttributeType = "environment"
)

// ConditionOperator 条件操作符
type ConditionOperator string

const (
	ConditionOperatorEqual    ConditionOperator = "eq"
	ConditionOperatorNotEqual ConditionOperator = "ne"
	ConditionOperatorIn       ConditionOperator = "in"
	ConditionOperatorNotIn    ConditionOperator = "not_in"
	ConditionOperatorContains ConditionOperator = "contains"
	ConditionOperatorGT       ConditionOperator = "gt" // greater than
	ConditionOperatorGTE      ConditionOperator = "gte"
	ConditionOperatorLT       ConditionOperator = "lt" // less than
	ConditionOperatorLTE      ConditionOperator = "lte"
	ConditionOperatorMatch    ConditionOperator = "match" // 正则匹配
)

// PolicyEffect 策略效果
type PolicyEffect string

const (
	PolicyEffectAllow PolicyEffect = "allow"
	PolicyEffectDeny  PolicyEffect = "deny"
)

// ACLPermissionChecker ACL 权限校验器接口
type ACLPermissionChecker interface {
	PermissionChecker
	// CheckACL 直接检查 ACL 规则
	CheckACL(userID uint, resource, action string) (bool, error)
	// GrantACL 授予 ACL 权限
	GrantACL(userID uint, resource, action string) error
	// RevokeACL 撤销 ACL 权限
	RevokeACL(userID uint, resource, action string) error
}

// ---- DI 注册 ----

func RegisterPermissionChecker(impl PermissionChecker) {
	di.Provide[PermissionChecker]("permission.checker", impl)
}

func GetPermissionChecker() PermissionChecker {
	return di.Require[PermissionChecker]("permission.checker")
}

func RegisterRBACPermissionChecker(impl RBACPermissionChecker) {
	di.Provide[RBACPermissionChecker]("permission.checker.rbac", impl)
}

func GetRBACPermissionChecker() RBACPermissionChecker {
	return di.Require[RBACPermissionChecker]("permission.checker.rbac")
}

func RegisterABACPermissionChecker(impl ABACPermissionChecker) {
	di.Provide[ABACPermissionChecker]("permission.checker.abac", impl)
}

func GetABACPermissionChecker() ABACPermissionChecker {
	return di.Require[ABACPermissionChecker]("permission.checker.abac")
}

func RegisterACLPermissionChecker(impl ACLPermissionChecker) {
	di.Provide[ACLPermissionChecker]("permission.checker.acl", impl)
}

func GetACLPermissionChecker() ACLPermissionChecker {
	return di.Require[ACLPermissionChecker]("permission.checker.acl")
}
