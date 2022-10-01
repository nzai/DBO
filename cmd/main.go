package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nzai/dbo"
	"github.com/nzai/log"
)

const batch = 20

func main() {
	ctx := context.Background()

	_dbo, err := dbo.NewWithConfig(func(c *dbo.Config) {
		c.ConnectionString = "root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	})
	if err != nil {
		log.Panic(ctx, "create dbo failed", log.Err(err))
	}
	dbo.ReplaceGlobal(_dbo)

	// tx1 := dbo.MustGetDB(ctx)
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err = delete(ctx, tx)
		if err != nil {
			return err
		}

		err = insert(ctx, tx)
		if err != nil {
			return err
		}

		err = query(ctx, tx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Panic(ctx, "create dbo failed", log.Err(err))
	}
}

func insert(ctx context.Context, tx *dbo.DBContext) error {
	items := make([]*TestTable, 0, batch)
	for index := 0; index < batch; index++ {
		items = append(items, &TestTable{
			ID:   strconv.Itoa(index),
			Name: fmt.Sprintf("TAB-%04d", index),
		})
	}

	// _, err := dbo.BaseDA{}.InsertTx(ctx, tx, items)
	_, err := dbo.BaseDA{}.InsertInBatchesTx(ctx, tx, items, 1000)
	if err != nil {
		log.Error(ctx, "insert failed", log.Err(err))
		return err
	}

	return nil
}

func delete(ctx context.Context, tx *dbo.DBContext) error {
	err := tx.Where("id <= ?", batch).Delete(&TestTable{}).Error
	if err != nil {
		log.Error(ctx, "delete failed", log.Err(err))
		return err
	}

	return nil
}

func query(ctx context.Context, tx *dbo.DBContext) error {
	items := []*TestTable{}
	err := dbo.BaseDA{}.QueryRawSQLTx(ctx, tx, &items, "select 'aaa' id, name from test_table where id in (?) order by id desc", []string{"1", "2"})
	if err != nil {
		log.Error(ctx, "delete failed", log.Err(err))
		return err
	}

	log.Info(ctx, "query success", log.Any("items", items))
	return nil
}

type TestTable struct {
	ID   string `gorm:"column:id;primary_key"`
	Name string `gorm:"column:name"`
}

func (TestTable) TableName() string {
	return "test_table"
}
