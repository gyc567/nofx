package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nofx/config"
)

// HandleGetUsers 处理获取用户列表请求
func (h *BaseHandler) HandleGetUsers(c *gin.Context) {
	// 解析参数
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	search := c.Query("search")
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	// 验证排序字段
	validSortFields := []string{"created_at", "email"}
	sortValid := false
	for _, field := range validSortFields {
		if sort == field {
			sortValid = true
			break
		}
	}
	if !sortValid {
		sort = "created_at"
	}

	// 验证排序方向
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// 权限检查
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未认证的访问",
		})
		return
	}

	currentUser := user.(*config.User)
	if !currentUser.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "权限不足，需要管理员权限",
		})
		return
	}

	// 调用数据库方法
	users, total, err := h.Database.GetUsers(page, limit, search, sort, order)
	if err != nil {
		log.Printf("获取用户列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取用户列表失败",
		})
		return
	}

	// 计算分页信息
	totalPages := (total + limit - 1) / limit // 向上取整
	hasNext := page < totalPages
	hasPrev := page > 1

	// 构建响应
	response := gin.H{
		"users": users,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	}

	// 记录访问日志
	log.Printf("管理员 %s 查询用户列表 (page=%d, limit=%d, search=%s, sort=%s, order=%s)",
		currentUser.Email, page, limit, search, sort, order)

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "获取用户列表成功",
	})
}
