# LandcPermission

基于 RBAC 模型的权限管理系统，提供角色管理、权限管理、用户-角色绑定、权限校验等功能，支持组织架构驱动的自动授权。

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
- 查询用户的所有角色和有效权限

### 权限校验
- `CheckPermission(userID, resource, action)` — 一行代码完成权限校验
- 聚合直接分配的角色 + 组织继承的角色（需集成 LandcOrg）
- 通配符资源匹配

### 组织驱动授权（可选）
- 通过 `OrgRoleProvider` 接口对接 LandcOrg
- 自动聚合用户从部门/岗位继承的角色
- 未注册 Provider 时仅使用直接分配的角色（向后兼容）

## 技术栈

- **框架**: landc-go/mvc（DI / Gateway / Web）
- **Web框架**: Gin
- **数据库**: MySQL (GORM)
- **缓存**: Redis
- **日志**: landc-go/log
- **API规范**: landc-go/api

## 项目结构

```
LandcPermission/
├── api/                        # API接口定义
│   ├── permission/             # 权限管理接口（含角色绑定、用户绑定、权限校验）
│   └── role/                   # 角色管理接口
├── dao/                        # 数据访问层接口
├── internal/                   # 内部实现（可通过 DI Override 替换）
│   ├── controller/             # 控制器层
│   ├── dao_impl/              # DAO实现层
│   └── service_impl/          # 服务实现层
├── model/                      # 数据模型
├── sdk/                        # HTTP SDK（远程调用）
├── service/                    # 服务层接口
├── sqls/                       # SQL脚本
├── utility/                    # 工具包
│   ├── config/                # 配置管理
│   ├── db/                    # 数据库连接
│   └── logger/                # 日志管理
├── cmd/api/main.go            # 独立运行入口
└── LandcPermission.go          # 主入口文件（支持嵌入集成）
```

## 快速开始

### 1. 初始化数据库

```bash
mysql -u root -p < sqls/init.sql
```

### 2. 配置文件

```bash
cp config.example.yml config.yml
# 修改数据库、Redis等配置
```

### 3. 独立运行

```go
package main

import (
    "fmt"
    "log"
    landcpermission "github.com/LandcLi/LandcPermission"
)

func main() {
    engine, err := landcpermission.InitDefault("config.yml")
    if err != nil {
        log.Fatalf("Failed to initialize: %v", err)
    }
    fmt.Println("LandcPermission starting on :8082")
    engine.Run(":8082")
}
```

### 4. 嵌入到其他系统

```go
// 本地集成模式（嵌入到 SSO 或其他系统中）
permCfg := &config.Config{...}
landcpermission.InitWithConfigAndGin(permCfg, ginEngine)
```

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
| CheckPermission | POST /api/permission/check | 权限校验 |

## 权限模型

```
用户 ──(user_roles)──> 角色 ──(role_permissions)──> 权限(resource + action)
                         ↑
         组织继承（可选） │
                         │
用户 ──(org_members)──> 部门 ──(department_roles)──> 角色
                    └──> 岗位 ──(position_roles)──> 角色
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

## 许可证

MIT License
