package permission

import (
	"fmt"
	"time"

	"github.com/LandcLi/LandcPermission/dao"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/service"
)

func init() {
	service.RegisterPermissionService(&permissionServiceImpl{})
	// 把缓存失效能力交给 service 包，供 OverrideOrgRoleProvider 与运维入口调用。
	service.RegisterCacheInvalidator(func() { effectiveCache().Clear() })
}

type permissionServiceImpl struct{}

func (s *permissionServiceImpl) CreatePermission(name, code, resource, action, description string) (*model.Permission, error) {
	existing, _ := dao.GetPermissionDAO().GetByCode(code)
	if existing != nil {
		return nil, fmt.Errorf("permission code '%s' already exists", code)
	}

	perm := &model.Permission{
		Name:        name,
		Code:        code,
		Resource:    resource,
		Action:      action,
		Description: description,
		Status:      1,
	}
	if err := dao.GetPermissionDAO().Create(perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *permissionServiceImpl) UpdatePermission(id uint, name, code, resource, action, description string, status *int) error {
	perm, err := dao.GetPermissionDAO().GetByID(id)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	if name != "" {
		perm.Name = name
	}
	if code != "" {
		existing, _ := dao.GetPermissionDAO().GetByCode(code)
		if existing != nil && existing.ID != id {
			return fmt.Errorf("permission code '%s' already exists", code)
		}
		perm.Code = code
	}
	if resource != "" {
		perm.Resource = resource
	}
	if action != "" {
		perm.Action = action
	}
	if description != "" {
		perm.Description = description
	}
	if status != nil {
		perm.Status = *status
	}

	if err := dao.GetPermissionDAO().Update(perm); err != nil {
		return err
	}
	// resource/action/status 都可能被改，判定结果随之变化，整体作废。
	effectiveCache().Clear()
	return nil
}

func (s *permissionServiceImpl) DeletePermission(id uint) error {
	if err := dao.GetPermissionDAO().Delete(id); err != nil {
		return err
	}
	effectiveCache().Clear()
	return nil
}

func (s *permissionServiceImpl) GetPermission(id uint) (*model.Permission, error) {
	return dao.GetPermissionDAO().GetByID(id)
}

func (s *permissionServiceImpl) ListPermissions(page, pageSize int, resource, keyword string) ([]model.Permission, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return dao.GetPermissionDAO().List(page, pageSize, resource, keyword)
}

func (s *permissionServiceImpl) AssignPermissionsToRole(roleID uint, permissionIDs []uint) error {
	// Verify role exists
	if _, err := dao.GetRoleDAO().GetByID(roleID); err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	if err := dao.GetRolePermissionDAO().AssignPermissions(roleID, permissionIDs); err != nil {
		return err
	}
	// 角色权限变了，所有持有该角色的用户缓存全部作废。
	// 无法廉价反查「谁持有这个角色」，只能整体清空；该操作本身不频繁。
	effectiveCache().Clear()
	return nil
}

func (s *permissionServiceImpl) RevokePermissionsFromRole(roleID uint, permissionIDs []uint) error {
	if err := dao.GetRolePermissionDAO().RevokePermissions(roleID, permissionIDs); err != nil {
		return err
	}
	effectiveCache().Clear()
	return nil
}

func (s *permissionServiceImpl) GetRolePermissions(roleID uint) ([]model.Permission, error) {
	permIDs, err := dao.GetRolePermissionDAO().GetPermissionIDsByRoleID(roleID)
	if err != nil {
		return nil, err
	}
	if len(permIDs) == 0 {
		return []model.Permission{}, nil
	}
	return dao.GetPermissionDAO().GetByIDs(permIDs)
}

func (s *permissionServiceImpl) AssignRolesToUser(userID uint, roleIDs []uint) error {
	if err := dao.GetUserRoleDAO().AssignRoles(userID, roleIDs); err != nil {
		return err
	}
	effectiveCache().Invalidate(userID)
	return nil
}

func (s *permissionServiceImpl) RevokeRolesFromUser(userID uint, roleIDs []uint) error {
	if err := dao.GetUserRoleDAO().RevokeRoles(userID, roleIDs); err != nil {
		return err
	}
	effectiveCache().Invalidate(userID)
	return nil
}

func (s *permissionServiceImpl) GetUserRoles(userID uint) ([]model.Role, error) {
	allRoleIDs, err := s.getAllUserRoleIDs(userID)
	if err != nil {
		return nil, err
	}
	if len(allRoleIDs) == 0 {
		return []model.Role{}, nil
	}
	return dao.GetRoleDAO().GetByIDs(allRoleIDs)
}

func (s *permissionServiceImpl) GetUserPermissions(userID uint) ([]model.Permission, error) {
	allRoleIDs, err := s.getAllUserRoleIDs(userID)
	if err != nil {
		return nil, err
	}
	if len(allRoleIDs) == 0 {
		return []model.Permission{}, nil
	}

	permIDs, err := dao.GetRolePermissionDAO().GetPermissionIDsByRoleIDs(allRoleIDs)
	if err != nil {
		return nil, err
	}
	if len(permIDs) == 0 {
		return []model.Permission{}, nil
	}

	return dao.GetPermissionDAO().GetByIDs(permIDs)
}

// CheckPermission 校验用户是否拥有指定权限。
// 生效来源为「角色权限 ∪ 未过期的直接权限」，两者任一命中即放行。
func (s *permissionServiceImpl) CheckPermission(userID uint, resource, action string) (bool, error) {
	perms, err := s.loadEffectivePermissions(userID)
	if err != nil {
		return false, err
	}
	return matchPermission(perms, resource, action), nil
}

// CheckPermissions 批量判定同一用户的多个权限。
//
// 为什么需要：前端渲染菜单/按钮时要一次判几十个权限，逐个调 CheckPermission
// 即使有缓存也要几十次 RPC 往返。批量接口复用同一份有效权限集合，
// 无论多少项都只取一次权限。
func (s *permissionServiceImpl) CheckPermissions(userID uint, checks []model.PermissionCheck) ([]model.PermissionCheckResult, error) {
	results := make([]model.PermissionCheckResult, 0, len(checks))
	if len(checks) == 0 {
		return results, nil
	}

	perms, err := s.loadEffectivePermissions(userID)
	if err != nil {
		return nil, err
	}

	for _, c := range checks {
		results = append(results, model.PermissionCheckResult{
			Resource: c.Resource,
			Action:   c.Action,
			Allowed:  matchPermission(perms, c.Resource, c.Action),
		})
	}
	return results, nil
}

// matchPermission 在权限集合中查找是否命中 (resource, action)。
//
// 命中规则：
//   - resource 与 action 全等；
//   - 权限的 action 为 admin 时通配该 resource 下的任意动作；
//   - 权限的 resource 支持尾部 * 前缀匹配（如 user.* 命中 user.profile），
//     resource 为 * 时通配全部。
func matchPermission(perms []model.Permission, resource, action string) bool {
	for _, p := range perms {
		if p.Status != 1 {
			continue
		}
		if p.Resource == resource && (p.Action == action || p.Action == "admin") {
			return true
		}
		// Wildcard resource match: "user.*" matches "user.profile"
		if p.Resource == "*" || (len(p.Resource) > 1 && p.Resource[len(p.Resource)-1] == '*' &&
			len(resource) >= len(p.Resource)-1 &&
			resource[:len(p.Resource)-1] == p.Resource[:len(p.Resource)-1]) {
			if p.Action == action || p.Action == "admin" || p.Action == "*" {
				return true
			}
		}
	}
	return false
}

// --- User-Permission 直接授权 ---

// GrantUserPermissions 授予用户直接权限。
// 校验两点：权限必须存在；过期时间必须晚于当前时刻（不允许授一个「一出生就过期」的权限）。
func (s *permissionServiceImpl) GrantUserPermissions(userID uint, grants []model.UserPermission) error {
	if len(grants) == 0 {
		return nil
	}

	now := time.Now()
	ids := make([]uint, 0, len(grants))
	seen := make(map[uint]struct{}, len(grants))

	for _, g := range grants {
		if g.PermissionID == 0 {
			return fmt.Errorf("permission_id is required")
		}
		if g.ExpireAt != nil && !g.ExpireAt.After(now) {
			return fmt.Errorf("permission %d: expire_at must be in the future", g.PermissionID)
		}
		if _, ok := seen[g.PermissionID]; ok {
			continue
		}
		seen[g.PermissionID] = struct{}{}
		ids = append(ids, g.PermissionID)
	}

	// 权限必须真实存在，否则会产生永远无法命中的授权记录
	found, err := dao.GetPermissionDAO().GetByIDs(ids)
	if err != nil {
		return fmt.Errorf("verify permissions: %w", err)
	}
	if len(found) != len(ids) {
		existing := make(map[uint]struct{}, len(found))
		for _, p := range found {
			existing[p.ID] = struct{}{}
		}
		missing := make([]uint, 0, len(ids))
		for _, id := range ids {
			if _, ok := existing[id]; !ok {
				missing = append(missing, id)
			}
		}
		return fmt.Errorf("permissions not found: %v", missing)
	}

	if err := dao.GetUserPermissionDAO().Grant(userID, grants); err != nil {
		return err
	}
	effectiveCache().Invalidate(userID)
	return nil
}

// RevokeUserPermissions 撤销用户的直接权限。
func (s *permissionServiceImpl) RevokeUserPermissions(userID uint, permissionIDs []uint) error {
	if len(permissionIDs) == 0 {
		return nil
	}
	if err := dao.GetUserPermissionDAO().Revoke(userID, permissionIDs); err != nil {
		return err
	}
	effectiveCache().Invalidate(userID)
	return nil
}

// GetUserDirectPermissions 返回用户的直接授权记录（含已过期，供管理端查看全貌）。
// 调用方可借助 model.UserPermission.IsExpired 标记每条的当前状态。
func (s *permissionServiceImpl) GetUserDirectPermissions(userID uint) ([]model.UserPermission, error) {
	return dao.GetUserPermissionDAO().GetByUserID(userID)
}

// GetUserDirectPermissionDetails 返回直接授权记录，并补齐权限编码 / 名称与过期状态。
// 权限信息用一次批量查询取回，避免 N+1。
func (s *permissionServiceImpl) GetUserDirectPermissionDetails(userID uint) ([]model.UserPermissionDetail, error) {
	records, err := dao.GetUserPermissionDAO().GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return []model.UserPermissionDetail{}, nil
	}

	ids := make([]uint, 0, len(records))
	for _, r := range records {
		ids = append(ids, r.PermissionID)
	}

	permByID := make(map[uint]model.Permission, len(ids))
	if perms, err := dao.GetPermissionDAO().GetByIDs(ids); err == nil {
		for _, p := range perms {
			permByID[p.ID] = p
		}
	}

	now := time.Now()
	details := make([]model.UserPermissionDetail, 0, len(records))
	for _, r := range records {
		d := model.UserPermissionDetail{
			UserPermission: r,
			Expired:        r.IsExpired(now),
		}
		if p, ok := permByID[r.PermissionID]; ok {
			d.PermissionCode = p.Code
			d.PermissionName = p.Name
		}
		details = append(details, d)
	}
	return details, nil
}

