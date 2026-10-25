package sqlite_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"

	"github.com/gin-gonic/gin"

	landcpermission "github.com/LandcLi/LandcPermission"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/utility/config"
)

const configTemplate = `
server:
  port: 0
  host: "127.0.0.1"
  use_default_routes: true
database:
  driver: "sqlite"
  dsn: "%s"
  auto_migrate: true
redis:
  addr: ""
jwt:
  secret: "sqlite-smoke-test-secret-key-at-least-32-chars"
  issuer: "landc-permission-test"
log:
  level: "error"
  output: "console"
other:
  server:
    gin_mode: "test"
    use_default_route: true
`

// setupSQLiteEnv 用临时目录中的 SQLite 完成一次完整初始化。
func setupSQLiteEnv(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "data", "permission.db")
	configPath := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(fmt.Sprintf(configTemplate, filepath.ToSlash(dbPath))), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := frameconfig.InitGlobalConfigWithPathAndEnv(configPath); err != nil {
		t.Fatalf("init framework config: %v", err)
	}
	t.Cleanup(func() { _ = framedb.Close() })

	// 与生产 initInfrastructure 顺序一致：先连库，再建表
	if err := framedb.InitGlobalDBWithDefault(); err != nil {
		t.Fatalf("init db: %v", err)
	}
	if err := landcpermission.EnsureSchema(); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}
}

// TestSQLiteSchemaCreated 验证 auto_migrate 建出了 RBAC 四张表。
func TestSQLiteSchemaCreated(t *testing.T) {
	setupSQLiteEnv(t)

	migrator := framedb.GetDB().Migrator()
	for _, m := range landcpermission.Models() {
		if !migrator.HasTable(m) {
			t.Errorf("table missing after EnsureSchema: %T", m)
		}
	}
}

// TestInitDefaultSQLite 验证完整入口（InitDefault）在 SQLite 下可用。
func TestInitDefaultSQLite(t *testing.T) {
	setupSQLiteEnv(t)

	cfg := config.GetConfig()
	if cfg == nil {
		t.Fatal("config should be loaded by framework")
	}
	server, err := landcpermission.InitWithConfig(cfg)
	if err != nil {
		t.Fatalf("InitWithConfig: %v", err)
	}
	if server == nil || server.Engine() == nil {
		t.Fatal("expected non-nil server")
	}

	// 控制器路由应从 meta 标签自动生成
	found := false
	for _, r := range server.Routes() {
		if r.Path == "/api/role/create" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected route /api/role/create to be registered")
	}
}

// TestLibraryMode 验证库模式：把控制器挂到宿主已有 gin 引擎，且宿主可继续追加路由。
func TestLibraryMode(t *testing.T) {
	setupSQLiteEnv(t)

	engine := gin.New()
	if err := landcpermission.RegisterHandlers(engine); err != nil {
		t.Fatalf("RegisterHandlers: %v", err)
	}

	// 宿主自定义路由
	called := false
	engine.GET("/host/ping", func(c *gin.Context) {
		called = true
		c.String(http.StatusOK, "pong")
	})

	srv := httptest.NewServer(engine)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/host/ping")
	if err != nil {
		t.Fatalf("get host route: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !called {
		t.Errorf("host route should remain reachable, status=%d", resp.StatusCode)
	}
}

// TestCreateRoleOverHTTP 走完整 HTTP 链路创建角色，验证 SQLite 落库。
func TestCreateRoleOverHTTP(t *testing.T) {
	setupSQLiteEnv(t)

	engine := gin.New()
	if err := landcpermission.RegisterHandlers(engine); err != nil {
		t.Fatalf("RegisterHandlers: %v", err)
	}
	srv := httptest.NewServer(engine)
	defer srv.Close()

	body, _ := json.Marshal(map[string]string{
		"name":        "管理员",
		"code":        "admin",
		"description": "sqlite smoke",
	})
	resp, err := http.Post(srv.URL+"/api/role/create", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post create role: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create role status = %d, raw=%s", resp.StatusCode, string(raw))
	}

	var count int64
	if err := framedb.GetDB().Model(&model.Role{}).Where("code = ?", "admin").Count(&count).Error; err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 role persisted, got %d (raw=%s)", count, string(raw))
	}
}
