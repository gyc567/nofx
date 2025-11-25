package database

import (
	"testing"
)

func TestMigrateData(t *testing.T) {
	// 设置测试配置
	sqlitePath := "test.db"
	neonDSN := "host=your-neon-host port=5432 user=your-user dbname=your-db password=your-password sslmode=require"
	
	// 测试迁移
	err := MigrateData(sqlitePath, neonDSN)
	if err != nil {
		t.Fatal(err)
	}
}
