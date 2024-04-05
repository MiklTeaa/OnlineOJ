package transactionx

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"

	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/storage"
)

func DoTransaction(
	ctx context.Context,
	storage *storage.Storage,
	logger *log.Logger,
	f func(context.Context, storage.RDBClient) error,
	opts *sql.TxOptions,
) (err error) {
	tx, err := storage.RDBTransaction(ctx, opts)
	if err != nil {
		logger.Error(err, "start transaction failed")
		return errorx.InternalErr(err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = fmt.Errorf("%v\n%s", r, string(debug.Stack()))
			return
		}
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	return f(ctx, tx)
}
