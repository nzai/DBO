package dbo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nzai/log"
)

// GetTrans begin a transaction
func GetTrans(ctx context.Context, fn func(ctx context.Context, tx *DBContext) error) error {
	log.Debug(ctx, "begin transaction")

	dbo, err := GetGlobal()
	if err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, dbo.config.TransactionTimeout)
	defer cancel()

	db, err := GetDB(ctxWithTimeout)
	if err != nil {
		return err
	}

	//db.DB = db.BeginTx(ctxWithTimeout, &sql.TxOptions{})
	db.DB = db.Begin(&sql.TxOptions{})
	funcDone := make(chan error)
	go func() {
		defer func() {
			if err1 := recover(); err1 != nil {
				log.Warn(ctxWithTimeout, "transaction panic", log.Any("recover error", err1))
				funcDone <- fmt.Errorf("transaction panic: %+v", err1)
			}
		}()

		// call func
		funcDone <- fn(ctxWithTimeout, db)
	}()

	select {
	case err = <-funcDone:
		log.Debug(ctxWithTimeout, "transaction fn done")
	case <-ctxWithTimeout.Done():
		// context deadline exceeded
		err = ctxWithTimeout.Err()
		log.Warn(ctxWithTimeout, "transaction context deadline exceeded", log.Err(err))
	}

	if err != nil {
		err1 := db.Rollback().Error
		if err1 != nil {
			log.Warn(ctxWithTimeout, "rollback transaction failed", log.String("outer error", err.Error()), log.Err(err1))
		} else {
			log.Debug(ctxWithTimeout, "rollback transaction successfully")
		}
		return err
	}

	err = db.Commit().Error
	if err != nil {
		log.Warn(ctxWithTimeout, "commit transaction failed", log.Err(err))
		return err
	}

	log.Debug(ctxWithTimeout, "commit transaction successfully")

	return nil
}

type transactionResult struct {
	Result interface{}
	Error  error
}

// GetTransResult begin a transaction, get result of callback
func GetTransResult(ctx context.Context, fn func(ctx context.Context, tx *DBContext) (interface{}, error)) (interface{}, error) {
	log.Debug(ctx, "begin transaction")

	dbo, err := GetGlobal()
	if err != nil {
		return nil, err
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, dbo.config.TransactionTimeout)
	defer cancel()

	db, err := GetDB(ctxWithTimeout)
	if err != nil {
		return nil, err
	}

	db.DB = db.Begin(&sql.TxOptions{})

	funcDone := make(chan *transactionResult)
	go func() {
		defer func() {
			if err1 := recover(); err1 != nil {
				log.Warn(ctxWithTimeout, "transaction panic", log.Any("recover error", err1))
				funcDone <- &transactionResult{Error: fmt.Errorf("transaction panic: %+v", err1)}
			}
		}()

		// call func
		result, err := fn(ctxWithTimeout, db)
		funcDone <- &transactionResult{Result: result, Error: err}
	}()

	var funcResult *transactionResult
	select {
	case funcResult = <-funcDone:
		log.Debug(ctxWithTimeout, "transaction fn done")
	case <-ctxWithTimeout.Done():
		// context deadline exceeded
		funcResult = &transactionResult{Error: ctxWithTimeout.Err()}
		log.Warn(ctxWithTimeout, "transaction context deadline exceeded", log.Err(ctxWithTimeout.Err()))
	}

	if funcResult.Error != nil {
		log.Warn(ctxWithTimeout, "transaction failed", log.Err(funcResult.Error))

		err1 := db.Rollback().Error
		if err1 != nil {
			log.Warn(ctxWithTimeout, "rollback transaction failed",
				log.String("transaction error", funcResult.Error.Error()),
				log.Err(err1))
		} else {
			log.Debug(ctxWithTimeout, "rollback transaction successfully", log.Err(funcResult.Error))

		}
		return nil, funcResult.Error
	}

	err = db.Commit().Error
	if err != nil {
		log.Warn(ctxWithTimeout, "commit transaction failed", log.Err(err))
		return nil, err
	}

	log.Debug(ctxWithTimeout, "commit transaction successfully")

	return funcResult.Result, nil
}
