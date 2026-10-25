# LandcPermission

基于 RBAC 模型的权限管理系统，提供角色管理、权限管理、用户-角色绑定、权限校验等功能，支持组织架构驱动的自动授权。**支持多种权限模型扩展（RBAC、ABAC、ACL）**。

## 功能特性

### 角色管理
- 角色 CRUD（唯一 code 约束）
- 角色启用/禁用

### 权限管理
- 权限 CRUD（resource + action 模型）
- 支持资源通配符匹配（如 `user.*` 匹配 `user.profile`）
- 支持 action 通配符（`admin` 和 `*` 匹配所有操作）

### 授权绑定
- 角色-权限绑定（assign/revoke）
- 用户-角色绑定（assign/revoke）
- **用户直接权限**：绕过角色直接把权限授予某个用户，支持设置过期时间与授权原因
- 查询用户的所有角色和有效权限

### 用户直接权限

角色是「批量授权」的抽象，但总有几类场景套不进角色模型：

- **应急授权**：线上故障时临时给运维开一个写权限，事后自动失效
- **外部协作**：外包/顾问只做一件事，不值得为它建一个角色
- **例外处理**：某人比其角色多一项权限，但不应影响同角色的其他人

`user_permissions` 表承载这类定向授权。最终有效权限的合成规则：

```
有效权限 = 角色权限（含组织继承） ∪ 未过期的直接权限
```

- `expire_at` 为 `NULL` 表示永久有效；非空且已过期时在**校验阶段**被过滤，
  正确性不依赖任何后台清理任务
- 重复授予同一权限是幂等操作，会刷新过期时间与授权原因
- 授权记录保留 `reason` 字段，便于事后审计「这次提权是谁批的、为什么」

### 权限校验
- `CheckPermission(userID, resource, action)` — 一行代码完成权限校验
- `CheckPermissionWithContext(ctx)` — 支持多种权限模型的扩展校验
- 聚合直接分配的角色 + 组织继承的角色 + 未过期的直接权限
- 通配符资源匹配

### 多权限模型支持
- **RBAC**（默认）：基于角色的访问控制
- **ABAC**：基于属性的访问控制（用户属性、资源属性、环境属性）
- **ACL**：访问控制列表
- 通过 DI 机制可灵活切换和扩展权限模型

### 组织驱动授权（可选）
- 通过 `OrgRoleProvider` 接口对接 LandcOrg，自动聚合用户从部门/岗位继承的角色
- 未注册 Provider 时仅使用直接分配的角色（向后兼容，默认关闭）

启用方式：在配置文件里填 `other.org_service.base_url`，服务启动时会自动注入
`integration/orgrole` 的远程适配器：

```yaml
other:
  org_service:
    base_url: "http://127.0.0.1:8083"   # LandcOrg 地址
```

有效权限的完整口径：

```
有效权限 = 直接分配角色的权限 ∪ 组织继承角色的权限 ∪ 未过期直接权限
```

设计要点：
- Permission 不依赖 LandcOrg 的代码，两者通过 HTTP 契约解耦，可独立发布
- 路由契约由 `meta.Meta` 声明（`/api/org-role/get-user-inherited-roles`），
  由 `tests/sqlite/org_provider_test.go` 的假服务端守住一致性
- 查询超时 3 秒；LandcOrg 不可用时自动降级为只用直接角色与直接权限，
  权限判定不报错、不阻塞

## 技术栈

- **框架**：`github.com/LandcLi/landc-go/frame`（配置 / DB / 缓存 / JWT / DI / Web / 健康检查）
- **Web**：Gin；路由由 `meta.Meta` 标签声明，框架自动注册
- **数据库**：SQLite（默认，纯 Go 驱动，零外部依赖）或 MySQL（生产）
- **缓存**：Redis（可选，缺失时自动降级为进程内 LRU）
- **日志**：`landc-go/log`

## 项目结构

```
LandcPermission/
├── api/                        接口定义（请求/响应 + meta 路由标签）
│   ├── permission/             权限管理（含角色绑定、用户绑定、权限校验）
│   └── role/                   角色管理
├── dao/                        数据访问接口
├── internal/                   内部实现（可通过 DI Override 替换）
│   ├── controller/
│   ├── dao_impl/
│   └── service_impl/
├── model/                      数据模型（Role / Permission / RolePermission / UserRole）
├── sdk/                        HTTP SDK（远程调用）
├── service/                    服务接口
├── sqls/                       MySQL 初始化脚本
├── tests/sqlite/               SQLite 端到端冒烟
├── utility/                    配置 / DB / 日志
├── cmd/api/main.go             独立运行入口
└── LandcPermission.go          库入口（InitDefault / InitWithConfig / RegisterHandlers）
```

