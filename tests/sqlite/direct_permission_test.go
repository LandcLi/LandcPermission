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

	landcpermission "github.com/LandcLi/LandcPermission"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/service"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
)

// createPermission 通过 service 层建一个权限点，返回实体。
func createPermission(t *testing.T, name, code, resource, action string) *model.Permission {
	t.Helper()
	perm, err := service.GetPermissionService().CreatePermission(name, code, resource, action, "")
	if err != nil {
		t.Fatalf("create permission %q: %v", code, err)
	}
	return perm
}

// TestDirectPermissionGrantsAccess 验证核心语义：
// 用户没有任何角色，仅凭直接授权即可通过权限校验。
func TestDirectPermissionGrantsAccess(t *testing.T) {
	setupSQLiteEnv(t)

	svc := service.GetPermissionService()
	perm := createPermission(t, "订单导出", "order_export", "order", "export")

	const userID uint = 9001

	allowed, err := svc.CheckPermission(userID, "order", "export")
	if err != nil {
		t.Fatalf("check before grant: %v", err)
	}
	if allowed {
		t.Fatal("unexpected allow: user has neither role nor direct grant")
	}

	if err := svc.GrantUserPermissions(userID, []model.UserPermission{
		{PermissionID: perm.ID, Reason: "线上故障应急授权"},
	}); err != nil {
		t.Fatalf("grant: %v", err)
	}

	allowed, err = svc.CheckPermission(userID, "order", "export")
	if err != nil {
		t.Fatalf("check after grant: %v", err)
	}
	if !allowed {
		t.Fatal("expected allow after direct grant")
	}

	// 有效权限并集里应包含该权限
	effective, err := svc.GetUserEffectivePermissions(userID)
	if err != nil {
		t.Fatalf("effective permissions: %v", err)
	}
	if !containsPermission(effective, perm.ID) {
		t.Errorf("effective permissions should contain %d, got %+v", perm.ID, effective)
	}

	// 撤销后应恢复拒绝
	if err := svc.RevokeUserPermissions(userID, []uint{perm.ID}); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	allowed, err = svc.CheckPermission(userID, "order", "export")
	if err != nil {
		t.Fatalf("check after revoke: %v", err)
	}
	if allowed {
		t.Fatal("expected deny after revoke")
	}
}

// TestExpiredDirectPermissionIsIgnored 验证过期授权在校验阶段被过滤，
// 不依赖任何后台清理任务。
func TestExpiredDirectPermissionIsIgnored(t *testing.T) {
	setupSQLiteEnv(t)

	svc := service.GetPermissionService()
	perm := createPermission(t, "审计日志读取", "audit_read", "audit", "read")

	const userID uint = 9002
	past := time.Now().Add(-time.Minute)

	// 直接落库：service 层会拒绝「出生即过期」的授权，这里模拟自然到期
	if err := framedb.GetDB().Create(&model.UserPermission{
		UserID:       userID,
		PermissionID: perm.ID,
		ExpireAt:     &past,
		Reason:       "已到期的临时授权",
	}).Error; err != nil {
		t.Fatalf("seed expired grant: %v", err)
	}

	allowed, err := svc.CheckPermission(userID, "audit", "read")
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if allowed {
		t.Fatal("expired direct grant must not grant access")
	}

	// 管理端仍应看到这条记录，且被标记为已过期
	details, err := svc.GetUserDirectPermissionDetails(userID)
	if err != nil {
		t.Fatalf("details: %v", err)
	}
	if len(details) != 1 {
		t.Fatalf("expected 1 grant record, got %d", len(details))
	}
	if !details[0].Expired {
		t.Error("record should be marked as expired")
	}
	if details[0].PermissionCode != "audit_read" {
		t.Errorf("permission code not enriched, got %q", details[0].PermissionCode)
	}
}

// TestDirectPermissionNotLeakedAcrossUsers 验证过期条件不会把其他用户的授权带出来。
// 这是一个容易踩的 SQL 优先级陷阱：Where("a") + Where("b OR c") 若不加括号，
// 会被解析成 (a AND b) OR c。
func TestDirectPermissionNotLeakedAcrossUsers(t *testing.T) {
	setupSQLiteEnv(t)

	svc := service.GetPermissionService()
	perm := createPermission(t, "财务对账", "finance_reconcile", "finance", "reconcile")

	const ownerID uint = 9003
	const otherID uint = 9004
	future := time.Now().Add(time.Hour)

	// 只给 ownerID 授权
	if err := svc.GrantUserPermissions(ownerID, []model.UserPermission{
		{PermissionID: perm.ID, ExpireAt: &future},
	}); err != nil {
		t.Fatalf("grant: %v", err)
	}

	// otherID 不应因此获得任何直接权限
	allowed, err := svc.CheckPermission(otherID, "finance", "reconcile")
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if allowed {
		t.Fatal("direct grant leaked across users")
	}

	directIDs, err := service.GetPermissionService().GetUserDirectPermissions(otherID)
	if err != nil {
		t.Fatalf("get direct permissions: %v", err)
	}
	if len(directIDs) != 0 {
		t.Fatalf("other user should have no direct grants, got %d", len(directIDs))
	}
}

