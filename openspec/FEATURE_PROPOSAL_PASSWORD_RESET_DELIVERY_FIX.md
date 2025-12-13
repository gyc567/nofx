# 密码重置邮件递送系统 - 功能修复与增强提案

**提案编号**: FP-2025-001-PASSWORD-RESET-DELIVERY
**优先级**: 🔴 **高** (用户阻塞问题)
**工作量**: 3-5天
**复杂度**: 🟢 **低** (配置修复 + 质量增强)
**依赖**: 无

---

## 📋 提案摘要

当前密码重置系统存在**邮件未送达问题**，用户无法收到重置邮件。本提案通过以下方案解决:

1. **诊断工具** - 快速定位问题
2. **错误可见性** - 详细日志和监控
3. **自动恢复** - 重试机制和队列
4. **追踪系统** - 邮件发送日志
5. **故障排查** - 完整的文档和工具

---

## 🎯 目标与成功指标

### 主要目标

✅ **解决邮件未送达问题**
- 确保用户能收到密码重置邮件
- 预期送达率: 99%+

✅ **快速故障诊断**
- 从故障发生到定位原因 < 5分钟
- 管理员能清楚看到邮件发送状态

✅ **自动化恢复**
- 临时性故障自动重试
- 不需要人工干预

✅ **用户体验改进**
- 清晰的反馈信息
- 知道邮件已发送

### 成功指标

| 指标 | 目标 | 衡量方式 |
|------|------|---------|
| 邮件送达率 | 99%+ | 监控面板 |
| 故障诊断时间 | < 5分钟 | 日志分析 |
| 自动恢复成功率 | 95%+ | 邮件队列统计 |
| 用户投诉率 | < 0.1% | 用户反馈 |
| 系统可用性 | 99.9% | 健康检查 |

---

## 🔍 需求分析

### 功能需求 (FRs)

**FR1: 邮件发送诊断**
- 标准: 系统能识别邮件发送失败的原因
- 实现: 详细的错误日志和分类

**FR2: 邮件发送健康检查**
- 标准: 提供API端点检查邮件服务状态
- 实现: `GET /api/health/email`

**FR3: 邮件发送重试**
- 标准: 临时性故障自动重试 (指数退避)
- 实现: 最多重试3次，延迟: 1s, 2s, 4s

**FR4: 邮件发送追踪**
- 标准: 记录每一封邮件的发送状态
- 实现: `email_logs` 表和相关API

**FR5: 异步邮件队列**
- 标准: 邮件发送不阻塞用户请求
- 实现: 后台工作进程处理邮件队列

### 非功能需求 (NFRs)

**NFR1: 安全性**
- 不泄露用户信息 (继续使用通用成功消息)
- API密钥不出现在错误消息中

**NFR2: 性能**
- 邮件发送延迟 < 100ms (使用后台队列)
- 健康检查响应 < 500ms

**NFR3: 可靠性**
- 自动重试失败邮件
- 邮件队列持久化 (数据库)

**NFR4: 可维护性**
- KISS原则: 代码简洁易懂
- 高内聚低耦合的架构
- 完整的故障排查文档

**NFR5: 监控性**
- 清晰的日志分类
- 可查询的邮件发送历史
- 实时的邮件队列状态

---

## 🛠 技术方案

### 方案概述

```
用户请求
   ↓
[验证邮箱] → [生成令牌] → [保存DB] → [入队邮件]
                                          ↓
                              [后台工作进程]
                                 ↓
                          [发送邮件(重试)]
                                 ↓
                          [记录邮件日志]
                                 ↓
                          [更新队列状态]
```

### 修复1: 错误日志增强

**文件**: `/api/server.go`

```go
func (s *Server) handleRequestPasswordReset(c *gin.Context) {
	// ... 存在检查 ...

	// 发送密码重置邮件
	err = s.emailClient.SendPasswordResetEmail(req.Email, token, frontendURL)
	if err != nil {
		// ✅ 改进: 详细的诊断日志
		log.Printf("🔴 [PASSWORD_RESET_FAILED] 邮件发送失败")
		log.Printf("   用户邮箱: %s", req.Email)
		log.Printf("   错误: %v", err)
		log.Printf("   诊断信息:")
		log.Printf("     - API Key配置: %t", s.emailClient.HasAPIKey())
		log.Printf("     - 发件人地址: %s", s.emailClient.GetFromEmail())
		log.Printf("     - 前端URL: %s", frontendURL)

		// 可选: 发送告警通知
		s.alertManager.SendCritical("邮件服务故障", err)
	} else {
		log.Printf("✅ [PASSWORD_RESET_SUCCESS] 邮件已发送")
		log.Printf("   收件人: %s", req.Email)
	}

	// 返回用户消息 (保持一致，防止邮箱枚举)
	c.JSON(http.StatusOK, gin.H{
		"message": "如果该邮箱已注册，您将收到密码重置邮件",
	})
}
```