## 快速开始

### 1. 准备配置并启动

```bash
cp config.example.yaml config.yaml
go run ./cmd/api
```

默认使用 SQLite，首次启动自动在 `data/landc_permission.db` 建表
（`database.auto_migrate: true`），无需 MySQL / Redis。

```bash
curl http://127.0.0.1:8082/health
```

创建角色：

```bash
curl -X POST http://127.0.0.1:8082/api/role/create \
  -H 'Content-Type: application/json' \
  -d '{"name":"管理员","code":"admin","description":"默认管理员角色"}'
```

### 2. 切换到 MySQL（生产）

```yaml
database:
  driver: "mysql"
  dsn: "root:password@tcp(127.0.0.1:3306)/landc_permission?charset=utf8mb4&parseTime=True&loc=Local"
  auto_migrate: false
```

```bash
mysql -u root -p < sqls/init.sql
```

### 3. 嵌入到其他系统（库模式）

```go
// 方式一：挂到宿主已有的 gin 引擎（不自带中间件链，由宿主决定）
if err := landcpermission.RegisterHandlers(ginEngine); err != nil {
    return err
}

// 方式二：连配置与中间件一起装配（使用独立 Server）
server, err := landcpermission.InitDefault("config.yaml")

// 方式三：用显式配置 + 已有 gin 引擎
landcpermission.InitWithConfigAndGin(permCfg, ginEngine)
```

库模式只注册控制器，不会创建 engine、不会调用 `Run`，宿主可继续追加自有路由。

### 4. Docker 部署

```bash
docker build -t landc-permission:0.2.1 .

docker run -d --name landc-permission \
  -p 8082:8082 \
  -v "$(pwd)/data:/app/data" \
  landc-permission:0.2.1
```

镜像要点：

- 多阶段构建，产物是 `CGO_ENABLED=0` 静态二进制（纯 Go SQLite 驱动，运行阶段无需 gcc）
- 以非 root 用户 `app` 运行，`/app/data` 与 `/app/logs` 已预建并授权
- 内置 `HEALTHCHECK` 打 `/health`，可直接被 compose / k8s 探活
- 时区固定 `Asia/Shanghai`（`tzdata` 已安装）——权限的 `expire_at` 是绝对时间，
  缺时区库会导致过期判断偏移

挂载 `data/` 是关键：不挂载时容器重建会丢掉 SQLite 库。


## 使用 SDK（远程调用模式）

```go
import (
    "github.com/LandcLi/LandcPermission/api/permission"
    "github.com/LandcLi/LandcPermission/sdk"
)

// 注册远程实现
permission.PermissionGateway.Override(
    sdk.NewRemotePermissionController("http://permission-service:8082"),
)

// 使用方式与本地完全一致
ctrl := permission.GetPermissionController()
resp, err := ctrl.CheckPermission(ctx, &v1.CheckPermissionRequest{
    UserID:   1,
    Resource: "user.profile",
    Action:   "write",
})
fmt.Println(resp.Allowed) // true or false
```

## API 接口

### 角色接口

| 接口 | 路径 | 描述 |
|------|------|------|
| CreateRole | POST /api/role/create | 创建角色 |
| UpdateRole | POST /api/role/update | 更新角色 |
| DeleteRole | POST /api/role/delete | 删除角色 |
| GetRole | POST /api/role/get | 获取角色详情 |
| ListRoles | POST /api/role/list | 角色列表（分页） |

### 权限接口

| 接口 | 路径 | 描述 |
|------|------|------|
| CreatePermission | POST /api/permission/create | 创建权限 |
| UpdatePermission | POST /api/permission/update | 更新权限 |
| DeletePermission | POST /api/permission/delete | 删除权限 |
| GetPermission | POST /api/permission/get | 获取权限详情 |
| ListPermissions | POST /api/permission/list | 权限列表（分页） |
| AssignPermissions | POST /api/permission/assign-to-role | 给角色分配权限 |
| RevokePermissions | POST /api/permission/revoke-from-role | 撤销角色权限 |
| GetRolePermissions | POST /api/permission/get-role-permissions | 获取角色的所有权限 |
| AssignRoles | POST /api/permission/assign-roles | 给用户分配角色 |
| RevokeRoles | POST /api/permission/revoke-roles | 撤销用户角色 |
| GetUserRoles | POST /api/permission/get-user-roles | 获取用户的所有角色 |
| GetUserPermissions | POST /api/permission/get-user-permissions | 获取用户的有效权限 |
| GrantUserPermissions | POST /api/permission/grant-direct | 授予用户直接权限（可设过期时间） |
| RevokeUserPermissions | POST /api/permission/revoke-direct | 撤销用户直接权限 |
| GetUserDirectPermissions | POST /api/permission/get-direct | 查询用户直接授权记录 |
| GetUserEffectivePermissions | POST /api/permission/get-effective | 查询用户最终有效权限（角色 ∪ 直接） |
| CheckPermission | POST /api/permission/check | 权限校验 |