// TestGrantIsIdempotent 验证重复授予同一权限为幂等操作，且刷新过期时间与原因。
func TestGrantIsIdempotent(t *testing.T) {
	setupSQLiteEnv(t)

	svc := service.GetPermissionService()
	perm := createPermission(t, "配置写入", "config_write", "config", "write")

	const userID uint = 9005
	first := time.Now().Add(time.Hour)
	second := time.Now().Add(2 * time.Hour)

	if err := svc.GrantUserPermissions(userID, []model.UserPermission{
		{PermissionID: perm.ID, ExpireAt: &first, Reason: "首次授权"},
	}); err != nil {
		t.Fatalf("first grant: %v", err)
	}
	if err := svc.GrantUserPermissions(userID, []model.UserPermission{
		{PermissionID: perm.ID, ExpireAt: &second, Reason: "延长授权"},
	}); err != nil {
		t.Fatalf("second grant: %v", err)
	}

	var count int64
	if err := framedb.GetDB().Model(&model.UserPermission{}).
		Where("user_id = ? AND permission_id = ?", userID, perm.ID).
		Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 grant record after duplicate grant, got %d", count)
	}

	details, err := svc.GetUserDirectPermissionDetails(userID)
	if err != nil {
		t.Fatalf("details: %v", err)
	}
	if len(details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(details))
	}
	if details[0].Reason != "延长授权" {
		t.Errorf("reason should be refreshed, got %q", details[0].Reason)
	}
	if details[0].ExpireAt == nil || !details[0].ExpireAt.After(first) {
		t.Errorf("expire_at should be extended to the later time, got %v", details[0].ExpireAt)
	}
}

// TestGrantRejectsInvalidInput 验证入参校验：不存在的权限、已过去的过期时间都要拒绝。
func TestGrantRejectsInvalidInput(t *testing.T) {
	setupSQLiteEnv(t)

	svc := service.GetPermissionService()

	t.Run("unknown permission", func(t *testing.T) {
		err := svc.GrantUserPermissions(9006, []model.UserPermission{
			{PermissionID: 999999},
		})
		if err == nil {
			t.Fatal("expected error for unknown permission id")
		}
	})

	t.Run("past expire_at", func(t *testing.T) {
		perm := createPermission(t, "临时读取", "tmp_read", "tmp", "read")
		past := time.Now().Add(-time.Second)
		err := svc.GrantUserPermissions(9007, []model.UserPermission{
			{PermissionID: perm.ID, ExpireAt: &past},
		})
		if err == nil {
			t.Fatal("expected error for past expire_at")
		}
	})

	t.Run("empty grants", func(t *testing.T) {
		if err := svc.GrantUserPermissions(9008, nil); err != nil {
			t.Fatalf("empty grants should be a no-op, got %v", err)
		}
	})
}

// TestDirectPermissionOverHTTP 走完整 HTTP 链路：授予 → 校验 → 查询直接授权。
func TestDirectPermissionOverHTTP(t *testing.T) {
	setupSQLiteEnv(t)

	engine := gin.New()
	if err := landcpermission.RegisterHandlers(engine); err != nil {
		t.Fatalf("RegisterHandlers: %v", err)
	}
	srv := httptest.NewServer(engine)
	defer srv.Close()

	perm := createPermission(t, "报表下载", "report_download", "report", "download")
	const userID = 9100

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

	// 1. 授予
	post("/api/permission/grant-direct", map[string]interface{}{
		"user_id": userID,
		"grants": []map[string]interface{}{
			{"permission_id": perm.ID, "reason": "HTTP 链路测试"},
		},
	})

	// 2. 校验应放行
	raw := post("/api/permission/check", map[string]interface{}{
		"user_id":  userID,
		"resource": "report",
		"action":   "download",
	})
	if !bytes.Contains(raw, []byte(`"allowed":true`)) {
		t.Fatalf("expected allowed=true, raw=%s", string(raw))
	}

	// 3. 查询直接授权，应带权限编码
	raw = post("/api/permission/get-direct", map[string]interface{}{"user_id": userID})
	if !bytes.Contains(raw, []byte("report_download")) {
		t.Fatalf("expected permission code in response, raw=%s", string(raw))
	}
}

func containsPermission(perms []model.Permission, id uint) bool {
	for _, p := range perms {
		if p.ID == id {
			return true
		}
	}
	return false
}
