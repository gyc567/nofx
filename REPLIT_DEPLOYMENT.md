# 🚀 NOFX后端部署到Replit指南

## 📋 部署可行性分析

### ✅ **完全支持部署**

NOFX后端程序**完全可以部署到Replit.com**，原因：

- ✅ **纯Go实现** - 无C库依赖，技术指标使用纯Go计算
- ✅ **Go 1.25** - Replit原生支持
- ✅ **Go Modules** - 依赖自动下载
- ✅ **端口暴露** - 8080端口可在Replit中暴露
- ✅ **环境变量** - 支持机密信息配置

---

## 🎯 部署步骤

### **Step 1: 创建Replit项目**

1. **访问 [Replit.com](https://replit.com)**
2. **点击 "Create Repl"**
3. **选择 "Go" 作为语言**
4. **项目名称**: `nofx-backend` (或任意名称)

### **Step 2: 上传项目文件**

将NOFX项目文件上传到Replit：

**方法一: 打包上传**
```bash
# 在本地项目目录
tar -czf nofx.tar.gz --exclude='.git' --exclude='node_modules' --exclude='web' .

# 在Replit中上传并解压
```

**方法二: 复制粘贴**
1. 在Replit中创建文件结构
2. 将以下文件内容复制到对应文件：
   - `main.go` (根目录)
   - `go.mod`
   - `go.sum`
   - `config/` 目录
   - `trader/` 目录
   - `api/` 目录
   - `manager/` 目录
   - `market/` 目录
   - `pool/` 目录
   - `decision/` 目录
   - `logger/` 目录
   - `mcp/` 目录

### **Step 3: 配置环境变量**

在Replit项目中，点击左侧 **"🔒 Secrets"** 按钮，添加以下环境变量：

#### **必需配置**

```bash
# 币安API配置 (Binance API)
BINANCE_API_KEY=your_binance_api_key_here
BINANCE_SECRET_KEY=your_binance_secret_key_here

# DeepSeek AI配置 (DeepSeek AI)
DEEPSEEK_API_KEY=sk-your_deepseek_api_key_here

# 可选: Qwen AI配置 (Qwen AI)
QWEN_API_KEY=sk-your_qwen_api_key_here
QWEN_API_URL=https://dashscope.aliyuncs.com/compatible-mode/v1

# 可选: Hyperliquid配置
HYPERLIQUID_PRIVATE_KEY=your_ethereum_private_key_without_0x
HYPERLIQUID_WALLET_ADDR=your_ethereum_address

# 可选: Aster DEX配置
ASTER_USER=0xYourMainWalletAddress
ASTER_SIGNER=0xYourApiWalletAddress
ASTER_PRIVATE_KEY=your_aster_api_wallet_private_key
```

**⚠️ 安全提示**: 这些密钥非常重要，请妥善保管！

#### **API密钥获取方法**

**1. 币安API (Binance)**
- 注册 [Binance](https://www.binance.com)
- 进入 "API Management"
- 创建新API Key
- 启用 "Futures" 权限
- 复制 API Key 和 Secret Key

**2. DeepSeek AI**
- 注册 [platform.deepseek.com](https://platform.deepseek.com)
- 创建API Key
- 充值账户（最低$5）

**3. Qwen AI (可选)**
- 注册 [dashscope.aliyuncs.com](https://dashscope.aliyuncs.com)
- 创建API Key

### **Step 4: 创建配置文件**

在Replit中创建 `config.json` 文件：

```json
{
  "traders": [
    {
      "id": "replit_trader",
      "name": "Replit DeepSeek Trader",
      "enabled": true,
      "ai_model": "deepseek",
      "exchange": "binance",
      "binance_api_key": "${BINANCE_API_KEY}",
      "binance_secret_key": "${BINANCE_SECRET_KEY}",
      "deepseek_key": "${DEEPSEEK_API_KEY}",
      "initial_balance": 1000.0,
      "scan_interval_minutes": 3
    }
  ],
  "leverage": {
    "btc_eth_leverage": 5,
    "altcoin_leverage": 5
  },
  "use_default_coins": true,
  "default_coins": [
    "BTCUSDT",
    "ETHUSDT",
    "SOLUSDT",
    "BNBUSDT",
    "XRPUSDT",
    "DOGEUSDT",
    "ADAUSDT"
  ],
  "api_server_port": 8080
}
```

**💡 提示**: Replit会自动将Secrets中的环境变量注入到 `${VARIABLE_NAME}` 格式的字符串中。

### **Step 5: 运行程序**

在Replit的 **Shell** 面板中执行：

```bash
# 1. 下载依赖
go mod download

# 2. 构建程序
go build -o nofx

# 3. 运行程序
./nofx
```

**或者使用Replit的自动运行**:
- 点击 "Run" 按钮
- 设置运行命令为: `go run main.go config.json`

### **Step 6: 暴露端口**

在Replit的 **"🔌 Ports"** 面板中：
1. 点击 "Add Port"
2. 端口号: `8080`
3. 命名为: `NOFX API`
4. 选择 "Public" 访问

**访问地址**: `https://your-repl-name.your-username.repl.co`

---

## 🔧 优化配置

### **自动运行设置**

在Replit根目录创建 `.replit` 文件：

```toml
run = "go run main.go config.json"

[deployment]
run = ["sh", "-c", "go run main.go config.json"]
deploymentTarget = "autostart"
```

### **Nix配置 (可选)**

如果需要系统依赖，创建 `nix.conf`：

```nix
{ pkgs }: {
  deps = [
    pkgs.go-1-25
    pkgs.gcc
  ];
}
```

---

## 📡 API访问

### **本地访问**
```bash
# 健康检查
curl https://your-repl-name.your-username.repl.co/health

# 获取账户信息
curl https://your-repl-name.your-username.repl.co/api/account

# 获取持仓
curl https://your-repl-name.your-username.repl.co/api/positions
```

### **前端配置**

如果部署前端，需修改 `web/src/lib/api.ts`:

```typescript
const API_BASE = 'https://your-repl-name.your-username.repl.co/api'
```

---

## ⚠️ 注意事项

### **1. Replit限制**

- **冷启动**: 长时间不活跃会进入休眠
- **计算资源**: 免费版CPU/内存有限
- **网络**: 币安API可能有访问限制
- **持久化**: 重启后文件丢失（需上传到GitHub）

### **2. 安全建议**

- **不要在代码中硬编码API密钥**
- **定期轮换API密钥**
- **为API设置IP白名单**
- **使用子账户进行测试**

### **3. 性能优化**

- **减少API调用频率** (已在代码中配置3分钟间隔)
- **合理设置超时**
- **监控内存使用**

### **4. 故障排除**

**构建失败**:
```bash
# 清理缓存
go clean -cache

# 重新下载依赖
rm -rf go.sum
go mod download
```

**运行错误**:
```bash
# 检查配置文件
cat config.json

# 检查环境变量
echo $BINANCE_API_KEY
```

**端口无法访问**:
- 确认Ports面板中已添加8080端口
- 检查防火墙设置
- 确认程序正在运行

---

## 📊 监控建议

### **日志监控**

程序运行日志会显示在Replit控制台：

```
✓ 配置加载成功，共1个trader参赛
✓ 已启用默认主流币种列表（共7个币种）
🤖 AI全权决策模式:
  • AI将自主决定每笔交易的杠杆倍数
✓ Hyperliquid交易器初始化成功
✓ API服务器启动在端口 8080
📊 开始交易监控...
```

### **API监控**

定期检查API健康状态：

```bash
# 监控脚本
while true; do
  curl -s https://your-repl-name.repl.co/health || echo "API Down!"
  sleep 60
done
```

---

## 🆘 常见问题

### **Q1: 程序启动失败，提示 "config.json not found"**

A: 确认配置文件存在且格式正确

```bash
ls -la config.json
cat config.json
```

### **Q2: API密钥无效错误**

A: 检查Secrets配置是否正确

```bash
# 在Replit控制台检查
env | grep BINANCE
```

### **Q3: 币安API调用失败**

A: 可能原因：
- API密钥权限不足
- IP未添加到白名单
- 网络访问限制（尝试使用代理）

### **Q4: 程序运行一段时间后停止**

A: Replit免费版的限制，建议：
- 升级到付费版
- 或使用云服务器部署

---

## 🔄 更新程序

### **方法一: GitHub集成**

1. 将代码推送到GitHub
2. 在Replit中连接GitHub仓库
3. 设置自动部署

### **方法二: 重新上传**

1. 本地修改代码
2. 重新上传到Replit
3. 点击 "Run" 重启

---

## 📈 扩展建议

### **多实例部署**
- 在Replit中创建多个副本
- 使用不同配置文件
- 实现负载均衡

### **数据持久化**
- 连接云数据库（如PostgreSQL）
- 使用Redis缓存市场数据
- 存储历史决策日志

### **监控告警**
- 集成Prometheus + Grafana
- 设置告警规则
- 邮件/钉钉通知

---

## ✅ 部署检查清单

- [ ] Replit项目已创建
- [ ] 所有文件已上传
- [ ] 环境变量已配置
- [ ] config.json已创建
- [ ] 程序编译成功
- [ ] 程序运行正常
- [ ] 端口已暴露
- [ ] API可正常访问
- [ ] 日志输出正常

---

## 🎉 成功！

恭喜！NOFX后端已成功部署到Replit。

**访问地址**: `https://your-repl-name.your-username.repl.co`

现在你可以：
- 📊 监控AI交易决策
- 🔍 查看市场数据分析
- 📈 追踪交易表现
- 🤖 观察AI学习过程

**注意**: AI交易有风险，请谨慎使用！

---

## 📚 相关资源

- [Replit官方文档](https://docs.replit.com/)
- [Go官方文档](https://golang.org/doc/)
- [币安API文档](https://binance-docs.github.io/apidocs/futures/en/)
- [NOFX GitHub](https://github.com/tinkle-community/nofx)

---

**最后更新**: 2025-11-11
**支持版本**: NOFX v2.0.2+