### 修复2: 健康检查端点

**文件**: `/api/server.go`

```go
// HandleEmailHealthCheck 邮件服务健康检查
func (s *Server) handleEmailHealthCheck(c *gin.Context) {
	// 检查API Key配置
	if !s.emailClient.HasAPIKey() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"reason": "RESEND_API_KEY未配置",
		})
		return
	}

	// 可选: 发送测试邮件
	// testEmail := "healthcheck@example.com"
	// err := s.emailClient.SendTestEmail(testEmail)
	// if err != nil {
	//     c.JSON(http.StatusServiceUnavailable, ...)
	//     return
	// }

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "email",
		"provider": "resend",
		"timestamp": time.Now(),
	})
}

// 在init中注册路由
router.GET("/api/health/email", s.handleEmailHealthCheck)
```

### 修复3: 邮件重试机制

**文件**: `/email/email.go`

```go
// SendEmailWithRetry 带重试的邮件发送
func (c *ResendClient) SendEmailWithRetry(to, subject, html, text string) error {
	const maxRetries = 3
	const baseDelay = time.Second

	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := c.SendEmail(to, subject, html, text)
		if err == nil {
			return nil  // 成功
		}

		lastErr = err
		log.Printf("⚠️  邮件发送失败 (尝试 %d/%d): %v", attempt, maxRetries, err)

		// 最后一次尝试后不再延迟
		if attempt < maxRetries {
			// 指数退避: 1s, 2s, 4s
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			log.Printf("   %v后重试...", delay)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("邮件发送失败，已重试%d次: %w", maxRetries, lastErr)
}
```

### 修复4: 邮件发送日志表

**文件**: `/config/database.go`

```go
func (d *Database) CreateEmailLogsTable() error {
	schema := `
	CREATE TABLE IF NOT EXISTS email_logs (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		email_type TEXT NOT NULL,          -- 'password_reset', 'verification', etc
		recipient TEXT NOT NULL,
		status TEXT DEFAULT 'pending',     -- 'pending', 'sent', 'failed', 'bounced'
		error_message TEXT,
		attempt_count INT DEFAULT 1,
		last_attempted_at TIMESTAMP,
		sent_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

		INDEX idx_user (user_id),
		INDEX idx_email_type (email_type),
		INDEX idx_status (status),
		INDEX idx_created_at (created_at),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	return d.db.Exec(schema).Error
}

// LogEmailAttempt 记录邮件发送尝试
func (d *Database) LogEmailAttempt(userID, emailType, recipient string, err error) error {
	log := &EmailLog{
		ID:        uuid.New().String(),
		UserID:    userID,
		EmailType: emailType,
		Recipient: recipient,
		Status:    "pending",
	}

	if err != nil {
		log.Status = "failed"
		log.ErrorMessage = err.Error()
	} else {
		log.Status = "sent"
		log.SentAt = time.Now()
	}

	log.LastAttemptedAt = time.Now()
	return d.db.Create(log).Error
}

// GetEmailLogs 查询邮件发送历史
func (d *Database) GetEmailLogs(userID, emailType string, limit int) ([]EmailLog, error) {
	var logs []EmailLog
	return logs, d.db.
		Where("user_id = ? AND email_type = ?", userID, emailType).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
}

// GetPendingEmails 获取待发送的邮件
func (d *Database) GetPendingEmails(limit int) ([]EmailLog, error) {
	var logs []EmailLog
	return logs, d.db.
		Where("status = ?", "pending").
		Order("created_at ASC").
		Limit(limit).
		Find(&logs).Error
}

// UpdateEmailStatus 更新邮件状态
func (d *Database) UpdateEmailStatus(id, status string, err error) error {
	update := map[string]interface{}{
		"status": status,
		"updated_at": time.Now(),
	}

	if err != nil {
		update["error_message"] = err.Error()
		update["attempt_count"] = gorm.Expr("attempt_count + 1")
	} else {
		update["sent_at"] = time.Now()
	}

	return d.db.Model(&EmailLog{}).Where("id = ?", id).Updates(update).Error
}
```

### 修复5: 邮件队列和后台工作进程

**文件**: `/service/email_queue_service.go` (新建)

```go
package service

import (
	"context"
	"fmt"
	"log"
	"time"
)

