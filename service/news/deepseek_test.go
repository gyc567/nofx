package news

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeepSeekProcessor_Process(t *testing.T) {
	// Mock DeepSeek API Server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 验证请求方法和头信息
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		// 2. 解析请求体
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)

		// 3. 验证 Prompt 结构 (Token 优化验证)
		messages := reqBody["messages"].([]interface{})
		assert.Equal(t, 2, len(messages))

		systemMsg := messages[0].(map[string]interface{})
		userMsg := messages[1].(map[string]interface{})

		assert.Equal(t, "system", systemMsg["role"])
		assert.Contains(t, systemMsg["content"], "Role: Financial Analyst")
		assert.Contains(t, systemMsg["content"], "Translate news to zh-CN")

		assert.Equal(t, "user", userMsg["role"])
		assert.Contains(t, userMsg["content"], "Headline: Test News")
		assert.Contains(t, userMsg["content"], "Summary: This is a test.")
		
		// 4. 验证 max_tokens
		assert.Equal(t, float64(200), reqBody["max_tokens"])

		// 5. 模拟返回响应
		response := AIResponse{}
		response.Choices = make([]struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}, 1)
		
		mockOutput := AIOutput{
			Title:     "测试新闻",
			Summary:   "这是一个测试摘要",
			Sentiment: "NEUTRAL",
		}
		outputJSON, _ := json.Marshal(mockOutput)
		
		response.Choices[0].Message.Content = string(outputJSON)

		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Initialize Processor with Mock Server URL
	processor := NewDeepSeekProcessor("test-key", mockServer.URL, "zh-CN")

	article := &Article{
		Headline: "Test News",
		Summary:  "This is a test.",
	}

	// Execute
	err := processor.Process(article)

	// Verify
	assert.NoError(t, err)
	assert.True(t, article.AIProcessed)
	assert.Equal(t, "测试新闻", article.TranslatedHeadline)
	assert.Equal(t, "这是一个测试摘要", article.TranslatedSummary)
	assert.Equal(t, "NEUTRAL", article.Sentiment)
}

func TestDeepSeekProcessor_Process_CleanJSON(t *testing.T) {
	// 测试 Markdown 清洗功能
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := AIResponse{}
		response.Choices = make([]struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}, 1)
		
		// 模拟带 Markdown 标记的返回
		response.Choices[0].Message.Content = "```json\n{\"title\": \"Cleaned\", \"summary\": \"Done\", \"sentiment\": \"POSITIVE\"}\n```"

		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	processor := NewDeepSeekProcessor("k", mockServer.URL, "en")
	article := &Article{Headline: "H", Summary: "S"}

	err := processor.Process(article)
	assert.NoError(t, err)
	assert.Equal(t, "Cleaned", article.TranslatedHeadline)
}
