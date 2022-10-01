package dbo

import (
	// mysql driver
	"context"
	"sync"

	"github.com/nzai/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	globalDBO   *DBO
	globalMutex sync.Mutex
)

// DBO database operator
type DBO struct {
	db     *gorm.DB
	config *Config
}

// MustGetDB get db context otherwise panic
func MustGetDB(ctx context.Context) *DBContext {
	dbContext, err := GetDB(ctx)
	if err != nil {
		log.Panic(ctx, "get db context failed", log.Err(err))
	}

	return dbContext
}

// GetDB get db context
func GetDB(ctx context.Context) (*DBContext, error) {
	dbo, err := GetGlobal()
	if err != nil {
		return nil, err
	}

	return dbo.GetDB(ctx), nil
}

// ReplaceGlobal replace global dbo instance
func ReplaceGlobal(dbo *DBO) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	globalDBO = dbo
}

// GetGlobal get global dbo
func GetGlobal() (*DBO, error) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	if globalDBO == nil {
		dbo, err := New()
		if err != nil {
			return nil, err
		}

		globalDBO = dbo
	}

	return globalDBO, nil
}

// New create new database operator
func New(options ...Option) (*DBO, error) {
	return NewWithConfig(options...)
}

// NewWithConfig create new database operator
func NewWithConfig(options ...Option) (*DBO, error) {
	// init config with default values
	config := getDefaultConfig()

	for _, option := range options {
		option(config)
	}

	ctx := context.Background()
	var db *gorm.DB
	var err error
	switch config.DBType {
	case MySQL:
		db, err = gorm.Open(mysql.New(mysql.Config{
			DriverName: config.DBType.DriverName(),
			DSN:        config.ConnectionString,
		}), &gorm.Config{QueryFields: true})
	default:
		log.Panic(ctx, "unsupported database type", log.String("databaseType", config.DBType.String()))
	}

	if err != nil {
		log.Warn(ctx, "init database connection failed",
			log.Err(err),
			log.String("databaseType", config.DBType.String()),
			log.String("connectionString", config.ConnectionString))
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Warn(ctx, "get DB failed",
			log.Err(err),
			log.String("databaseType", config.DBType.String()),
			log.String("connectionString", config.ConnectionString))
		return nil, err
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Warn(ctx, "ping datebase failed",
			log.Err(err),
			log.String("databaseType", config.DBType.String()),
			log.String("connectionString", config.ConnectionString))
		return nil, err
	}

	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	}

	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	}

	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	if config.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}

	return &DBO{db, config}, nil
}

func (s DBO) GetDB(ctx context.Context) *DBContext {
	ctxDB := &DBContext{DB: s.db.Session(&gorm.Session{
		Context:     ctx,
		NewDB:       true,
		QueryFields: true,
	})}

	ctxDB.Logger = logger.New(ctxDB, logger.Config{
		LogLevel:                  s.config.LogLevel.GormLogLevel(),
		SlowThreshold:             s.config.SlowThreshold,
		IgnoreRecordNotFoundError: false,
		Colorful:                  false,
	})

	return ctxDB
}
