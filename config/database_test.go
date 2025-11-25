package config

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseTimestamps 测试数据库时间戳字段的正确处理
// 根据数据库最佳实践文档，所有INSERT操作不应手动指定时间戳字段
// 所有UPDATE操作应更新updated_at字段（通过触发器或显式设置）

func TestDatabaseTimestamps(t *testing.T) {
	// 创建临时数据库
	tempDir, err := os.MkdirTemp("", "nofx-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := NewDatabase(dbPath)
	require.NoError(t, err)
	defer db.Close()

	userID := "test_user"

	t.Run("TestAIModelTimestamps", testAIModelTimestamps(db, userID))
	t.Run("TestExchangeTimestamps", testExchangeTimestamps(db, userID))
	t.Run("TestTraderTimestamps", testTraderTimestamps(db, userID))
	t.Run("TestUserSignalSourceTimestamps", testUserSignalSourceTimestamps(db, userID))
	t.Run("TestSystemConfigTimestamps", testSystemConfigTimestamps(db))
}

func testAIModelTimestamps(db *Database, userID string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// 测试创建AI模型配置
		t.Run("CreateAIModel", func(t *testing.T) {
			// 创建新记录
			err := db.UpdateAIModel(userID, "deepseek", true, "api_key_123", "https://custom.api", "custom_model")
			assert.NoError(t, err)

			// 验证记录存在且时间戳已自动设置
			models, err := db.GetAIModels(userID)
			assert.NoError(t, err)
			assert.Len(t, models, 1)

			model := models[0]
			assert.Equal(t, "deepseek", model.ID)
			assert.Equal(t, userID, model.UserID)
			assert.True(t, model.Enabled)
			assert.Equal(t, "api_key_123", model.APIKey)
			assert.NotZero(t, model.CreatedAt)
			assert.NotZero(t, model.UpdatedAt)
			assert.Equal(t, model.CreatedAt, model.UpdatedAt)

			createdAt := model.CreatedAt

			// 等待一小段时间确保时间戳不同
			time.Sleep(10 * time.Millisecond)

			// 更新现有记录
			err = db.UpdateAIModel(userID, "deepseek", false, "api_key_456", "", "")
			assert.NoError(t, err)

			// 验证updated_at已更新
			models, err = db.GetAIModels(userID)
			assert.NoError(t, err)
			assert.Len(t, models, 1)

			model = models[0]
			assert.False(t, model.Enabled)
			assert.Equal(t, "api_key_456", model.APIKey)
			assert.True(t, model.UpdatedAt.After(createdAt))

			t.Logf("✅ AI模型时间戳测试通过: CreatedAt=%v, UpdatedAt=%v", model.CreatedAt, model.UpdatedAt)
		})

		// 测试旧版兼容逻辑
		t.Run("BackwardCompatibility", func(t *testing.T) {
			// 使用provider而不是完整ID（测试旧版兼容）
			err := db.UpdateAIModel("default", "qwen", true, "qwen_key", "", "")
			assert.NoError(t, err)

			models, err := db.GetAIModels("default")
			assert.NoError(t, err)

			// 可能有多个模型，只要找到qwen相关即可
			var found bool
			for _, model := range models {
				if model.Provider == "qwen" {
					found = true
					assert.NotZero(t, model.CreatedAt)
					assert.NotZero(t, model.UpdatedAt)
					t.Logf("✅ 旧版兼容逻辑测试通过: Provider=%s, CreatedAt=%v", model.Provider, model.CreatedAt)
					break
				}
			}
			assert.True(t, found, "应该找到qwen模型配置")
		})
	}
}

