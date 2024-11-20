package dbo

import (
	"context"
	"reflect"
	"testing"
)

type tableA struct {
	ID     int64  `gorm:"id" json:"id"`         //  id
	Name   string `gorm:"name" json:"name"`     // 名称
	Remark string `gorm:"remark" json:"remark"` // 备注
}

func (tableA) TableName() string {
	return "table_a"
}

func TestGet(t *testing.T) {
	ctx := context.Background()

	got1, err := Get[*tableA](ctx, 9)
	if err != nil {
		t.Errorf("failed to get table_a due to %v", err)
	}

	want1 := &tableA{ID: 9, Name: "Unknown"}
	if !reflect.DeepEqual(got1, want1) {
		t.Errorf("Get() = %v, want %v", got1, want1)
	}

	got2, err := Get[tableA](ctx, 9)
	if err != nil {
		t.Errorf("failed to get table_a due to %v", err)
	}

	if !reflect.DeepEqual(&got2, want1) {
		t.Errorf("Get() = %v, want %v", got1, want1)
	}
}
