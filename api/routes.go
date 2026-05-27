package api

import (
	"github.com/LandcLi/LandcPermission/api/permission"
	"github.com/LandcLi/LandcPermission/api/role"
	"github.com/gin-gonic/gin"
)

// RegisterDefaultRoutes registers all default routes using Gateway auto-generation.
func RegisterDefaultRoutes(r *gin.Engine) {
	// Health check
	r.GET("/health", HealthCheck)

	// Register role routes via Gateway
	role.RoleGateway.RegisterRoutes(r)

	// Register permission routes via Gateway
	permission.PermissionGateway.RegisterRoutes(r)
}

// HealthCheck health check endpoint
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
