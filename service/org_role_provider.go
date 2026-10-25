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
func OverrideOrgRoleProvider(impl OrgRoleProvider) {
	di.Override[OrgRoleProvider]("org_role.provider", impl)
}

func GetOrgRoleProvider() OrgRoleProvider {
	v, err := di.Resolve[OrgRoleProvider]("org_role.provider")
	if err != nil {
		return nil
	}
	return v
}
