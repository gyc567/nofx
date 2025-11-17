#!/bin/bash

# 修复OKX交易所的type字段
echo "🔧 修复OKX交易所配置..."

# 方案：将type恢复为"okx"
# 原因：数据库初始化和系统逻辑依赖特定的type值

# 执行修复（如果需要）
# sqlite3 config.db "UPDATE exchanges SET type = 'okx' WHERE id = 'okx' AND user_id = 'default';"

echo "📋 当前OKX配置："
# sqlite3 config.db "SELECT id, name, type, enabled FROM exchanges WHERE id = 'okx' AND user_id = 'default';"

echo ""
echo "✅ 正确的配置应该是："
echo "   id: okx"
echo "   name: OKX Futures"
echo "   type: okx  (不是 cex!)"
echo "   enabled: 0 (需要启用并配置API密钥)"
