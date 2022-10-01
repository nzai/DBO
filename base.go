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

type BaseDA struct{}

func (s BaseDA) Insert(ctx context.Context, value interface{}) (interface{}, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return nil, err
	}

	return s.InsertTx(ctx, db, value)
}

func (s BaseDA) InsertTx(ctx context.Context, db *DBContext, value interface{}) (interface{}, error) {
	start := time.Now()
	err := db.ResetCondition().Create(value).Error
	if err != nil {
		me, ok := err.(*mysql.MySQLError)
		if ok && me.Number == 1062 {
			log.Warn(ctx, "insert duplicate record",
				log.Err(me),
				log.String("tableName", db.GetTableName(value)),
				log.Any("value", value),
				log.Duration("duration", time.Since(start)))
			return 0, ErrDuplicateRecord
		}

		log.Warn(ctx, "insert failed",
			log.Err(err),
			log.String("tableName", db.GetTableName(value)),
			log.Any("value", value),
			log.Duration("duration", time.Since(start)))
		return nil, err
	}

	log.Debug(ctx, "insert successfully",
		log.String("tableName", db.GetTableName(value)),
		log.Any("value", value),
		log.Duration("duration", time.Since(start)))

	return value, nil
}

// InsertInBatches Insert records in batch. visit https://gorm.io/docs/create.html for detail
func (s BaseDA) InsertInBatches(ctx context.Context, value interface{}, batchSize int) (interface{}, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return nil, err
	}

	return s.InsertInBatchesTx(ctx, db, value, batchSize)
}

// InsertInBatchesTx Insert records in batch with context. visit https://gorm.io/docs/create.html for detail
func (s BaseDA) InsertInBatchesTx(ctx context.Context, db *DBContext, value interface{}, batchSize int) (interface{}, error) {
	start := time.Now()
	err := db.ResetCondition().CreateInBatches(value, batchSize).Error
	if err != nil {
		me, ok := err.(*mysql.MySQLError)
		if ok && me.Number == 1062 {
			log.Warn(ctx, "insertBatches duplicate record",
				log.Err(me),
				log.String("tableName", db.GetTableName(value)),
				log.Any("value", value),
				log.Duration("duration", time.Since(start)))
			return 0, ErrDuplicateRecord
		}

		log.Warn(ctx, "insertBatches failed",
			log.Err(err),
			log.String("tableName", db.GetTableName(value)),
			log.Any("value", value),
			log.Duration("duration", time.Since(start)))
		return nil, err
	}

	log.Debug(ctx, "insertBatches successfully",
		log.String("tableName", db.GetTableName(value)),
		log.Any("value", value),
		log.Duration("duration", time.Since(start)))

	return value, nil
}

func (s BaseDA) Update(ctx context.Context, value interface{}) (int64, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return 0, err
	}

	return s.UpdateTx(ctx, db, value)
}

func (s BaseDA) UpdateTx(ctx context.Context, db *DBContext, value interface{}) (int64, error) {
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

func (s BaseDA) Save(ctx context.Context, value interface{}) error {
	db, err := GetDB(ctx)
	if err != nil {
		return err
	}

	return s.SaveTx(ctx, db, value)
}

func (s BaseDA) SaveTx(ctx context.Context, db *DBContext, value interface{}) error {
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

func (s BaseDA) Get(ctx context.Context, id interface{}, value interface{}) error {
	db, err := GetDB(ctx)
	if err != nil {
		return err
	}

	return s.GetTx(ctx, db, id, value)
}

func (s BaseDA) GetTx(ctx context.Context, db *DBContext, id interface{}, value interface{}) error {
	start := time.Now()
	err := db.ResetCondition().Where("id=?", id).First(value).Error
	if err == nil {
		log.Debug(ctx, "get by id successfully",
			log.Any("id", id),
			log.String("tableName", db.GetTableName(value)),
			log.Any("value", value),
			log.Duration("duration", time.Since(start)))
		return nil
	}

	log.Warn(ctx, "get by id failed",
		log.Err(err),
		log.Any("id", id),
		log.String("tableName", db.GetTableName(value)),
		log.Any("value", value),
		log.Duration("duration", time.Since(start)))

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound
	}

	return err
}

func (s BaseDA) Query(ctx context.Context, condition Conditions, values interface{}) error {
	db, err := GetDB(ctx)
	if err != nil {
		return err
	}

	return s.QueryTx(ctx, db, condition, values)
}

func (s BaseDA) QueryTx(ctx context.Context, db *DBContext, condition Conditions, values interface{}) error {
	db.ResetCondition()

	wheres, parameters := condition.GetConditions()
	if len(wheres) > 0 {
		db.DB = db.Where(strings.Join(wheres, " and "), parameters...)
	}

	orderBy := condition.GetOrderBy()
	if orderBy != "" {
		db.DB = db.Order(orderBy)
	}

	pager := condition.GetPager()
	if pager != nil && pager.Enable() {
		// pagination
		offset, limit := pager.Offset()
		db.DB = db.Offset(offset).Limit(limit)
	}

	start := time.Now()
	err := db.Find(values).Error
	if err != nil {
		log.Warn(ctx, "query values failed",
			log.Err(err),
			log.String("tableName", db.GetTableName(values)),
			log.Any("condition", condition),
			log.Any("pager", pager),
			log.String("orderBy", orderBy),
			log.Duration("duration", time.Since(start)))
		return err
	}

	log.Debug(ctx, "query values successfully",
		log.String("tableName", db.GetTableName(values)),
		log.Any("condition", condition),
		log.Any("pager", pager),
		log.String("orderBy", orderBy),
		log.Duration("duration", time.Since(start)))

	return nil
}

func (s BaseDA) Count(ctx context.Context, condition Conditions, values interface{}) (int, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return 0, err
	}

	return s.CountTx(ctx, db, condition, values)
}

func (s BaseDA) CountTx(ctx context.Context, db *DBContext, condition Conditions, value interface{}) (int, error) {
	db.ResetCondition()

	wheres, parameters := condition.GetConditions()
	if len(wheres) > 0 {
		db.DB = db.Where(strings.Join(wheres, " and "), parameters...)
	}

	start := time.Now()
	var total int64
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

func (s BaseDA) Page(ctx context.Context, condition Conditions, values interface{}) (int, error) {
	db, err := GetDB(ctx)
	if err != nil {
		return 0, err
	}

	return s.PageTx(ctx, db, condition, values)
}

func (s BaseDA) PageTx(ctx context.Context, db *DBContext, condition Conditions, values interface{}) (int, error) {
	total, err := s.CountTx(ctx, db, condition, values)
	if err != nil {
		return 0, err
	}

	err = s.QueryTx(ctx, db, condition, values)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s BaseDA) QueryRawSQL(ctx context.Context, values interface{}, sql string, parameters ...interface{}) error {
	db, err := GetDB(ctx)
	if err != nil {
		return err
	}

	return s.QueryRawSQLTx(ctx, db, values, sql, parameters...)
}

func (s BaseDA) QueryRawSQLTx(ctx context.Context, db *DBContext, values interface{}, sql string, parameters ...interface{}) error {
	start := time.Now()
	err := db.ResetCondition().Raw(sql, parameters...).Find(values).Error
	if err != nil {
		log.Warn(ctx, "query raw sql failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("parameters", parameters),
			log.Duration("duration", time.Since(start)))
		return err
	}

	log.Debug(ctx, "query raw sql successfully",
		log.String("sql", sql),
		log.Any("parameters", parameters),
		log.Duration("duration", time.Since(start)))

	return nil
}
