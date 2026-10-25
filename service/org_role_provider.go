package service

import "github.com/LandcLi/landc-go/frame/pkg/di"

// OrgRoleProvider is an optional interface that provides organization-inherited role IDs.
// When LandcOrg is integrated, it provides roles inherited from departments and positions.
// If not registered, permission checks only use directly assigned roles.
type OrgRoleProvider interface {
	// GetUserInheritedRoleIDs returns role IDs that a user inherits from org structure.
	GetUserInheritedRoleIDs(userID uint) ([]uint, error)
}

func RegisterOrgRoleProvider(impl OrgRoleProvider) {
	di.Provide[OrgRoleProvider]("org_role.provider", impl)
}

// OverrideOrgRoleProvider 强制替换已注册的实现。
// 用于装配期切换数据来源（本地实现 / 远程 LandcOrg），或重复初始化时避免残留旧实例。
//
// 切换后会清空有效权限缓存：数据源变了，按旧数据源算出来的判定结果必须作废。
func OverrideOrgRoleProvider(impl OrgRoleProvider) {
	di.Override[OrgRoleProvider]("org_role.provider", impl)
	InvalidatePermissionCache()
}

// cacheInvalidator 由具体实现包在 init 时注入，默认空操作。
//
// 这样安排是为了保持依赖倒置：service 包定义契约、不感知实现，
// 而「清缓存」需要访问实现包里的缓存实例。
var cacheInvalidator = func() {}

// RegisterCacheInvalidator 注册缓存失效函数，由内部实现包调用。
func RegisterCacheInvalidator(fn func()) {
	if fn != nil {
		cacheInvalidator = fn
	}
}

// InvalidatePermissionCache 清空「用户有效权限」缓存。
//
// 什么时候需要手动调用：
//   - 组织架构侧（LandcOrg）改了部门/岗位角色 —— 跨服务变更不会通知本服务，
//     正常只能等 TTL 过期；要求立刻生效时调这个；
//   - 排查「权限改了不生效」时先清缓存，确认不是缓存导致的。
func InvalidatePermissionCache() {
	cacheInvalidator()
}

func GetOrgRoleProvider() OrgRoleProvider {
	v, err := di.Resolve[OrgRoleProvider]("org_role.provider")
	if err != nil {
		return nil
	}
	return v
}