// EmailQueue 邮件队列服务
type EmailQueue struct {
	db         *Database
	email      *email.ResendClient
	maxWorkers int
	stopped    chan bool
}

// NewEmailQueue 创建邮件队列
func NewEmailQueue(db *Database, email *email.ResendClient, maxWorkers int) *EmailQueue {
	return &EmailQueue{
		db:         db,
		email:      email,
		maxWorkers: maxWorkers,
		stopped:    make(chan bool, maxWorkers),
	}
}

// Start 启动邮件队列处理
func (eq *EmailQueue) Start(ctx context.Context) {
	for i := 0; i < eq.maxWorkers; i++ {
		go eq.worker(ctx, i)
	}
	log.Printf("✅ 邮件队列已启动 (工作进程: %d)", eq.maxWorkers)
}

// worker 单个工作进程
func (eq *EmailQueue) worker(ctx context.Context, id int) {
	ticker := time.NewTicker(5 * time.Second)  // 每5秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			eq.stopped <- true
			return

		case <-ticker.C:
			// 获取待发送的邮件
			logs, err := eq.db.GetPendingEmails(10)
			if err != nil {
				log.Printf("⚠️  获取待发送邮件失败: %v", err)
				continue
			}

			// 处理每一封邮件
			for _, log := range logs {
				eq.processEmail(log)
			}
		}
	}
}

// processEmail 处理单封邮件
func (eq *EmailQueue) processEmail(log *EmailLog) {
	// 根据邮件类型调用相应的发送函数
	var err error
	switch log.EmailType {
	case "password_reset":
		// 从数据库获取重置令牌
		token := eq.getPasswordResetToken(log.UserID)
		err = eq.email.SendPasswordResetEmail(log.Recipient, token, "https://...")

	case "verification":
		// ... 其他邮件类型

	default:
		err = fmt.Errorf("未知的邮件类型: %s", log.EmailType)
	}

	// 更新邮件状态
	if err != nil {
		log.AttemptCount++
		if log.AttemptCount >= 3 {
			// 超过重试次数，标记为失败
			eq.db.UpdateEmailStatus(log.ID, "failed", err)
			log.Printf("❌ 邮件发送失败 (已放弃): %s", log.Recipient)
		} else {
			// 继续重试
			eq.db.UpdateEmailStatus(log.ID, "pending", err)
			log.Printf("⚠️  邮件发送失败，将重试: %s (尝试 %d/3)", log.Recipient, log.AttemptCount)
		}
	} else {
		// 发送成功
		eq.db.UpdateEmailStatus(log.ID, "sent", nil)
		log.Printf("✅ 邮件发送成功: %s", log.Recipient)
	}
}

// Stop 停止邮件队列
func (eq *EmailQueue) Stop() {
	for i := 0; i < eq.maxWorkers; i++ {
		<-eq.stopped
	}
	log.Printf("✅ 邮件队列已停止")
}
```

### 修复6: 更新密码重置API

**文件**: `/api/server.go`

```go
func (s *Server) handleRequestPasswordReset(c *gin.Context) {
	// ... 验证代码 ...

	// 生成令牌
	token, err := auth.GeneratePasswordResetToken()
	if err != nil {
		log.Printf("生成令牌失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成重置令牌失败",
		})
		return
	}

	// 保存令牌
	tokenHash := auth.HashPasswordResetToken(token)
	expiresAt := time.Now().Add(1 * time.Hour)
	err = s.database.CreatePasswordResetToken(user.ID, token, tokenHash, expiresAt)
	if err != nil {
		log.Printf("保存令牌失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建重置令牌失败",
		})
		return
	}

	// ✅ 异步发送邮件(不阻塞用户请求)
	go func() {
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "https://web-pink-omega-40.vercel.app"
		}

		// 使用重试机制发送
		err := s.emailClient.SendEmailWithRetry(
			req.Email,
			"密码重置 - Monnaire Trading Agent OS",
			generatePasswordResetHTML(fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)),
			generatePasswordResetText(fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)),
		)

		// 记录发送结果
		s.database.LogEmailAttempt(user.ID, "password_reset", req.Email, err)

		if err != nil {
			log.Printf("🔴 密码重置邮件发送失败: %v", err)
		} else {
			log.Printf("✅ 密码重置邮件已发送: %s", req.Email)
		}
	}()

	// 返回用户消息
	c.JSON(http.StatusOK, gin.H{
		"message": "如果该邮箱已注册，您将收到密码重置邮件",
	})
}
```

---

## 📊 测试计划

### 单元测试

```go
func TestEmailHealthCheck(t *testing.T) {
	// 测试API Key配置
	// 测试Resend连接状态
}