// GetUserEffectivePermissions 返回用户最终有效权限 = 角色权限 ∪ 未过期直接权限，按权限 ID 去重。
// GetUserEffectivePermissions 返回用户最终有效权限 = 角色权限 ∪ 未过期直接权限。
// 结果可能来自缓存，见 loadEffectivePermissions。
func (s *permissionServiceImpl) GetUserEffectivePermissions(userID uint) ([]model.Permission, error) {
	return s.loadEffectivePermissions(userID)
}

// loadEffectivePermissions 读缓存，未命中则回源计算并写入。
func (s *permissionServiceImpl) loadEffectivePermissions(userID uint) ([]model.Permission, error) {
	cache := effectiveCache()
	if perms, ok := cache.Get(userID); ok {
		return perms, nil
	}

	perms, err := s.computeEffectivePermissions(userID)
	if err != nil {
		return nil, err
	}
	cache.Set(userID, perms)
	return perms, nil
}

// computeEffectivePermissions 是权限聚合的真实计算，不做缓存。
func (s *permissionServiceImpl) computeEffectivePermissions(userID uint) ([]model.Permission, error) {
	rolePerms, err := s.GetUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	directPermIDs, err := dao.GetUserPermissionDAO().GetActivePermissionIDsByUserID(userID)
	if err != nil {
		return nil, err
	}
	if len(directPermIDs) == 0 {
		return rolePerms, nil
	}

	directPerms, err := dao.GetPermissionDAO().GetByIDs(directPermIDs)
	if err != nil {
		return nil, err
	}

	seen := make(map[uint]struct{}, len(rolePerms)+len(directPerms))
	result := make([]model.Permission, 0, len(rolePerms)+len(directPerms))
	for _, p := range rolePerms {
		if _, ok := seen[p.ID]; ok {
			continue
		}
		seen[p.ID] = struct{}{}
		result = append(result, p)
	}
	for _, p := range directPerms {
		if _, ok := seen[p.ID]; ok {
			continue
		}
		seen[p.ID] = struct{}{}
		result = append(result, p)
	}
	return result, nil
}

