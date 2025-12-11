# Bug Proposal: AI API 402 Insufficient Balance Error Handling

## 1. Issue Description
During the decision cycle, the application fails to get an AI decision with the following error:
```
❌ 获取AI决策失败: 调用AI API失败: API返回错误 (status 402): {"error":{"message":"Insufficient Balance","type":"unknown_error","param":null,"code":"invalid_request_error"}}
```
This indicates that the configured AI provider (e.g., DeepSeek) account has run out of credits/balance. The current error handling simply propagates the raw API error, which halts the decision process abruptly without clear guidance.

## 2. Root Cause Analysis
- **Location:** `decision/engine.go` calls `mcpClient.CallWithMessages`.
- **Implementation:** `mcp/client.go` -> `callOnce` performs an HTTP POST. It checks `resp.StatusCode`. If it's not 200, it returns a generic formatted error string: `fmt.Errorf("API返回错误 (status %d): %s", resp.StatusCode, string(body))`.
- **Reason:** The external API provider returns HTTP 402 Payment Required when the quota is exhausted. The application does not handle this specific status code to provide a actionable user warning or fallback.

## 3. Proposed Fix
To improve user experience and system robustness:

1.  **Enhance `mcp/client.go`**:
    *   Define a specific `InsufficientBalanceError` type or check.
    *   In `callOnce`, explicitly check for status code `402`.
    *   Return a wrapped, clear error message indicating the API key has run out of funds.

2.  **Update `decision/engine.go`**:
    *   In `GetFullDecisionWithCustomPrompt`, detect this specific error.
    *   Log a friendly warning advising the user to check their API provider balance or switch API keys.
    *   (Future) Potentially trigger a fallback to a secondary model if configured (out of scope for immediate fix but good for design).

## 4. Implementation Plan
1.  Modify `mcp/client.go`: Add 402 handling in `callOnce`.
2.  Modify `decision/engine.go`: Catch the error and log a user-friendly message.
