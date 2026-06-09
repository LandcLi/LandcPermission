package api

import (
	"github.com/LandcLi/LandcPermission/api/permission"
	"github.com/LandcLi/LandcPermission/api/role"
	"github.com/LandcLi/landc-go/frame/pkg/response"
	"github.com/gin-gonic/gin"
)

// RegisterDefaultRoutes registers all default routes using Gateway auto-generation.
func RegisterDefaultRoutes(r *gin.Engine) {
	// Health check
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})
}

// GetRoleController returns the role controller implementation.
func GetRoleController() role.RoleController {
	return role.RoleGateway.Get()
}

// GetPermissionController returns the permission controller implementation.
func GetPermissionController() permission.PermissionController {
	return permission.PermissionGateway.Get()
}