func TestEmailRetry(t *testing.T) {
	// 测试重试逻辑
	// 测试指数退避延迟
	// 测试最大重试次数
}

func TestEmailQueueWorker(t *testing.T) {
	// 测试邮件队列处理
	// 测试状态更新
	// 测试错误处理
}

func TestPasswordResetEmail(t *testing.T) {
	// 测试生成令牌
	// 测试保存数据库
	// 测试邮件发送
}
```

### 集成测试

```bash
# 1. 测试完整的密码重置流程
curl -X POST http://localhost:8080/api/request-password-reset \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'

# 2. 检查邮件日志
curl http://localhost:8080/api/admin/email-logs?user_id=xxx

# 3. 检查邮件队列状态
curl http://localhost:8080/api/admin/email-queue

# 4. 检查邮件服务健康
curl http://localhost:8080/api/health/email
```

### E2E测试

- [ ] 用户邮箱收到邮件
- [ ] 点击邮件中的链接能重置密码
- [ ] 超时的令牌无法使用
- [ ] 邮件发送失败自动重试

---

## 🎯 KISS & 架构原则

### 保持简单

✅ **低复杂度**:
- 邮件发送逻辑集中在一个包
- 清晰的错误处理
- 最小化的依赖

✅ **易于理解**:
- 函数名明确表达意图
- 日志消息清晰
- 代码行数适中

### 高内聚

✅ **相关代码在一起**:
- `/email/` - 所有邮件相关
- `/service/email_queue_service.go` - 队列处理
- `/models/email_log.go` - 邮件日志模型

### 低耦合

✅ **依赖注入**:
```go
// ❌ 不好 - 全局变量
var globalEmailClient *ResendClient

// ✅ 好 - 依赖注入
type Server struct {
	emailClient *ResendClient
}
```

✅ **接口而非实现**:
```go
type EmailSender interface {
	Send(to, subject, body string) error
}
```

### 无外部依赖增加

✅ **使用现有依赖**:
- `gorm` - 已在使用
- `standard library` - 优先使用标准库
- `Resend API` - 已在使用

❌ **不引入**:
- 消息队列库 (如RabbitMQ)
- 缓存库 (除非必要)

---

## 📋 实施计划

### 第一天: 诊断 & 快速修复

- [ ] (1h) 检查环境变量配置
- [ ] (1h) 查看错误日志
- [ ] (1h) 修复配置问题
- [ ] (1h) 测试邮件发送
- [ ] (1h) 部署到测试环境

### 第二天: 错误可见性增强

- [ ] (2h) 添加详细诊断日志
- [ ] (2h) 实现健康检查端点
- [ ] (2h) 测试健康检查
- [ ] (1h) 文档更新

### 第三天: 重试机制

- [ ] (2h) 实现邮件重试逻辑
- [ ] (2h) 编写单元测试
- [ ] (1h) 测试重试机制
- [ ] (1h) 部署

### 第四天: 邮件队列

- [ ] (2h) 创建email_logs表
- [ ] (2h) 实现邮件队列服务
- [ ] (2h) 后台工作进程
- [ ] (1h) 集成测试

### 第五天: 文档 & 最终测试

- [ ] (2h) 编写故障排查文档
- [ ] (2h) E2E测试
- [ ] (1h) 生产部署
- [ ] (1h) 监控和验证

---

## 📝 交付物清单

- [ ] 修复后的 `/api/server.go` (健康检查 + 日志增强)
- [ ] 修复后的 `/email/email.go` (重试机制)
- [ ] 新建 `/config/email_log.go` (邮件日志模型)
- [ ] 新建 `/service/email_queue_service.go` (队列处理)
- [ ] 数据库迁移文件 (邮件日志表)
- [ ] 完整的单元测试
- [ ] 集成测试脚本
- [ ] 故障排查文档 (`/docs/EMAIL_TROUBLESHOOTING.md`)
- [ ] 监控配置 (日志告警)

---

## ✨ 预期结果

### 立即改进 (第1天后)
- ✅ 用户能收到密码重置邮件
- ✅ 管理员能看到错误日志
- ✅ 问题定位时间 < 5分钟

### 短期改进 (1周后)
- ✅ 完整的邮件日志追踪
- ✅ 自动重试机制
- ✅ 健康检查端点

### 长期改进 (1个月后)
- ✅ 邮件送达率 99%+
- ✅ 零投诉的故障诊断
- ✅ 完整的可观测性系统

---

**此提案遵循 KISS 原则、高内聚低耦合设计，保证最小化代码变更同时最大化问题解决效果。**

🎯 **目标**: 24小时内修复邮件未送达问题，实现生产级别的可靠性和可观测性。

