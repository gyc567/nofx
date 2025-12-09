package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"nofx/api/credits"
	"nofx/api/handlers"
	"nofx/config"
	"nofx/email"
	"nofx/manager"
	"nofx/middleware"
	creditsService "nofx/service/credits"
)

// Server HTTP API服务器
type Server struct {
	router        *gin.Engine
	traderManager *manager.TraderManager
	database      *config.Database
	port          int
	handler       *handlers.BaseHandler
}

// NewServer 创建API服务器
func NewServer(traderManager *manager.TraderManager, database *config.Database, port int) *Server {
	// 设置为Release模式（减少日志输出）
	gin.SetMode(gin.ReleaseMode)

	// 使用gin.New()而不是gin.Default()，以便我们可以自定义中间件顺序
	router := gin.New()

	// 添加Logger中间件
	router.Use(gin.Logger())

	// 启用CORS（必须在Recovery之前，确保即使panic也能设置CORS头）
	router.Use(corsMiddleware())

	// 添加安全头中间件
	router.Use(middleware.SecurityHeadersMiddleware())

	// 添加频率限制中间件（基础限制）
	router.Use(middleware.RateLimitByIP(60, time.Minute))

	// 添加自定义Recovery中间件，确保panic时也返回带CORS头的响应
	router.Use(corsRecoveryMiddleware())

	// 创建积分服务
	creditService := creditsService.NewCreditService(database)
	creditHandler := credits.NewHandler(creditService)
	emailClient := email.NewResendClient()

	// 初始化 BaseHandler
	baseHandler := handlers.NewBaseHandler(traderManager, database, emailClient, creditService, creditHandler)

	s := &Server{
		router:        router,
		traderManager: traderManager,
		database:      database,
		port:          port,
		handler:       baseHandler,
	}

	// 设置路由
	s.setupRoutes()

	return s
}

