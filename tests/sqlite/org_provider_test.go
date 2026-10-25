package sqlite_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/LandcLi/LandcPermission/integration/orgrole"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/service"
)

// noopOrgProvider 用于测试结束后恢复「未集成 LandcOrg」的状态，
// 避免 DI 中的远程 provider 泄漏到同包的其他测试。
type noopOrgProvider struct{}

func (noopOrgProvider) GetUserInheritedRoleIDs(uint) ([]uint, error) { return nil, nil }

func resetOrgProvider(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { service.OverrideOrgRoleProvider(noopOrgProvider{}) })
}

// fakeOrgService 模拟 LandcOrg 的「查询用户继承角色」端点。
//
// 它同时充当契约守门人：断言 provider 实际请求的路径、方法与请求体，
// 这样「Permission 侧的 path 声明」与「LandcOrg 侧的 meta 声明」一旦不一致，
// 这里就会红，而不必让两个仓库互相依赖。
type fakeOrgService struct {
	srv     *httptest.Server
	roleIDs atomic.Value
	failing atomic.Bool
	hits    atomic.Int32
}

func newFakeOrgService(t *testing.T) *fakeOrgService {
	t.Helper()

	f := &fakeOrgService{}
	f.roleIDs.Store([]uint{}) // 默认空数组，贴近真实响应

	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.hits.Add(1)

		if r.URL.Path != "/api/org-role/get-user-inherited-roles" {
			t.Errorf("provider 请求了错误路径 %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("provider 使用了错误方法 %q", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("读取请求体失败: %v", err)
		}
		var req struct {
			UserID uint `json:"user_id"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("请求体不是合法 JSON（%s）: %v", body, err)
		}
		if req.UserID == 0 {
			t.Errorf("请求体缺少 user_id: %s", body)
		}

		if f.failing.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"org service unavailable"}`))
			return
		}

		ids, _ := f.roleIDs.Load().([]uint)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    10000,
			"message": "success",
			"data":    map[string]interface{}{"role_ids": ids},
		})
	}))
	t.Cleanup(f.srv.Close)
	return f
}

// TestInheritedOrgRolesGrantPermission 验证「组织继承角色」真的参与权限判定。
//
// 这是 LandcPermission 与 LandcOrg 协同的核心语义：用户没有任何直接角色，
// 仅因为在组织架构中处于某个部门 / 岗位（由 Org 侧算出其继承的角色）而获得权限。
func TestInheritedOrgRolesGrantPermission(t *testing.T) {
	setupSQLiteEnv(t)
	resetOrgProvider(t)

	permSvc := service.GetPermissionService()
	roleSvc := service.GetRoleService()

	role, err := roleSvc.CreateRole("组织继承角色", "org_inherited", "由部门/岗位继承")
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	perm, err := permSvc.CreatePermission("部门查看", "dept_view", "dept", "view", "")
	if err != nil {
		t.Fatalf("create permission: %v", err)
	}
	if err := permSvc.AssignPermissionsToRole(role.ID, []uint{perm.ID}); err != nil {
		t.Fatalf("bind permission to role: %v", err)
	}

	const userID uint = 7001

	allowed, err := permSvc.CheckPermission(userID, "dept", "view")
	if err != nil {
		t.Fatalf("check before wiring: %v", err)
	}
	if allowed {
		t.Fatal("用户既无直接角色也无组织归属，不应通过")
	}

	// 接入 LandcOrg：该用户在组织架构中继承了上面的角色
	fake := newFakeOrgService(t)
	fake.roleIDs.Store([]uint{role.ID})
	service.OverrideOrgRoleProvider(orgrole.NewRemoteProvider(fake.srv.URL))

	allowed, err = permSvc.CheckPermission(userID, "dept", "view")
	if err != nil {
		t.Fatalf("check after wiring: %v", err)
	}
	if !allowed {
		t.Fatal("接入 LandcOrg 后，组织继承的角色应使权限校验通过")
	}
	if fake.hits.Load() == 0 {
		t.Error("provider 未被调用，说明协同链路没有真正生效")
	}
}

// TestOrgProviderFailureFallsBackToDirectRoles 验证降级语义：
// LandcOrg 不可用时，权限判定不能报错，也不能把已授予的权限一起丢掉。
func TestOrgProviderFailureFallsBackToDirectRoles(t *testing.T) {
	setupSQLiteEnv(t)
	resetOrgProvider(t)

	permSvc := service.GetPermissionService()
	perm := createPermission(t, "订单查看", "order_view", "order", "view")

	const userID uint = 7002
	if err := permSvc.GrantUserPermissions(userID, []model.UserPermission{{PermissionID: perm.ID}}); err != nil {
		t.Fatalf("grant direct permission: %v", err)
	}

	fake := newFakeOrgService(t)
	fake.failing.Store(true)
	service.OverrideOrgRoleProvider(orgrole.NewRemoteProvider(fake.srv.URL))

	allowed, err := permSvc.CheckPermission(userID, "order", "view")
	if err != nil {
		t.Fatalf("组织服务故障不应让权限判定返回错误，实际: %v", err)
	}
	if !allowed {
		t.Fatal("降级后仍应使用用户的直接权限")
	}
}
