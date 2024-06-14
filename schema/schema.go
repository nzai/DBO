package schema

import (
	"regexp"

	"github.com/gobeam/stringy"
)

type Table struct {
	Name           string
	NomarlizedName string
	SingularName   string
	IsPlural       bool
	Columns        []*Column
	PrimaryKeys    []string
}

type Column struct {
	Name            string
	NomarlizedName  string
	Type            ColumnType
	Len             int
	GoType          string
	NullableType    string
	IsPrimary       bool
	IsNotNull       bool
	IsUnique        bool
	IsBinary        bool
	IsAutoIncrement bool
	IsID            bool
	IsPlural        bool
	Comment         string
	Enums           map[string]string
}

type ColumnType string

func (t ColumnType) String() string {
	return string(t)
}

const (
	TypeUnknown     ColumnType = ""
	TypeUnspecified ColumnType = "Unspecified"
	TypeTinyInt     ColumnType = "Tiny"
	TypeSmallInt    ColumnType = "Small"
	TypeBigInt      ColumnType = "Big"
	TypeFloat       ColumnType = "Float"
	TypeDouble      ColumnType = "Double"
	TypeNull        ColumnType = "Null"
	TypeTimestamp   ColumnType = "Timestamp"
	TypeLonglong    ColumnType = "Longlong"
	TypeInt24       ColumnType = "Int24"
	TypeDate        ColumnType = "Date"
	TypeDuration    ColumnType = "Duration"
	TypeDatetime    ColumnType = "Datetime"
	TypeYear        ColumnType = "Year"
	TypeNewDate     ColumnType = "NewDate"
	TypeVarchar     ColumnType = "Varchar"
	TypeBit         ColumnType = "Bit"
	TypeJSON        ColumnType = "JSON"
	TypeNewDecimal  ColumnType = "NewDecimal"
	TypeEnum        ColumnType = "Enum"
	TypeSet         ColumnType = "Set"
	TypeTinyBlob    ColumnType = "TinyBlob"
	TypeMediumBlob  ColumnType = "MediumBlob"
	TypeLongBlob    ColumnType = "LongBlob"
	TypeBlob        ColumnType = "Blob"
	TypeVarString   ColumnType = "VarString"
	TypeString      ColumnType = "String"
	TypeGeometry    ColumnType = "Geometry"
)

type Parser interface {
	ParseCreateTable(ddl string) ([]*Table, error)
}

func GetParser() Parser {
	return NewTidbParser()
}

var idRegex = regexp.MustCompile(`(id|Id|ID)$`)

func Normalize(v string) string {
	v = stringy.New(v).CamelCase()
	return idRegex.ReplaceAllString(v, "ID")
}
