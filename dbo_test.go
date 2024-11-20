package dbo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nzai/log"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	dboHandler, err := NewWithConfig(func(c *Config) {
		c.ConnectionString = "root:123456@tcp(127.0.0.1:3306)/testdb?parseTime=true&charset=utf8mb4"
		c.DBType = MySQL
		c.TransactionTimeout = time.Minute * 10
	})
	if err != nil {
		log.Panic(ctx, "create dbo failed", log.Err(err))
	}
	ReplaceGlobal(dboHandler)

	os.Exit(m.Run())
}
