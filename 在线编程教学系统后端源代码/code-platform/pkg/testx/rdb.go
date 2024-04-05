package testx

import (
	"context"
	"fmt"
	"strings"

	"code-platform/storage"

	"github.com/jmoiron/sqlx"
)

func MustTruncateTable(ctx context.Context, rdbClient storage.RDBClient, tableNames ...string) {
	// 防止 Truncate 误操作
	sqlStr := `SELECT DATABASE()`
	var database string
	if err := sqlx.GetContext(ctx, rdbClient, &database, sqlStr); err != nil {
		panic(err)
	}
	if !strings.Contains(database, "test") {
		panic("Shouldn't truncate in normal database")
	}
	for _, tableName := range tableNames {
		if _, err := rdbClient.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s", tableName)); err != nil {
			panic(err)
		}
	}
}
