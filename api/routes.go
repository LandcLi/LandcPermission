package api

import (
	"github.com/gin-gonic/gin"
)

// RegisterDefaultRoutes registers all default routes using Gateway auto-generation.
// 健康检查已由框架自动注册（config.use_default_routes: true）
func RegisterDefaultRoutes(r *gin.Engine) {
}

// GetRoleController returns the role controller implementation.
func GetRoleController() role.RoleController {
	return role.RoleGateway.Get()
}

// GetPermissionController returns the permission controller implementation.
func GetPermissionController() permission.PermissionController {
	return permission.PermissionGateway.Get()
}
