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

func GetOrgRoleProvider() OrgRoleProvider {
	v, err := di.Resolve[OrgRoleProvider]("org_role.provider")
	if err != nil {
		return nil
	}
	return v
}