func testExchangeTimestamps(db *Database, userID string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		t.Run("CreateExchange", func(t *testing.T) {
			// 创建新交易所配置
			err := db.UpdateExchange(userID, "binance", true, "binance_api", "binance_secret", false, "", "", "", "", "")
			assert.NoError(t, err)

			// 验证记录存在且时间戳已自动设置
			exchanges, err := db.GetExchanges(userID)
			assert.NoError(t, err)
			assert.Len(t, exchanges, 1)

			exchange := exchanges[0]
			assert.Equal(t, "binance", exchange.ID)
			assert.Equal(t, userID, exchange.UserID)
			assert.True(t, exchange.Enabled)
			assert.Equal(t, "binance_api", exchange.APIKey)
			assert.Equal(t, "binance_secret", exchange.SecretKey)
			assert.False(t, exchange.Testnet)
			assert.NotZero(t, exchange.CreatedAt)
			assert.NotZero(t, exchange.UpdatedAt)
			assert.Equal(t, exchange.CreatedAt, exchange.UpdatedAt)

			createdAt := exchange.CreatedAt

			// 等待一小段时间确保时间戳不同
			time.Sleep(10 * time.Millisecond)

			// 更新现有记录
			err = db.UpdateExchange(userID, "binance", false, "new_api", "new_secret", true, "", "", "", "", "")
			assert.NoError(t, err)

			// 验证updated_at已更新
			exchanges, err = db.GetExchanges(userID)
			assert.NoError(t, err)
			assert.Len(t, exchanges, 1)

			exchange = exchanges[0]
			assert.False(t, exchange.Enabled)
			assert.Equal(t, "new_api", exchange.APIKey)
			assert.Equal(t, "new_secret", exchange.SecretKey)
			assert.True(t, exchange.Testnet)
			assert.True(t, exchange.UpdatedAt.After(createdAt))

			t.Logf("✅ 交易所时间戳测试通过: CreatedAt=%v, UpdatedAt=%v", exchange.CreatedAt, exchange.UpdatedAt)
		})

		// 测试多个交易所
		t.Run("MultipleExchanges", func(t *testing.T) {
			// 创建另一个交易所
			err := db.UpdateExchange(userID, "hyperliquid", true, "hl_api", "hl_secret", false, "wallet_addr", "", "", "", "")
			assert.NoError(t, err)

			// 验证两个交易所都存在
			exchanges, err := db.GetExchanges(userID)
			assert.NoError(t, err)
			assert.Len(t, exchanges, 2)

			for _, exchange := range exchanges {
				assert.NotZero(t, exchange.CreatedAt)
				assert.NotZero(t, exchange.UpdatedAt)
				t.Logf("  - %s: CreatedAt=%v", exchange.ID, exchange.CreatedAt)
			}

			t.Logf("✅ 多交易所时间戳测试通过")
		})
	}
}

func testTraderTimestamps(db *Database, userID string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		t.Run("CreateTrader", func(t *testing.T) {
			// 先创建必要的依赖
			err := db.UpdateAIModel(userID, "deepseek", true, "api_key", "", "")
			assert.NoError(t, err)

			err = db.UpdateExchange(userID, "binance", true, "api_key", "secret", false, "", "", "", "", "")
			assert.NoError(t, err)

			// 创建交易员
			traderID := GenerateUUID()
			trader := &TraderRecord{
				ID:             traderID,
				UserID:         userID,
				Name:           "TestTrader",
				AIModelID:      "deepseek",
				ExchangeID:     "binance",
				InitialBalance: 10000.0,
			}

			err = db.CreateTrader(trader)
			assert.NoError(t, err)

			// 验证交易员记录
			traders, err := db.GetTraders(userID)
			assert.NoError(t, err)
			assert.Len(t, traders, 1)

			assert.Equal(t, traderID, traders[0].ID)
			assert.Equal(t, userID, traders[0].UserID)
			assert.NotZero(t, traders[0].CreatedAt)
			assert.NotZero(t, traders[0].UpdatedAt)

			t.Logf("✅ 交易员时间戳测试通过: CreatedAt=%v, UpdatedAt=%v", traders[0].CreatedAt, traders[0].UpdatedAt)
		})
	}
}

func testUserSignalSourceTimestamps(db *Database, userID string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		t.Run("CreateUserSignalSource", func(t *testing.T) {
			// 创建用户信号源
			err := db.CreateUserSignalSource(userID, "https://coin-pool.api", "https://oi-top.api")
			assert.NoError(t, err)

			// 验证记录
			source, err := db.GetUserSignalSource(userID)
			assert.NoError(t, err)
			assert.Equal(t, userID, source.UserID)
			assert.Equal(t, "https://coin-pool.api", source.CoinPoolURL)
			assert.Equal(t, "https://oi-top.api", source.OITopURL)
			assert.NotZero(t, source.CreatedAt)
			assert.NotZero(t, source.UpdatedAt)

			createdAt := source.CreatedAt

			// 等待一小段时间
			time.Sleep(10 * time.Millisecond)

			// 更新信号源
			err = db.UpdateUserSignalSource(userID, "https://new-pool.api", "https://new-oi.api")
			assert.NoError(t, err)

			// 验证updated_at已更新
			source, err = db.GetUserSignalSource(userID)
			assert.NoError(t, err)
			assert.Equal(t, "https://new-pool.api", source.CoinPoolURL)
			assert.Equal(t, "https://new-oi.api", source.OITopURL)
			assert.True(t, source.UpdatedAt.After(createdAt))

			t.Logf("✅ 用户信号源时间戳测试通过: CreatedAt=%v, UpdatedAt=%v", source.CreatedAt, source.UpdatedAt)
		})
	}
}

func testSystemConfigTimestamps(db *Database) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		t.Run("CreateSystemConfig", func(t *testing.T) {
			testKey := "test_config_key"
			testValue := "test_value"

			// 创建设置
			err := db.SetSystemConfig(testKey, testValue)
			assert.NoError(t, err)

			// 验证可以获取
			value, err := db.GetSystemConfig(testKey)
			assert.NoError(t, err)
			assert.Equal(t, testValue, value)

			// 注意：SystemConfig没有时间戳字段，所以不测试时间戳
			t.Logf("✅ 系统配置测试通过")
		})
	}
}

