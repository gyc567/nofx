# 架构审计报告：DeepSeek 新闻处理提示词优化

**审计人**: 架构师 Agent
**日期**: 2025年12月06日
**审计对象**: `service/news/deepseek.go` 及其 Prompt 设计
**目标**: 提高提示词效率，减少 Token 消耗，增强输出稳定性。

---

## 1. 现状评估 (Status Assessment)

### 1.1 当前 Prompt
```text
You are a professional crypto market analyst.
Task: Translate the news into %s and summarize it.
Input Headline: %s
Input Summary: %s

Output strictly in JSON format:
{
  "title": "Translated Headline",
  "summary": "One sentence summary",
  "sentiment": "POSITIVE/NEGATIVE/NEUTRAL"
}
```
**System Prompt**: "You are a helpful assistant that outputs only JSON."

### 1.2 问题分析
1.  **角色冗余**: System Prompt ("helpful assistant") 与 User Prompt ("crypto analyst") 角色定义冲突且重复。
2.  **指令松散**: "Translate the news" 和 "summarize it" 指令不够具体，可能导致模型生成过长的摘要。
3.  **JSON Schema 定义不严**: 虽然示例给出了 JSON 结构，但没有明确字段约束（如长度限制），容易导致生成废话。
4.  **Token 浪费**: "Input Headline:" 等前缀词略显冗余。
5.  **输出格式**: 虽然使用了 `response_format: json_object`，但提示词中的 JSON 示例占据了较多 Token。

---

## 2. 优化建议 (Recommendations)

### 2.1 优化策略
1.  **System Prompt 专门化**: 将角色定义和核心任务移至 System Prompt，User Prompt 仅提供数据。
2.  **精简 User Prompt**: 移除所有非数据内容，仅保留 JSON 结构的输入数据。
3.  **明确约束**: 限制摘要字数（例如 "Under 30 words"），强制使用简练语言。
4.  **移除示例**: 对于简单的 JSON 结构，如果使用了 `json_object` 模式，只需在 System Prompt 中描述字段即可，无需提供完整 JSON 样例，或者提供更紧凑的样例。

### 2.2 推荐 Prompt 设计

**System Message**:
```text
Role: Financial Analyst.
Task: Translate news to {TargetLang}, summarize briefly (<30 words), and analyze sentiment.
Output JSON: {"title": "translated_title", "summary": "concise_summary", "sentiment": "POSITIVE/NEGATIVE/NEUTRAL"}
```

**User Message**:
```text
Headline: {Headline}
Summary: {Summary}
```

### 2.3 预期收益
- **Input Token 减少**: 约 30-40%（移除了重复指令和冗长的 JSON 示例格式）。
- **Output Token 减少**: 通过限制摘要长度，不仅省钱，还提升了用户阅读体验（Telegram 消息更紧凑）。
- **稳定性**: 将指令固化在 System Prompt 中通常比混合在 User Prompt 中更稳定。

---

## 3. 代码修改建议 (Code Changes)

建议修改 `Process` 方法中的 Prompt 构建逻辑：

```go
// 优化后的 Process 方法逻辑
func (p *DeepSeekProcessor) Process(article *Article) error {
    // 构造 System Prompt (静态部分，Token 消耗固定且高效)
    systemPrompt := fmt.Sprintf(
        `Role: Financial Analyst. Task: Translate news to %s, summarize (<30 words), analyze sentiment. Output JSON: {"title":"...","summary":"...","sentiment":"POSITIVE/NEGATIVE/NEUTRAL"}`,
        p.targetLang,
    )

    // 构造 User Prompt (仅包含动态数据)
    userPrompt := fmt.Sprintf("Headline: %s\nSummary: %s", article.Headline, article.Summary)

    reqBody := map[string]interface{}{
        "model": "deepseek-chat",
        "messages": []map[string]string{
            {"role": "system", "content": systemPrompt},
            {"role": "user", "content": userPrompt},
        },
        "response_format": map[string]string{"type": "json_object"},
        "temperature": 0.3,
        "max_tokens": 100, // 限制最大输出 Token，防止模型发疯
    }
    // ... 发送请求逻辑保持不变
}
```

## 4. 结论

当前 Prompt 虽然可用，但在大规模运行时存在 Token 浪费和输出不可控的风险。实施上述优化方案后，预计每个请求可节省约 50-80 tokens，且响应速度会有所提升。

**建议**: 立即采纳优化方案。
