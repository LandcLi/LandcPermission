# Changelog

本文件记录 LandcPermission 的显著变更。
格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

## [0.2.2] - 2026-10-26

### 新增

- **有效权限判定缓存**。一次 `CheckPermission` 此前要 3~5 次 SQL + 1 次 3s 超时的
  组织服务 RPC，前端渲染菜单一次要判几十项，慢且脆弱。现在按用户缓存有效权限
  （进程内 TTL 缓存，默认 30s），读操作走缓存，写操作（角色/权限/用户授权变更）
  主动失效；`InvalidatePermissionCache()` 供运维入口，TTL 配 0 关闭缓存以排查。
- **批量判定 `POST /api/permission/check-batch`**。一次请求判定最多 200 项
  `{resource, action}`，返回逐项 `allowed`，前端菜单渲染从 N 次往返收敛为 1 次。

### 修复

- **组织角色提供者切换后缓存不失效**。`OverrideOrgRoleProvider` 换数据源后，
  旧的「无权限」判定仍被缓存命中，导致测试 `TestInheritedOrgRolesGrantPermission`
  假失败。现在切换即清缓存，并通过 `RegisterCacheInvalidator` 保持实现包不依赖
  service 包（依赖倒置）。

## [0.2.1] - 2026-10-25

打通与 LandcOrg 的协同：组织继承角色首次真正参与权限判定。

### 新增

- **组织继承角色提供者（`integration/orgrole`）**。`service.OrgRoleProvider` 接口此前
  只有声明、没有任何实现，`OrgRoleProvider` 恒为空 → 组织架构里配的角色对权限判定
  毫无影响（文档却说支持，属于静默失效）。
  - `RemoteProvider` 用 `di.ProxyClient` 调 `/api/org-role/get-user-inherited-roles`
  - **刻意不 import LandcOrg**：接口定义在 Permission、实现也在 Permission，
    靠请求结构体的 `meta.Meta` 契约解耦，LandcOrg 换版本/改实现都不影响本系统编译
  - 默认 3s 超时；组织服务不可用时**降级**为只算用户直接角色，不阻断授权判定
- `service.OverrideOrgRoleProvider(impl)`：用 `di.Override` 强制替换已注册实现。
  `di.Provide` 不覆盖，装配期切换数据来源必须用 Override。
- 配置项 `other.org_service.base_url`：留空则完全不启用组织协同（保持纯 RBAC 行为）。

### 修复

- **配置读取误用 `GetOther` 取嵌套值**：`other.server.gin_mode`、
  `other.org_service.base_url` 等多层配置此前恒取不到（`other` 段在框架侧多套了一层），
  导致 `gin_mode` 从不生效、`base_url` 永远为空。改用框架新增的
  `GetOtherString` / `GetOtherBool`，并删除本系统自建的 `otherLookup` 等实现。

### 测试

- `tests/sqlite/org_provider_test.go`：`fakeOrgService` 充当**契约守门人**
  （断言 path、method、body 里的 `user_id`），覆盖「组织继承角色授予权限」
  与「组织服务故障降级」两条路径。
- `utility/config/config_test.go`：嵌套 `other` 读取、无 `other` 段、类型不符三例。

### 依赖

- 升级 `frame v0.8.6`（含 `other` 段提升、`GetOtherPath`、HTTP 状态码透传）。

## [0.2.0] - 2026-10-25

### 新增

- **用户直接权限（`user_permissions`）**：绕过角色的定向授权能力，用于应急提权、
  外部协作、例外授权等场景。
  - `GrantUserPermissions` / `RevokeUserPermissions` / `GetUserDirectPermissions` /
    `GetUserEffectivePermissions` 四个 service 方法
  - 对应 4 个 HTTP 接口：`/api/permission/grant-direct`、`/revoke-direct`、
    `/get-direct`、`/get-effective`
  - 支持 `expire_at` 过期时间（`NULL` = 永久）与 `reason` 授权原因（供审计）
  - 重复授予幂等：刷新过期时间与原因，不产生重复记录
  - service 层拒绝「不存在的权限」与「已过去的过期时间」两类无效入参
- `model.UserPermissionDetail`：直接授权记录 + 权限编码/名称 + 过期状态，
  由 service 层一次批量补齐，避免调用方 N+1 查询

### 变更

- **有效权限语义扩展**：`CheckPermission` 现在基于
  「角色权限（含组织继承） ∪ 未过期直接权限」，此前仅看角色。
  未使用直接授权的系统行为不变（并集为空）。

### 修复

- **`user_role` / `role_permission` DAO 的缓存永久失效**：原实现把
  `db.GetRedis()` 的结果在包 `init()` 阶段抓取并存入结构体字段，而 Redis 要等
  `main` 之后才初始化，导致 `cache` 字段恒为 `nil`，缓存从不命中也不报错。
  改为惰性获取，并把裸类型断言 `c.(framecache.Cache)` 换成 comma-ok 形式避免 panic。
- **直接权限查询的 SQL 优先级陷阱**：`user_id = ? AND expire_at IS NULL OR
  expire_at > ?` 会被 SQL 解析为 `(user_id = ? AND expire_at IS NULL) OR
  (expire_at > ?)`，从而把**其他用户**的未过期授权一起返回。已显式加括号，
  并补 `TestDirectPermissionNotLeakedAcrossUsers` 回归用例。

## [0.1.0] - 2026-02-04

### 变更

- 适配 `landc-go/frame v0.8.4`
- 移除「临时 Server 复制路由表」的注册 hack，改用框架导出的
  `web.RegisterHandlers(router, instance, opts...)`
- `LandcPermission.go` 统一对外暴露 `RegisterToRouter` / `RegisterHandlers` /
  `Models` / `EnsureSchema`
- 删除失效的 `api/routes.go`（引用了不存在的包，导致构建失败）
- `config.example.yml` 补 `database.driver: sqlite`、`auto_migrate: true`、`jwt.secret`

### 新增

- SQLite 开箱即用（纯 Go 驱动，零外部依赖）
- `tests/sqlite` 冒烟测试：建表 / 默认初始化 / 库模式 / HTTP 建角色
