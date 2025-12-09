package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"nofx/auth"
)

// AuthMiddleware JWT认证中间件
func (h *BaseHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// OPTIONS预检请求不需要认证，直接放行
		// CORS预检请求由corsMiddleware处理
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// 检查是否开启admin模式
		adminModeStr, _ := h.Database.GetSystemConfig("admin_mode")
		isAdminMode := adminModeStr == "true"

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 如果是admin模式，使用admin用户
			if isAdminMode {
				user, err := h.Database.GetUserByID("admin")
				if err != nil {
					log.Printf("获取admin用户失败: %v", err)
					c.JSON(http.StatusUnauthorized, gin.H{"error": "admin用户不存在"})
					c.Abort()
					return
				}
				c.Set("user", user)
				c.Set("user_id", "admin")
				c.Next()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少Authorization头"})
			c.Abort()
			return
		}

		// 检查Bearer token格式
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Authorization格式"})
			c.Abort()
			return
		}

		// 验证JWT token
		claims, err := auth.ValidateJWT(tokenParts[1])
		if err != nil {
			// 如果是admin模式且token验证失败，使用admin用户
			if isAdminMode {
				user, err := h.Database.GetUserByID("admin")
				if err != nil {
					log.Printf("获取admin用户失败: %v", err)
					c.JSON(http.StatusUnauthorized, gin.H{"error": "admin用户不存在"})
					c.Abort()
					return
				}
				c.Set("user", user)
				c.Set("user_id", "admin")
				c.Next()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token: " + err.Error()})
			c.Abort()
			return
		}

		// 获取完整的用户信息
		user, err := h.Database.GetUserByID(claims.UserID)
		if err != nil {
			log.Printf("获取用户信息失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的用户",
			})
			c.Abort()
			return
		}

		// 将完整的用户对象存储到上下文中
		c.Set("user", user)
		// 为了向后兼容，同时保留user_id
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func (h *BaseHandler) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := ""
		if uid, exists := c.Get("userID"); exists {
			userID = uid.(string)
		}

		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户未认证",
			})
			c.Abort()
			return
		}

		// TODO: 实现管理员权限检查
		// 这里简化处理，实际应该查询数据库验证管理员身份
		// 暂时允许所有认证用户（测试用）
		c.Next()
	}
}
