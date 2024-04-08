package dbo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/nzai/log"
	"gorm.io/gorm"
)

func Insert[T any](ctx context.Context, value *T) (int64, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return 0, err
	}

	return InsertTx[T](ctx, db, value)
}

func InsertTx[T any](ctx context.Context, db *DBContext, value *T) (int64, error) {
	start := time.Now()
	newDB := db.ResetCondition().Create(value)
	if newDB.Error != nil {
		me, ok := newDB.Error.(*mysql.MySQLError)
		if ok && me.Number == 1062 {
			log.Warn(ctx, "insert duplicate record",
				log.Err(me),
				log.String("tableName", db.GetTableName(value)),
				log.Any("value", value),
				log.Duration("duration", time.Since(start)))
			return 0, ErrDuplicateRecord
		}

		log.Warn(ctx, "insert failed",
			log.Err(newDB.Error),
			log.String("tableName", db.GetTableName(value)),
			log.Any("value", value),
			log.Duration("duration", time.Since(start)))
		return 0, newDB.Error
	}

	log.Debug(ctx, "insert successfully",
		log.String("tableName", db.GetTableName(value)),
		log.Any("value", value),
		log.Duration("duration", time.Since(start)))

	return newDB.RowsAffected, nil
}

// InsertInBatches Insert records in batch. visit https://gorm.io/docs/create.html for detail
func InsertInBatches[T any](ctx context.Context, value []*T, batchSize int) (int64, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return 0, err
	}

	return InsertInBatchesTx[T](ctx, db, value, batchSize)
}

// InsertInBatchesTx Insert records in batch with context. visit https://gorm.io/docs/create.html for detail
func InsertInBatchesTx[T any](ctx context.Context, db *DBContext, value []*T, batchSize int) (int64, error) {
	start := time.Now()
	newDB := db.ResetCondition().CreateInBatches(value, batchSize)
	if newDB.Error != nil {
		me, ok := newDB.Error.(*mysql.MySQLError)
		if ok && me.Number == 1062 {
			log.Warn(ctx, "insertBatches duplicate record",
				log.Err(me),
				log.String("tableName", db.GetTableName(value)),
				log.Any("value", value),
				log.Duration("duration", time.Since(start)))
			return 0, ErrDuplicateRecord
		}

		log.Warn(ctx, "insertBatches failed",
			log.Err(newDB.Error),
			log.String("tableName", db.GetTableName(value)),
			log.Any("value", value),
			log.Duration("duration", time.Since(start)))
		return 0, newDB.Error
	}

	log.Debug(ctx, "insertBatches successfully",
		log.String("tableName", db.GetTableName(value)),
		log.Any("value", value),
		log.Duration("duration", time.Since(start)))

	return newDB.RowsAffected, nil
}

func Update[T any](ctx context.Context, value *T) (int64, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return 0, err
	}

	return UpdateTx[T](ctx, db, value)
}

func UpdateTx[T any](ctx context.Context, db *DBContext, value *T) (int64, error) {
	start := time.Now()
	newDB := db.ResetCondition().Save(value)
	if newDB.Error != nil {
		me, ok := newDB.Error.(*mysql.MySQLError)
		if ok && me.Number == 1062 {
			log.Warn(ctx, "update duplicate record",
				log.Err(me),
				log.String("tableName", db.GetTableName(value)),
				log.Any("value", value),
				log.Duration("duration", time.Since(start)))
			return 0, ErrDuplicateRecord
		}

		log.Warn(ctx, "update failed",
			log.Err(newDB.Error),
			log.String("tableName", db.GetTableName(value)),
			log.Any("value", value),
			log.Duration("duration", time.Since(start)))
		return 0, newDB.Error
	}

	log.Debug(ctx, "update successfully",
		log.String("tableName", db.GetTableName(value)),
		log.Any("value", value),
		log.Duration("duration", time.Since(start)))

	return newDB.RowsAffected, nil
}

func Save[T any](ctx context.Context, value *T) error {
	db, err := GetDB(ctx)
	if err != nil {
		return err
	}

	return SaveTx[T](ctx, db, value)
}