// corsRecoveryMiddleware 自定义Recovery中间件，确保panic时也返回带CORS头的响应
func corsRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic日志
				log.Printf("❌ Panic recovered: %v", err)

				// 确保CORS头已设置（如果还没设置的话）
				if c.Writer.Header().Get("Access-Control-Allow-Origin") == "" {
					origin := c.Request.Header.Get("Origin")
					if origin != "" {
						c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
						c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
					}
				}

				// 返回500错误
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}()
		c.Next()
	}
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	// 从环境变量获取允许的域名列表，默认为开发环境和Vercel域名
	allowedOrigins := []string{
		// 开发环境
		"http://localhost:3000",
		"http://localhost:5173",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:5173",

		// Vercel部署域名 - 主要实例
		"https://web-3c7a7psvt-gyc567s-projects.vercel.app",
		"https://web-pink-omega-40.vercel.app",
		"https://web-gyc567s-projects.vercel.app",
		"https://web-7jc87z3u4-gyc567s-projects.vercel.app",
		"https://web-gyc567-gyc567s-projects.vercel.app",
		// 新部署实例 - 2025-11-26
		"https://agentrade-nd2sevhec-gyc567s-projects.vercel.app",

		// Vercel部署域名 - 历史实例
		"https://web-fej4rs4y2-gyc567s-projects.vercel.app",
		"https://web-fco5upt1e-gyc567s-projects.vercel.app",
		"https://web-2ybunmaej-gyc567s-projects.vercel.app",
		"https://web-ge79k4nzy-gyc567s-projects.vercel.app",
		// 生产前端域名（含www和不含www）
		"https://www.agentrade.xyz",
		"https://agentrade.xyz",

		// Replit部署域名
		"https://nofx-gyc567.replit.app",
	}

	// 如果设置了环境变量，使用环境变量中的值
	if envOrigins := os.Getenv("ALLOWED_ORIGINS"); envOrigins != "" {
		allowedOrigins = strings.Split(envOrigins, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查origin是否在白名单中
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		// 对白名单域名设置CORS头（必须在任何响应之前设置）
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// 始终设置这些头，确保预检请求和错误响应都包含CORS头
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Authorization, Cache-Control, X-Requested-With, X-Requested-By, If-Modified-Since, Pragma, Origin, Accept")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// Root health check for Replit deployment
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "Monnaire Trading Agent OS",
		})
	})

	// Serve static files from web/dist for production
	s.router.Static("/assets", "./web/dist/assets")
	s.router.StaticFile("/index.html", "./web/dist/index.html")

	// Catch-all route for SPA routing - serve index.html for non-API routes
	s.router.NoRoute(func(c *gin.Context) {
		// If the request is for an API route that doesn't exist, return 404 JSON
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		// Otherwise serve the frontend index.html for SPA routing
		c.File("./web/dist/index.html")
	})

	// API路由组
	api := s.router.Group("/api")
	{
		// 健康检查
		api.Any("/health", s.handler.HandleHealth)

		// 认证相关路由（无需认证）
		api.POST("/register", s.handler.HandleRegister)
		api.POST("/login", s.handler.HandleLogin)

		// 密码重置路由（无需认证）
		api.POST("/request-password-reset", s.handler.HandleRequestPasswordReset)
		api.POST("/reset-password", s.handler.HandleResetPassword)

		// 系统支持的模型和交易所（无需认证）
		api.GET("/supported-models", s.handler.HandleGetSupportedModels)
		api.GET("/supported-exchanges", s.handler.HandleGetSupportedExchanges)

		// 系统配置（无需认证）
		api.GET("/config", s.handler.HandleGetSystemConfig)

		// 系统提示词模板管理（无需认证）
		api.GET("/prompt-templates", s.handler.HandleGetPromptTemplates)
		api.GET("/prompt-templates/:name", s.handler.HandleGetPromptTemplate)

		// 积分系统 - 公开接口（无需认证，但有频率限制）
		creditPublic := api.Group("/")
		creditPublic.Use(middleware.RateLimitByIP(60, time.Minute)) // 每分钟最多60次查询
		{
			creditPublic.GET("/credit-packages", s.handler.CreditHandler.HandleGetCreditPackages)
			creditPublic.GET("/credit-packages/:id", s.handler.CreditHandler.HandleGetCreditPackage)
		}

		// 公开的竞赛数据（无需认证）
		api.GET("/traders", s.handler.HandlePublicTraderList)
		api.GET("/competition", s.handler.HandlePublicCompetition)
		api.GET("/top-traders", s.handler.HandleTopTraders)
		api.GET("/equity-history", s.handler.HandleEquityHistory)
		api.POST("/equity-history-batch", s.handler.HandleEquityHistoryBatch)
		api.GET("/traders/:id/public-config", s.handler.HandleGetPublicTraderConfig)

		// 需要认证的路由
		protected := api.Group("/", s.handler.AuthMiddleware())
		{
			// AI交易员管理
			protected.GET("/my-traders", s.handler.HandleTraderList)
			protected.GET("/traders/:id/config", s.handler.HandleGetTraderConfig)
			protected.POST("/traders", s.handler.HandleCreateTrader)
			protected.PUT("/traders/:id", s.handler.HandleUpdateTrader)
			protected.DELETE("/traders/:id", s.handler.HandleDeleteTrader)
			protected.POST("/traders/:id/start", s.handler.HandleStartTrader)
			protected.POST("/traders/:id/stop", s.handler.HandleStopTrader)
			protected.PUT("/traders/:id/prompt", s.handler.HandleUpdateTraderPrompt)

			// AI模型配置
			protected.GET("/models", s.handler.HandleGetModelConfigs)
			protected.PUT("/models", s.handler.HandleUpdateModelConfigs)

			// 交易所配置
			protected.GET("/exchanges", s.handler.HandleGetExchangeConfigs)
			protected.PUT("/exchanges", s.handler.HandleUpdateExchangeConfigs)

			// 用户信号源配置
			protected.GET("/user/signal-sources", s.handler.HandleGetUserSignalSource)
			protected.POST("/user/signal-sources", s.handler.HandleSaveUserSignalSource)

			// 指定trader的数据（使用query参数 ?trader_id=xxx）
			protected.GET("/status", s.handler.HandleStatus)
			protected.GET("/account", s.handler.HandleAccount)
			protected.GET("/positions", s.handler.HandlePositions)
			protected.GET("/decisions", s.handler.HandleDecisions)
			protected.GET("/decisions/latest", s.handler.HandleLatestDecisions)
			protected.GET("/statistics", s.handler.HandleStatistics)
			protected.GET("/performance", s.handler.HandlePerformance)

			// 用户管理
			protected.GET("/users", s.handler.HandleGetUsers)

			// 积分系统 - 用户接口（需要认证，有用户级别的频率限制）
			creditUser := protected.Group("/user/")
			creditUser.Use(middleware.RateLimitByUser(10, time.Minute)) // 每分钟最多10次积分操作
			{
				creditUser.GET("/credits", s.handler.CreditHandler.HandleGetUserCredits)
				creditUser.GET("/credits/transactions", s.handler.CreditHandler.HandleGetUserTransactions)
				creditUser.GET("/credits/summary", s.handler.CreditHandler.HandleGetUserCreditSummary)
			}
		}

		// 管理员接口（需要认证和管理员权限）
		admin := api.Group("/admin/")
		admin.Use(s.handler.AuthMiddleware())
		admin.Use(s.handler.AdminMiddleware())
		{
			// 积分套餐管理（管理员级别频率限制）
			creditAdmin := admin.Group("/")
			creditAdmin.Use(middleware.RateLimitAdmin(30, time.Minute)) // 管理员每分钟最多30次操作
			{
				creditAdmin.POST("/credit-packages", s.handler.CreditHandler.HandleCreateCreditPackage)
				creditAdmin.PUT("/credit-packages/:id", s.handler.CreditHandler.HandleUpdateCreditPackage)
				creditAdmin.DELETE("/credit-packages/:id", s.handler.CreditHandler.HandleDeleteCreditPackage)

				// 用户积分管理
				creditAdmin.POST("/users/:id/credits/adjust", s.handler.CreditHandler.HandleAdjustUserCredits)
				creditAdmin.GET("/users/:id/credits", s.handler.CreditHandler.HandleGetUserCreditsByAdmin)
				creditAdmin.GET("/users/:id/credits/transactions", s.handler.CreditHandler.HandleGetUserTransactionsByAdmin)
			}
		}
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 绑定到 0.0.0.0 确保可以从外部访问
	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	log.Printf("🌐 API服务器启动在 http://0.0.0.0:%d", s.port)
	log.Printf("📊 API文档:")
	log.Printf("  • GET  /api/health           - 健康检查")
	log.Printf("  • GET  /api/traders          - 公开的AI交易员排行榜前50名（无需认证）")
	log.Printf("  • GET  /api/competition      - 公开的竞赛数据（无需认证）")
	log.Printf("  • GET  /api/top-traders      - 前5名交易员数据（无需认证，表现对比用）")
	log.Printf("  • GET  /api/equity-history?trader_id=xxx - 公开的收益率历史数据（无需认证，竞赛用）")
	log.Printf("  • GET  /api/equity-history-batch?trader_ids=a,b,c - 批量获取历史数据（无需认证，表现对比优化）")
	log.Printf("  • GET  /api/traders/:id/public-config - 公开的交易员配置（无需认证，不含敏感信息）")
	log.Printf("  • POST /api/traders          - 创建新的AI交易员")
	log.Printf("  • DELETE /api/traders/:id    - 删除AI交易员")
	log.Printf("  • POST /api/traders/:id/start - 启动AI交易员")
	log.Printf("  • POST /api/traders/:id/stop  - 停止AI交易员")
	log.Printf("  • GET  /api/models           - 获取AI模型配置")
	log.Printf("  • PUT  /api/models           - 更新AI模型配置")
	log.Printf("  • GET  /api/exchanges        - 获取交易所配置")
	log.Printf("  • PUT  /api/exchanges        - 更新交易所配置")
	log.Printf("  • GET  /api/status?trader_id=xxx     - 指定trader的系统状态")
	log.Printf("  • GET  /api/account?trader_id=xxx    - 指定trader的账户信息")
	log.Printf("  • GET  /api/positions?trader_id=xxx  - 指定trader的持仓列表")
	log.Printf("  • GET  /api/decisions?trader_id=xxx  - 指定trader的决策日志")
	log.Printf("  • GET  /api/decisions/latest?trader_id=xxx - 指定trader的最新决策")
	log.Printf("  • GET  /api/statistics?trader_id=xxx - 指定trader的统计信息")
	log.Printf("  • GET  /api/performance?trader_id=xxx - 指定trader的AI学习表现分析")
	log.Println()
	log.Printf("✅ API服务器就绪，等待请求...")

	return s.router.Run(addr)
}