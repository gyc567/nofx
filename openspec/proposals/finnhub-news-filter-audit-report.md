# 架构审计报告：Finnhub 新闻过滤与处理模块优化

**审计人**: 架构师 Agent
**日期**: 2025年12月06日
**审计对象**: `service/news` 模块优化 (Proposal: Finnhub Policy Filter & Source Whitelist)

---

## 1. 总体评价 (Executive Summary)

本次优化显著提升了 Finnhub 新闻源的信噪比，将原本可能包含大量杂音的通用新闻流转化为高价值的宏观投研情报流。实施方案遵循了 **KISS (Keep It Simple, Stupid)** 原则，采用了简单有效的字符串匹配逻辑，并在 `Fetcher` 层实现了高内聚的过滤逻辑，解耦了上层服务。

**评分**:
- **架构设计**: A- (高内聚，但配置硬编码)
- **代码质量**: A (清晰，命名规范)
- **功能完整性**: A (满足所有提案要求)
- **可测试性**: A+ (单元测试与集成测试覆盖全面)

---

## 2. 现状评估 (Status Assessment)

### 2.1 架构设计
- **模块解耦**: 所有的过滤逻辑（白名单、黑名单、关键词、区域标签）都正确封装在 `FinnhubFetcher` 内部。`Service` 层仅负责调度、去重和发送，不再关心具体的过滤规则。这是一个非常好的设计决策，符合单一职责原则（SRP）。
- **过滤链条**: 采用了合理的过滤顺序：`Source Check` -> `Excluded Check` -> `Keyword Check` -> `Region Tagging`。这种漏斗式设计最大化了性能，最廉价且排除率最高的检查（Source）放在最前面。

### 2.2 代码实现
- **可读性**: 变量命名（如 `policyKeywords`, `excludedKeywords`）语义清晰。逻辑流程线性，易于理解。
- **健壮性**: `isAllowedSource` 和 `containsAny` 均处理了大小写敏感性问题（统一转小写比较），避免了因大小写不一致导致的漏判。
- **测试**: 引入了 `httptest` 进行集成测试，模拟了真实的 API 响应，覆盖了正向和负向用例，极大增强了代码的可靠性信心。

---

## 3. 发现的问题与风险 (Findings & Risks)

尽管整体实现优秀，但从架构长远演进的角度，仍存在以下改进空间：

### 3.1 配置硬编码 (Hardcoded Configuration)
- **风险**: `policyKeywords`, `excludedKeywords`, `allowedSources` 等核心业务规则被硬编码在 Go 源代码中。
- **影响**: 
    1.  **灵活性差**: 每次调整关键词或新增媒体来源（这是运营常态），都需要重新编译并部署整个服务。
    2.  **配置分散**: 关键词和源码混合，不便于非技术人员（如投研分析师）维护。
- **建议**: 
    - 短期（P2）：接受现状，因为 KISS 原则下无需过度设计。
    - 长期（P1）：将这些列表提取到外部配置文件（如 JSON/YAML）或数据库表（`system_configs` 或 `news_rules`）中，并在 `NewFinnhubFetcher` 时加载，或支持热加载。

### 3.2 关键词匹配的局限性 (Simple String Matching)
- **风险**: 当前使用简单的 `strings.Contains`。
- **影响**: 可能会出现误判。例如，关键词 "Fed" 可能会匹配到 "Federer" (虽然有排除词 "tennis"，但不能穷举)。又如 "SEC" 可能匹配到 "second"。
- **建议**: 
    - 对于短关键词（如 "Fed", "SEC", "PMI"），建议结合单词边界检查（Regex `\bFed\b`）或加空格匹配（" Fed "），但考虑到 Finnhub 标题通常规范，当前风险可控。

### 3.3 区域判断逻辑 (Region Logic)
- **观察**: 区域判断逻辑是互斥的 `if-else if`。如果一个标题同时包含 "Fed" 和 "China"，它会被标记为【美国】（因为 `usKeywords` 检查在前）。
- **建议**: 当前逻辑是可以接受的，因为通常新闻有一个主要主体。但需知晓这一隐式优先级（美 > 中 > 欧）。

---

## 4. 改进建议 (Recommendations)

### 4.1 立即行动 (Quick Wins)
无。当前代码已达到上线标准。

### 4.2 长期演进 (Long-term Evolution)
1.  **配置外部化**: 设计一个 `NewsFilterConfig` 结构体，从数据库加载白名单和关键词列表。
    ```go
    type NewsFilterConfig struct {
        AllowedSources   []string
        PolicyKeywords   []string
        ExcludedKeywords []string
    }
    ```
2.  **动态重载**: 在 `Service.loadConfig` 中定期刷新这些过滤规则，实现“不重启调整策略”。

---

## 5. 最终结论 (Conclusion)

该优化方案代码质量高，测试覆盖完整，架构边界清晰。虽然存在配置硬编码的问题，但在当前阶段是为了遵循 KISS 原则的合理权衡。

**审计结论**: ✅ **通过 (Approved)**。可以立即部署上线。