授予直接权限示例：

```bash
curl -X POST http://127.0.0.1:8082/api/permission/grant-direct \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": 1001,
    "grants": [
      {"permission_id": 7, "expire_at": "2026-10-26T00:00:00Z", "reason": "线上故障应急"}
    ]
  }'
```

## 权限模型

```
用户 ──(user_roles)──> 角色 ──(role_permissions)──> 权限(resource + action)
  │                      ↑
  │      组织继承（可选） │
  │                      │
  ├──(org_members)──> 部门 ──(department_roles)──> 角色
  │                └──> 岗位 ──(position_roles)──> 角色
  │
  └──(user_permissions)──> 权限(resource + action)   ← 直接授权，绕过角色
```

### 权限匹配规则

| 权限定义 | 请求 | 是否匹配 |
|----------|------|----------|
| `resource=user, action=read` | `user, read` | ✅ |
| `resource=user, action=admin` | `user, read` | ✅ (admin 匹配所有) |
| `resource=*, action=admin` | `任意, 任意` | ✅ (超级权限) |
| `resource=user.*, action=write` | `user.profile, write` | ✅ (通配符) |
| `resource=user, action=write` | `user, read` | ❌ |

## 自定义扩展

所有接口层（Controller / Service / DAO）都可以通过 DI Override 替换：

```go
// 替换权限校验逻辑
service.RegisterPermissionService(myCustomPermissionService)

// 替换数据存储
dao.RegisterRoleDAO(myRedisRoleDAO)

// 对接组织架构（从 LandcOrg 获取继承角色）
service.RegisterOrgRoleProvider(orgRoleService)
```

## 多权限模型扩展

### 使用 ABAC 权限模型

```go
import (
    "github.com/LandcLi/LandcPermission/internal/service_impl/abac"
    "github.com/LandcLi/LandcPermission/service"
)

// 定义 ABAC 策略
policies := []service.ABACPolicy{
    {
        ID:          1,
        Name:        "部门数据访问策略",
        Description: "只允许访问本部门数据",
        Conditions: []service.PolicyCondition{
            {
                AttributeType: service.AttributeTypeUser,
                AttributeKey:  "department_id",
                Operator:      service.ConditionOperatorEqual,
                Value:         "resource.department_id", // 动态引用
            },
        },
        Effect:   service.PolicyEffectAllow,
        Priority: 100,
    },
}

// 注册 ABAC 校验器
abac.RegisterABACChecker(policies)

// 使用权限校验（带上下文）
permCtx := model.NewPermissionContext(ctx, userID, "user.profile", "read").
    WithUserAttributes(map[string]interface{}{
        "department_id": 1,
        "level":         "manager",
    }).
    WithResourceAttributes(map[string]interface{}{
        "department_id": 1,
        "sensitivity":   "high",
    }).
    WithEnvironmentAttrs(map[string]interface{}{
        "time": time.Now(),
        "ip":   "192.168.1.1",
    })

allowed, err := service.GetPermissionService().CheckPermissionWithContext(permCtx)
```

### 使用 ACL 权限模型

```go
// 实现 ACLPermissionChecker 接口
type MyACLChecker struct{}

func (c *MyACLChecker) CheckPermission(ctx *model.PermissionContext) (bool, error) {
    return c.CheckACL(ctx.UserID, ctx.Resource, ctx.Action)
}

func (c *MyACLChecker) GetModelType() service.PermissionModelType {
    return service.PermissionModelACL
}

// 注册 ACL 校验器
service.RegisterPermissionChecker(&MyACLChecker{})
```

### API 接口（支持多权限模型）

```bash
# RBAC 基础校验
POST /api/permission/check
{
    "user_id": 1,
    "resource": "user.profile",
    "action": "write"
}

# 多权限模型校验（支持 ABAC/ACL）
POST /api/permission/check-with-context
{
    "user_id": 1,
    "resource": "user.profile",
    "action": "write",
    "tenant_id": 100,
    "attrs": {
        "user_attributes": {
            "department_id": 1,
            "level": "manager"
        },
        "resource_attributes": {
            "owner_id": 1,
            "sensitivity": "high"
        },
        "environment_attrs": {
            "ip": "192.168.1.1",
            "time": "2024-01-01T10:00:00Z"
        }
    }
}
```

## 许可证

MIT License