// BenchmarkDatabaseTimestamps 性能基准测试
func BenchmarkDatabaseTimestamps(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "nofx-benchmark-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "benchmark.db")
	db, err := NewDatabase(dbPath)
	require.NoError(b, err)
	defer db.Close()

	userID := "benchmark_user"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 创建AI模型
		err := db.UpdateAIModel(userID, "deepseek", true, "api_key", "", "")
		if err != nil {
			b.Fatal(err)
		}

		// 更新AI模型
		err = db.UpdateAIModel(userID, "deepseek", false, "new_api", "", "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestTimestampAutoCreation 测试时间戳自动创建
// 根据最佳实践文档，INSERT语句不应手动指定时间戳
func TestTimestampAutoCreation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "nofx-timestamp-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "timestamp_test.db")
	db, err := NewDatabase(dbPath)
	require.NoError(t, err)
	defer db.Close()

	t.Run("VerifyNoManualTimestampInCode", func(t *testing.T) {
		// 这个测试验证我们的代码没有手动指定时间戳
		// 通过检查是否会产生正确的记录来验证

		// 创建AI模型（没有手动指定时间戳）
		err := db.UpdateAIModel("test", "deepseek", true, "key", "", "")
		assert.NoError(t, err)

		// 验证数据库自动设置了时间戳
		models, err := db.GetAIModels("test")
		assert.NoError(t, err)
		assert.Len(t, models, 1)

		// 验证时间戳是合理的（不是零值）
		assert.False(t, models[0].CreatedAt.IsZero(), "CreatedAt应该不为零")
		assert.False(t, models[0].UpdatedAt.IsZero(), "UpdatedAt应该不为零")

		// 验证时间戳格式正确
		// SQLite: 格式类似 "2025-11-25 23:00:00"
		// PostgreSQL: 格式类似 "2025-11-25T23:00:00Z"
		assert.Contains(t, models[0].CreatedAt.Format("2006"), "2025")
		assert.Contains(t, models[0].UpdatedAt.Format("2006"), "2025")

		t.Logf("✅ 验证自动时间戳创建: %v, %v", models[0].CreatedAt, models[0].UpdatedAt)
	})
}

// TestTriggerTimestamps 测试触发器是否正确工作
func TestTriggerTimestamps(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "nofx-trigger-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "trigger_test.db")
	db, err := NewDatabase(dbPath)
	require.NoError(t, err)
	defer db.Close()

	t.Run("VerifyUpdateTrigger", func(t *testing.T) {
		// 创建记录
		err := db.UpdateAIModel("test", "deepseek", true, "key", "", "")
		assert.NoError(t, err)

		// 获取初始时间戳
		models, err := db.GetAIModels("test")
		assert.NoError(t, err)
		initialCreated := models[0].CreatedAt
		initialUpdated := models[0].UpdatedAt

		// 等待一段时间
		time.Sleep(100 * time.Millisecond)

		// 更新记录
		err = db.UpdateAIModel("test", "deepseek", false, "new_key", "", "")
		assert.NoError(t, err)

		// 验证触发器更新了updated_at
		models, err = db.GetAIModels("test")
		assert.NoError(t, err)

		// 验证created_at保持不变，updated_at被更新
		assert.Equal(t, initialCreated, models[0].CreatedAt, "CreatedAt不应该改变")
		assert.True(t, models[0].UpdatedAt.After(initialUpdated), "UpdatedAt应该被触发器更新")

		t.Logf("✅ 验证触发器工作: CreatedAt=%v, UpdatedAt=%v (after update)", models[0].CreatedAt, models[0].UpdatedAt)
	})
}

// TestConcurrentTimestamps 测试并发插入时的时间戳行为
func TestConcurrentTimestamps(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "nofx-concurrent-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "concurrent_test.db")
	db, err := NewDatabase(dbPath)
	require.NoError(t, err)
	defer db.Close()

	t.Run("ConcurrentInserts", func(t *testing.T) {
		numGoroutines := 10
		done := make(chan bool, numGoroutines)

		// 并发创建多个用户
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				userID := "concurrent_user_" + string(rune('0'+id%10))
				err := db.UpdateAIModel(userID, "deepseek", true, "api_key", "", "")
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// 验证所有记录都被创建
		// 注意：由于每个goroutine使用不同的userID，这里不能简单地检查总数量
		// 我们只验证没有错误产生
		t.Logf("✅ 并发插入测试通过")
	})
}

// Helper function to check if timestamp is reasonable (not zero and in expected year)
func isReasonableTimestamp(t time.Time) bool {
	if t.IsZero() {
		return false
	}
	year := t.Year()
	return year >= 2020 && year <= 2030
}
