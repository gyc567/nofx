#!/bin/bash

# 🔍 登陆 401 错误综合诊断脚本
# 用途: 快速诊断和排除登陆故障

set -e

API_URL="${API_URL:-https://nofx-gyc567.replit.app}"
EMAIL="gyc567@gmail.com"
PASSWORD="eric8577HH"

echo "🔍 [登陆故障诊断] 开始..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 环境信息:"
echo "  API_URL: $API_URL"
echo "  邮箱: $EMAIL"
echo "  密码: $PASSWORD"
echo ""

# 1️⃣ 检查服务可用性
echo "1️⃣ 检查服务可用性"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if curl -s --connect-timeout 5 "$API_URL/api/health" > /dev/null 2>&1; then
    echo "✅ API 服务正常运行"
else
    echo "❌ API 服务无响应"
    echo "   请检查:"
    echo "   1. 服务是否启动"
    echo "   2. URL 是否正确"
    exit 1
fi

echo ""

# 2️⃣ 检查 beta_mode 配置
echo "2️⃣ 检查 beta_mode 配置"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

CONFIG_RESPONSE=$(curl -s "$API_URL/api/config")
BETA_MODE=$(echo "$CONFIG_RESPONSE" | grep -o '"beta_mode":[^,}]*' | cut -d':' -f2)

if [ "$BETA_MODE" = "false" ]; then
    echo "✅ beta_mode: false (内测模式关闭)"
elif [ "$BETA_MODE" = "true" ]; then
    echo "⚠️  beta_mode: true (内测模式开启)"
    echo "   ⚠️  用户需要有有效的内测码才能登陆"
else
    echo "❓ beta_mode: 无法读取"
fi

echo ""

# 3️⃣ 测试登陆
echo "3️⃣ 测试登陆请求"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

HTTP_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
BODY=$(echo "$LOGIN_RESPONSE" | head -n-1)

echo "HTTP 状态码: $HTTP_CODE"
echo ""
echo "响应体:"
echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
echo ""

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ 登陆成功!"
    TOKEN=$(echo "$BODY" | jq -r '.token' 2>/dev/null)
    if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
        echo "   Token: ${TOKEN:0:20}..."
    fi
elif [ "$HTTP_CODE" = "401" ]; then
    echo "🔴 登陆失败 (401 Unauthorized)"
    echo ""
    echo "可能原因:"
    echo "  1. 邮箱或密码错误"
    echo "  2. 用户不存在"
    echo "  3. beta_mode 开启且用户无内测码"
    echo "  4. 账户被禁用"
elif [ "$HTTP_CODE" = "400" ]; then
    echo "⚠️  请求格式错误 (400)"
    echo "   请检查邮箱和密码格式"
else
    echo "❓ 未知错误 ($HTTP_CODE)"
fi

echo ""

# 4️⃣ 检查用户是否存在 (如果有管理员权限)
echo "4️⃣ 诊断建议"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ "$HTTP_CODE" = "401" ]; then
    echo ""
    echo "如果您有数据库访问权限，运行以下SQL语句来检查用户:"
    echo ""
    echo "  SELECT id, email, is_active, beta_code FROM users"
    echo "  WHERE email = '$EMAIL';"
    echo ""
    echo "预期结果:"
    echo "  • 如果没有返回行 → 用户未注册，需要注册"
    echo "  • 如果 is_active=false → 账户被禁用"
    echo "  • 如果 beta_mode=true 且 beta_code=null → 需要添加内测码"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 诊断完成"