// CheckPermissionWithContext 支持多种权限模型的权限校验
// 默认实现为 RBAC，可通过 DI 替换为 ABAC/ACL 等其他模型
func (s *permissionServiceImpl) CheckPermissionWithContext(ctx *model.PermissionContext) (bool, error) {
	// 尝试使用注册的 PermissionChecker（支持多种权限模型）
	checker := service.GetPermissionChecker()
	if checker != nil {
		return checker.CheckPermission(ctx)
	}

	// 默认回退到 RBAC 实现
	return s.CheckPermission(ctx.UserID, ctx.Resource, ctx.Action)
}

// getAllUserRoleIDs aggregates directly assigned roles + organization-inherited roles.
func (s *permissionServiceImpl) getAllUserRoleIDs(userID uint) ([]uint, error) {
	// Get directly assigned roles
	directRoleIDs, err := dao.GetUserRoleDAO().GetRoleIDsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Get organization-inherited roles (if OrgRoleProvider is registered)
	provider := service.GetOrgRoleProvider()
	if provider == nil {
		return directRoleIDs, nil
	}

	inheritedRoleIDs, err := provider.GetUserInheritedRoleIDs(userID)
	if err != nil {
		// If org service is unavailable, fall back to direct roles only
		return directRoleIDs, nil
	}

	// Merge and deduplicate
	roleSet := make(map[uint]struct{})
	for _, id := range directRoleIDs {
		roleSet[id] = struct{}{}
	}
	for _, id := range inheritedRoleIDs {
		roleSet[id] = struct{}{}
	}

	result := make([]uint, 0, len(roleSet))
	for id := range roleSet {
		result = append(result, id)
	}
	return result, nil
}
