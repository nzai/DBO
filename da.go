package dbo

import (
	"context"
)

// DataAccesser data access contract
type DataAccesser interface {
	Inserter
	Updater
	Saver
	Geter
	Querier
}

type Inserter interface {
	Insert(context.Context, interface{}) (interface{}, error)
	InsertTx(context.Context, *DBContext, interface{}) (interface{}, error)
	InsertInBatches(context.Context, interface{}, int) (interface{}, error)
	InsertInBatchesTx(context.Context, *DBContext, interface{}, int) (interface{}, error)
}

type Updater interface {
	Update(context.Context, interface{}) (int64, error)
	UpdateTx(context.Context, *DBContext, interface{}) (int64, error)
}

type Saver interface {
	Save(context.Context, interface{}) error
	SaveTx(context.Context, *DBContext, interface{}) error
}

type Geter interface {
	Get(context.Context, interface{}, interface{}) error
	GetTx(context.Context, *DBContext, interface{}, interface{}) error
}

type Querier interface {
	Query(context.Context, Conditions, interface{}) error
	QueryTx(context.Context, *DBContext, Conditions, interface{}) error
	Count(context.Context, Conditions, interface{}) (int, error)
	CountTx(context.Context, *DBContext, Conditions, interface{}) (int, error)
	Page(context.Context, Conditions, interface{}) (int, error)
	PageTx(context.Context, *DBContext, Conditions, interface{}) (int, error)
	QueryRawSQL(context.Context, interface{}, string, ...interface{}) error
	QueryRawSQLTx(context.Context, *DBContext, interface{}, string, ...interface{}) error
}

type Conditions interface {
	GetConditions() ([]string, []interface{})
	GetPager() *Pager
	GetOrderBy() string
}