func SaveTx[T any](ctx context.Context, db *DBContext, value *T) error {
	start := time.Now()
	err := db.ResetCondition().Save(value).Error
	if err != nil {
		log.Warn(ctx, "save failed",
			log.Err(err),
			log.String("tableName", db.GetTableName(value)),
			log.Any("value", value),
			log.Duration("duration", time.Since(start)))
		return err
	}

	log.Debug(ctx, "save successfully",
		log.String("tableName", db.GetTableName(value)),
		log.Any("value", value),
		log.Duration("duration", time.Since(start)))

	return nil
}

func Get[T any](ctx context.Context, id any) (*T, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return nil, err
	}

	return GetTx[T](ctx, db, id)
}

func GetTx[T any](ctx context.Context, db *DBContext, id any) (*T, error) {
	start := time.Now()
	value := new(T)
	err := db.ResetCondition().Where("id=?", id).First(value).Error
	if err == nil {
		log.Debug(ctx, "get by id successfully",
			log.Any("id", id),
			log.String("tableName", db.GetTableName(value)),
			log.Any("value", value),
			log.Duration("duration", time.Since(start)))
		return value, nil
	}

	log.Warn(ctx, "get by id failed",
		log.Err(err),
		log.Any("id", id),
		log.String("tableName", db.GetTableName(value)),
		log.Any("value", value),
		log.Duration("duration", time.Since(start)))

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}

	return nil, err
}

func Query[T any](ctx context.Context, condition QueryCondition) ([]*T, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return nil, err
	}

	return QueryTx[T](ctx, db, condition)
}

func QueryTx[T any](ctx context.Context, db *DBContext, condition QueryCondition) ([]*T, error) {
	db.ResetCondition()

	wheres, parameters := condition.GetConditions()
	if len(wheres) > 0 {
		db.DB = db.Where(strings.Join(wheres, " and "), parameters...)
	}

	orderBy, ok := condition.(OrderByCondition)
	if ok {
		db.DB = db.Order(orderBy.GetOrderBy())
	}

	pc, ok := condition.(PagerCondition)
	if ok {
		pager := pc.GetPager()
		if pager != nil && pager.Enable() {
			// pagination
			offset, limit := pager.Offset()
			db.DB = db.Offset(offset).Limit(limit)
		}
	}

	start := time.Now()
	values := make([]*T, 0)
	err := db.Find(values).Error
	if err != nil {
		log.Warn(ctx, "query values failed",
			log.Err(err),
			log.String("tableName", db.GetTableName(values)),
			log.Any("condition", condition),
			log.Duration("duration", time.Since(start)))
		return nil, err
	}

	log.Debug(ctx, "query values successfully",
		log.String("tableName", db.GetTableName(values)),
		log.Any("condition", condition),
		log.Duration("duration", time.Since(start)))

	return values, nil
}

func Count[T any](ctx context.Context, condition QueryCondition) (int, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return 0, err
	}

	return CountTx[T](ctx, db, condition)
}

func CountTx[T any](ctx context.Context, db *DBContext, condition QueryCondition) (int, error) {
	db.ResetCondition()

	wheres, parameters := condition.GetConditions()
	if len(wheres) > 0 {
		db.DB = db.Where(strings.Join(wheres, " and "), parameters...)
	}

	start := time.Now()
	var total int64
	var value T
	tableName := db.GetTableName(value)
	err := db.Table(tableName).Count(&total).Error
	if err != nil {
		log.Warn(ctx, "count failed",
			log.Err(err),
			log.String("tableName", tableName),
			log.Any("condition", condition),
			log.Duration("duration", time.Since(start)))
		return 0, err
	}

	log.Debug(ctx, "count successfully",
		log.String("tableName", tableName),
		log.Any("condition", condition),
		log.Duration("duration", time.Since(start)))

	return int(total), nil
}

func Page[T any](ctx context.Context, condition QueryCondition) (int, []*T, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return 0, nil, err
	}

	return PageTx[T](ctx, db, condition)
}

func PageTx[T any](ctx context.Context, db *DBContext, condition QueryCondition) (int, []*T, error) {
	total, err := CountTx[T](ctx, db, condition)
	if err != nil {
		return 0, nil, err
	}

	values, err := QueryTx[T](ctx, db, condition)
	if err != nil {
		return 0, nil, err
	}

	return total, values, nil
}
