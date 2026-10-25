package sqlite_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	framedb "github.com/LandcLi/landc-go/frame/pkg/db"

	landcpermission "github.com/LandcLi/LandcPermission"
	"github.com/LandcLi/LandcPermission/integration/orgrole"
	permissionsvc "github.com/LandcLi/LandcPermission/internal/service_impl/permission"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/service"
)

// resetEffectiveCache 把缓存实例重建为 ttl，同时清空所有条目。
// 缓存是包级单例，测试之间必须显式重置，否则相互污染。
func resetEffectiveCache(t *testing.T, ttl time.Duration) {
	t.Helper()
	permissionsvc.SetEffectiveCacheTTL(ttl)
	t.Cleanup(func() { permissionsvc.SetEffectiveCacheTTL(30 * time.Second) })
}

// TestEffectivePermissionCacheServesStaleUntilReset 证明缓存真的在起作用。
//
// 手法：绕过服务层直接写库授权（不做任何失效），此时判定结果必须仍是旧的；
// 重置缓存后立刻能看到新授权。若缓存没生效，第二次判定就会直接放行。
func TestEffectivePermissionCacheServesStaleUntilReset(t *testing.T) {
	setupSQLiteEnv(t)
	resetEffectiveCache(t, 30*time.Second)

	svc := service.GetPermissionService()
	perm := createPermission(t, "订单导出", "cache_order_export", "order", "export")

	const userID uint = 9101

	allowed, err := svc.CheckPermission(userID, "order", "export")
	if err != nil {
		t.Fatalf("check #1: %v", err)
	}
	if allowed {
		t.Fatal("未授权时不应放行")
	}

	// 绕过服务层写库：不触发任何缓存失效
	if err := framedb.GetDB().Create(&model.UserPermission{
		UserID:       userID,
		PermissionID: perm.ID,
		Reason:       "测试：绕过服务层写入",
	}).Error; err != nil {
		t.Fatalf("直写授权记录: %v", err)
	}

	allowed, err = svc.CheckPermission(userID, "order", "export")
	if err != nil {
		t.Fatalf("check #2: %v", err)
	}
	if allowed {
		t.Fatal("缓存未生效：绕过服务层的写入立刻被读到了")
	}

	// 重置缓存 → 回源，应看到新授权
	permissionsvc.SetEffectiveCacheTTL(30 * time.Second)
	allowed, err = svc.CheckPermission(userID, "order", "export")
	if err != nil {
		t.Fatalf("check #3: %v", err)
	}
	if !allowed {
		t.Fatal("缓存重置后应能读到新授权")
	}
}

// TestWriteThroughInvalidatesCache 验证服务层写操作会主动失效缓存。
// 否则授权后要等 TTL 到期才生效，用户会觉得「授予了没反应」。
func TestWriteThroughInvalidatesCache(t *testing.T) {
	setupSQLiteEnv(t)
	resetEffectiveCache(t, 60*time.Second) // 故意用长 TTL，靠失效而非过期

	svc := service.GetPermissionService()
	perm := createPermission(t, "报表查看", "cache_report_view", "report", "view")

	const userID uint = 9102

	if allowed, _ := svc.CheckPermission(userID, "report", "view"); allowed {
		t.Fatal("初始不应放行")
	}

	// 授予 → 立刻可见
	if err := svc.GrantUserPermissions(userID, []model.UserPermission{{PermissionID: perm.ID}}); err != nil {
		t.Fatalf("grant: %v", err)
	}
	if allowed, err := svc.CheckPermission(userID, "report", "view"); err != nil || !allowed {
		t.Fatalf("授予后应立即放行，allowed=%v err=%v", allowed, err)
	}

	// 撤销 → 立刻不可见
	if err := svc.RevokeUserPermissions(userID, []uint{perm.ID}); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if allowed, err := svc.CheckPermission(userID, "report", "view"); err != nil || allowed {
		t.Fatalf("撤销后应立即拒绝，allowed=%v err=%v", allowed, err)
	}
}

