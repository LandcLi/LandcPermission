package abac

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/service"
)

func init() {
	// ABAC 校验器默认不自动注册，需要显式调用 RegisterABACChecker
}

// ABACChecker ABAC 权限校验器实现
type ABACChecker struct {
	policies []service.ABACPolicy
}

// NewABACChecker 创建 ABAC 校验器
func NewABACChecker(policies []service.ABACPolicy) *ABACChecker {
	return &ABACChecker{
		policies: policies,
	}
}

// CheckPermission 校验权限
func (c *ABACChecker) CheckPermission(ctx *model.PermissionContext) (bool, error) {
	policies, err := c.GetPolicies(ctx)
	if err != nil {
		return false, err
	}

	// 按优先级排序（已排序）
	for _, policy := range policies {
		allowed, err := c.EvaluatePolicy(ctx, &policy)
		if err != nil {
			return false, err
		}

		// 如果策略匹配且效果为 deny，直接拒绝
		if policy.Effect == service.PolicyEffectDeny && allowed {
			return false, nil
		}

		// 如果策略匹配且效果为 allow，允许访问
		if policy.Effect == service.PolicyEffectAllow && allowed {
			return true, nil
		}
	}

	// 默认拒绝
	return false, nil
}

// GetModelType 获取权限模型类型
func (c *ABACChecker) GetModelType() service.PermissionModelType {
	return service.PermissionModelABAC
}

// EvaluatePolicy 评估策略
func (c *ABACChecker) EvaluatePolicy(ctx *model.PermissionContext, policy *service.ABACPolicy) (bool, error) {
	for _, condition := range policy.Conditions {
		matched, err := c.evaluateCondition(ctx, condition)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}

// GetPolicies 获取适用策略
func (c *ABACChecker) GetPolicies(ctx *model.PermissionContext) ([]service.ABACPolicy, error) {
	// 返回所有策略（实际应用中可以根据 context 过滤）
	return c.policies, nil
}

// evaluateCondition 评估单个条件
func (c *ABACChecker) evaluateCondition(ctx *model.PermissionContext, condition service.PolicyCondition) (bool, error) {
	var attrValue interface{}
	var exists bool

	switch condition.AttributeType {
	case service.AttributeTypeUser:
		attrValue, exists = ctx.GetUserAttribute(condition.AttributeKey)
	case service.AttributeTypeResource:
		attrValue, exists = ctx.GetResourceAttribute(condition.AttributeKey)
	case service.AttributeTypeEnvironment:
		attrValue, exists = ctx.GetEnvironmentAttr(condition.AttributeKey)
	default:
		return false, fmt.Errorf("unknown attribute type: %s", condition.AttributeType)
	}

	if !exists {
		return false, nil
	}

	return c.compareValues(attrValue, condition.Operator, condition.Value)
}

// compareValues 比较值
func (c *ABACChecker) compareValues(attrValue interface{}, op service.ConditionOperator, expectedValue interface{}) (bool, error) {
	switch op {
	case service.ConditionOperatorEqual:
		return reflect.DeepEqual(attrValue, expectedValue), nil

	case service.ConditionOperatorNotEqual:
		return !reflect.DeepEqual(attrValue, expectedValue), nil

	case service.ConditionOperatorIn:
		return c.isInSlice(attrValue, expectedValue), nil

	case service.ConditionOperatorNotIn:
		return !c.isInSlice(attrValue, expectedValue), nil

	case service.ConditionOperatorContains:
		return c.contains(attrValue, expectedValue), nil

	case service.ConditionOperatorGT, service.ConditionOperatorGTE, service.ConditionOperatorLT, service.ConditionOperatorLTE:
		return c.compareNumbers(attrValue, op, expectedValue)

	case service.ConditionOperatorMatch:
		return c.matchRegex(attrValue, expectedValue)

	default:
		return false, fmt.Errorf("unknown operator: %s", op)
	}
}

// isInSlice 检查值是否在切片中
func (c *ABACChecker) isInSlice(value interface{}, slice interface{}) bool {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return false
	}

	for i := 0; i < rv.Len(); i++ {
		if reflect.DeepEqual(value, rv.Index(i).Interface()) {
			return true
		}
	}
	return false
}

// contains 检查字符串包含
func (c *ABACChecker) contains(value interface{}, substr interface{}) bool {
	str, ok1 := value.(string)
	sub, ok2 := substr.(string)
	if !ok1 || !ok2 {
		return false
	}
	return strings.Contains(str, sub)
}

// compareNumbers 比较数字
func (c *ABACChecker) compareNumbers(value interface{}, op service.ConditionOperator, expected interface{}) (bool, error) {
	v, err := c.toFloat64(value)
	if err != nil {
		return false, err
	}
	e, err := c.toFloat64(expected)
	if err != nil {
		return false, err
	}

	switch op {
	case service.ConditionOperatorGT:
		return v > e, nil
	case service.ConditionOperatorGTE:
		return v >= e, nil
	case service.ConditionOperatorLT:
		return v < e, nil
	case service.ConditionOperatorLTE:
		return v <= e, nil
	default:
		return false, fmt.Errorf("invalid comparison operator: %s", op)
	}
}

// toFloat64 转换为 float64
func (c *ABACChecker) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// matchRegex 正则匹配
func (c *ABACChecker) matchRegex(value interface{}, pattern interface{}) (bool, error) {
	str, ok1 := value.(string)
	pat, ok2 := pattern.(string)
	if !ok1 || !ok2 {
		return false, fmt.Errorf("invalid types for regex match")
	}

	re, err := regexp.Compile(pat)
	if err != nil {
		return false, err
	}
	return re.MatchString(str), nil
}

// RegisterABACChecker 注册 ABAC 校验器
func RegisterABACChecker(policies []service.ABACPolicy) {
	checker := NewABACChecker(policies)
	service.RegisterPermissionChecker(checker)
	service.RegisterABACPermissionChecker(checker)
}
