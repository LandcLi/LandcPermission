// Package orgrole 把 LandcOrg 的组织继承角色能力接入 LandcPermission。
//
// 职责边界：权限判定属于 LandcPermission，部门 / 岗位归属关系属于 LandcOrg。
// 两者通过 service.OrgRoleProvider 这个依赖倒置接口解耦 —— 接口由 Permission
// 定义（消费者侧），实现由本包提供，Permission 不反向依赖 LandcOrg 的代码。
//
// 本包刻意不 import LandcOrg：路由契约通过 meta.Meta 的 path / method 声明，
// 与 LandcOrg 的 api/org_role/v1.GetUserInheritedRolesRequest 保持一致。
// 契约一致性由 tests/sqlite/org_provider_test.go 用假服务端守住（断言实际
// 请求打到的路径与请求体），这样两个仓库都能独立发布。
package orgrole

import (
	"context"
	"time"

	"github.com/LandcLi/LandcPermission/service"
	"github.com/LandcLi/landc-go/frame/pkg/di"
	"github.com/LandcLi/landc-go/frame/pkg/meta"
)

// DefaultTimeout 是单次组织角色查询的上限。
//
// service.OrgRoleProvider 的方法签名没有 context —— 它是 RBAC 热路径上的同步
// 调用，加 ctx 会穿透整个 service 契约（GetUserRoles / CheckPermission 等都要
// 跟着改）。因此这里用 HTTP 层超时兜底：LandcOrg 无响应时最多阻塞这么久，
// 之后按「组织服务不可用」降级为只用直接角色，不会拖死权限判定。
const DefaultTimeout = 3 * time.Second

// ServiceKeyHeader 是发给 LandcOrg 的服务间凭证头。
// 必须与 LandcOrg 中间件读取的头名一致。
const ServiceKeyHeader = "X-Service-Key"

// getUserInheritedRolesRequest 对应 LandcOrg 的
// api/org_role/v1.GetUserInheritedRolesRequest，路径与之一致。
type getUserInheritedRolesRequest struct {
	meta.Meta `path:"/api/org-role/get-user-inherited-roles" method:"POST"`
	UserID    uint `json:"user_id"`
}

type getUserInheritedRolesResponse struct {
	RoleIDs []uint `json:"role_ids"`
}

// RemoteProvider 通过 HTTP 从 LandcOrg 查询用户继承的角色 ID。
type RemoteProvider struct {
	client *di.ProxyClient
}

var _ service.OrgRoleProvider = (*RemoteProvider)(nil)

// NewRemoteProvider 创建指向 LandcOrg 的 provider。
// 未显式传入 di.WithTimeout 时使用 DefaultTimeout。
func NewRemoteProvider(baseURL string, opts ...di.RemoteOption) *RemoteProvider {
	all := make([]di.RemoteOption, 0, len(opts)+1)
	all = append(all, di.WithTimeout(DefaultTimeout))
	all = append(all, opts...) // 调用方的选项后应用，可覆盖默认超时
	return &RemoteProvider{client: di.NewProxyClient(baseURL, all...)}
}

// GetUserInheritedRoleIDs 实现 service.OrgRoleProvider。
func (p *RemoteProvider) GetUserInheritedRoleIDs(userID uint) ([]uint, error) {
	resp, err := di.Call[getUserInheritedRolesResponse](
		p.client,
		context.Background(),
		"GetUserInheritedRoles",
		&getUserInheritedRolesRequest{UserID: userID},
	)
	if err != nil {
		return nil, err
	}
	return resp.RoleIDs, nil
}