// TestCheckPermissionsBatch 验证批量判定：一次调用判断多项，结果顺序与请求一致。
func TestCheckPermissionsBatch(t *testing.T) {
	setupSQLiteEnv(t)
	resetEffectiveCache(t, 30*time.Second)

	svc := service.GetPermissionService()
	exportPerm := createPermission(t, "订单导出", "batch_order_export", "order", "export")
	_ = createPermission(t, "用户查看", "batch_user_view", "user", "view")

	const userID uint = 9103
	if err := svc.GrantUserPermissions(userID, []model.UserPermission{{PermissionID: exportPerm.ID}}); err != nil {
		t.Fatalf("grant: %v", err)
	}

	checks := []model.PermissionCheck{
		{Resource: "order", Action: "export"},
		{Resource: "user", Action: "view"},
		{Resource: "order", Action: "delete"},
	}
	results, err := svc.CheckPermissions(userID, checks)
	if err != nil {
		t.Fatalf("CheckPermissions: %v", err)
	}
	if len(results) != len(checks) {
		t.Fatalf("结果条数 = %d, want %d", len(results), len(checks))
	}

	want := []bool{true, false, false}
	for i, r := range results {
		if r.Allowed != want[i] {
			t.Errorf("第 %d 项 %s.%s: allowed=%v, want %v", i, r.Resource, r.Action, r.Allowed, want[i])
		}
		if r.Resource != checks[i].Resource || r.Action != checks[i].Action {
			t.Errorf("第 %d 项回带的 resource/action 错位: %s.%s", i, r.Resource, r.Action)
		}
	}

	// 空请求应返回空结果而不是报错
	empty, err := svc.CheckPermissions(userID, nil)
	if err != nil {
		t.Fatalf("空 checks 应正常返回: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("空 checks 应返回 0 条，实际 %d", len(empty))
	}
}

// TestCheckPermissionsBatchOverHTTP 验证批量判定的路由声明与序列化契约。
func TestCheckPermissionsBatchOverHTTP(t *testing.T) {
	setupSQLiteEnv(t)
	resetEffectiveCache(t, 30*time.Second)

	engine := gin.New()
	if err := landcpermission.RegisterHandlers(engine); err != nil {
		t.Fatalf("RegisterHandlers: %v", err)
	}
	srv := httptest.NewServer(engine)
	defer srv.Close()

	perm := createPermission(t, "菜单查看", "batch_menu_view", "menu", "view")
	const userID = 9104

	post := func(path string, payload map[string]interface{}) []byte {
		t.Helper()
		body, _ := json.Marshal(payload)
		resp, err := http.Post(srv.URL+path, "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("post %s: %v", path, err)
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("post %s status=%d raw=%s", path, resp.StatusCode, string(raw))
		}
		return raw
	}

	post("/api/permission/grant-direct", map[string]interface{}{
		"user_id": userID,
		"grants":  []map[string]interface{}{{"permission_id": perm.ID}},
	})

	raw := post("/api/permission/check-batch", map[string]interface{}{
		"user_id": userID,
		"checks": []map[string]string{
			{"resource": "menu", "action": "view"},
			{"resource": "menu", "action": "delete"},
		},
	})

	// 框架的响应中间件会把结果包进 {code,message,data} 信封
	var resp struct {
		Data struct {
			Results []struct {
				Resource string `json:"resource"`
				Action   string `json:"action"`
				Allowed  bool   `json:"allowed"`
			} `json:"results"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("解析响应失败: %v raw=%s", err, string(raw))
	}
	results := resp.Data.Results
	if len(results) != 2 {
		t.Fatalf("results 条数 = %d, want 2, raw=%s", len(results), string(raw))
	}
	if !results[0].Allowed {
		t.Errorf("menu.view 应放行，raw=%s", string(raw))
	}
	if results[1].Allowed {
		t.Errorf("menu.delete 应拒绝，raw=%s", string(raw))
	}
}

// TestCacheDisabledByZeroTTL 验证 TTL=0 完全关闭缓存（排障开关）。
func TestCacheDisabledByZeroTTL(t *testing.T) {
	setupSQLiteEnv(t)
	resetEffectiveCache(t, 0)

	svc := service.GetPermissionService()
	perm := createPermission(t, "缓存关闭验证", "cache_off_perm", "cfg", "read")

	const userID uint = 9105
	if allowed, _ := svc.CheckPermission(userID, "cfg", "read"); allowed {
		t.Fatal("初始不应放行")
	}

	// 绕过服务层写库；缓存关闭时下一次判定必须立刻看到
	if err := framedb.GetDB().Create(&model.UserPermission{
		UserID:       userID,
		PermissionID: perm.ID,
	}).Error; err != nil {
		t.Fatalf("直写授权记录: %v", err)
	}

	if allowed, err := svc.CheckPermission(userID, "cfg", "read"); err != nil || !allowed {
		t.Fatalf("缓存关闭时应每次回源，allowed=%v err=%v", allowed, err)
	}
}

// TestInvalidatePermissionCacheAfterOrgChange 覆盖跨服务场景：
// LandcOrg 侧调整了部门角色后不会通知本服务，缓存里的旧判定会一直用到 TTL 过期。
// 运维要求立刻生效时调 InvalidatePermissionCache。
func TestInvalidatePermissionCacheAfterOrgChange(t *testing.T) {
	setupSQLiteEnv(t)
	resetEffectiveCache(t, 60*time.Second) // 长 TTL：不靠过期，只能靠手动失效
	resetOrgProvider(t)

	permSvc := service.GetPermissionService()
	role, err := service.GetRoleService().CreateRole("组织角色", "org_cache_role", "")
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	perm, err := permSvc.CreatePermission("部门查看", "org_cache_dept_view", "dept", "view", "")
	if err != nil {
		t.Fatalf("create permission: %v", err)
	}
	if err := permSvc.AssignPermissionsToRole(role.ID, []uint{perm.ID}); err != nil {
		t.Fatalf("bind: %v", err)
	}

	fake := newFakeOrgService(t)
	fake.roleIDs.Store([]uint{})
	service.OverrideOrgRoleProvider(orgrole.NewRemoteProvider(fake.srv.URL))

	const userID uint = 9106
	if allowed, err := permSvc.CheckPermission(userID, "dept", "view"); err != nil || allowed {
		t.Fatalf("组织侧尚无角色时不应放行, allowed=%v err=%v", allowed, err)
	}

	// 模拟 LandcOrg 侧新配了部门角色（此时不会主动通知本服务）
	fake.roleIDs.Store([]uint{role.ID})

	if allowed, err := permSvc.CheckPermission(userID, "dept", "view"); err != nil || allowed {
		t.Fatalf("缓存期内应仍是旧结果（这正是跨服务变更的延迟窗口）, allowed=%v err=%v", allowed, err)
	}

	// 手动失效后立刻生效
	service.InvalidatePermissionCache()
	if allowed, err := permSvc.CheckPermission(userID, "dept", "view"); err != nil || !allowed {
		t.Fatalf("清缓存后应立刻放行, allowed=%v err=%v", allowed, err)
	}
}
