# Changelog

本文件记录 LandcPermission 的显著变更。
格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

## [0.2.0] - 2026-02-05

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
