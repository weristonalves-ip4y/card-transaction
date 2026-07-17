package sqlserver

import (
	"context"
	"database/sql"
)

type queryExecutor interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
