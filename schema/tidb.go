package schema

import (
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/pingcap/tidb/pkg/parser/test_driver"
)

type TidbParser struct {
	columnTypeMapping   map[byte]ColumnType
	nullableTypeMapping map[string]string
	pluralize           *pluralize.Client
}

func NewTidbParser() *TidbParser {
	return &TidbParser{
		columnTypeMapping: map[byte]ColumnType{
			mysql.TypeUnspecified: TypeUnspecified,
			mysql.TypeTiny:        TypeTinyInt,
			mysql.TypeShort:       TypeSmallInt,
			mysql.TypeLong:        TypeBigInt,
			mysql.TypeFloat:       TypeFloat,
			mysql.TypeDouble:      TypeDouble,
			mysql.TypeNull:        TypeNull,
			mysql.TypeTimestamp:   TypeTimestamp,
			mysql.TypeLonglong:    TypeLonglong,
			mysql.TypeInt24:       TypeInt24,
			mysql.TypeDate:        TypeDate,
			mysql.TypeDuration:    TypeDuration,
			mysql.TypeDatetime:    TypeDatetime,
			mysql.TypeYear:        TypeYear,
			mysql.TypeNewDate:     TypeNewDate,
			mysql.TypeVarchar:     TypeVarchar,
			mysql.TypeBit:         TypeBit,
			mysql.TypeJSON:        TypeJSON,
			mysql.TypeNewDecimal:  TypeNewDecimal,
			mysql.TypeEnum:        TypeEnum,
			mysql.TypeSet:         TypeSet,
			mysql.TypeTinyBlob:    TypeTinyBlob,
			mysql.TypeMediumBlob:  TypeMediumBlob,
			mysql.TypeLongBlob:    TypeLongBlob,
			mysql.TypeBlob:        TypeBlob,
			mysql.TypeVarString:   TypeVarString,
			mysql.TypeString:      TypeString,
			mysql.TypeGeometry:    TypeGeometry,
		},
		nullableTypeMapping: map[string]string{
			"string": "dbo.NullString",
		},
		pluralize: pluralize.NewClient(),
	}
}

func (s TidbParser) ParseCreateTable(ddl string) ([]*Table, error) {
	nodes, _, err := parser.New().Parse(ddl, "", "")
	if err != nil {
		return nil, err
	}

	tables := make([]*Table, 0, len(nodes))
	for _, node := range nodes {
		stat, ok := node.(*ast.CreateTableStmt)
		if !ok {
			// create table stmt only!
			continue
		}

		table := &Table{
			Name:           stat.Table.Name.O,
			NomarlizedName: Normalize(stat.Table.Name.O),
			SingularName:   s.pluralize.Singular(Normalize(stat.Table.Name.O)),
			IsPlural:       s.pluralize.IsPlural(Normalize(stat.Table.Name.O)),
			Columns:        make([]*Column, 0, len(stat.Cols)),
			PrimaryKeys:    make([]string, 0, len(stat.Cols)),
		}

		for _, c := range stat.Cols {
			column := s.parseColumn(c)
			if column.IsPrimary {
				table.PrimaryKeys = append(table.PrimaryKeys, column.Name)
			}

			table.Columns = append(table.Columns, column)
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (s TidbParser) parseColumn(c *ast.ColumnDef) *Column {
	flag := c.Tp.GetFlag()

	column := &Column{
		Name:            c.Name.Name.O,
		NomarlizedName:  Normalize(c.Name.Name.O),
		Type:            s.columnTypeMapping[c.Tp.GetType()],
		Len:             c.Tp.GetFlen(),
		IsPrimary:       mysql.HasPriKeyFlag(flag),
		IsNotNull:       mysql.HasNotNullFlag(flag),
		IsUnique:        mysql.HasUniKeyFlag(flag),
		IsBinary:        mysql.HasBinaryFlag(flag),
		IsAutoIncrement: mysql.HasAutoIncrementFlag(flag),
		IsID:            strings.HasSuffix(Normalize(c.Name.Name.O), "ID"),
		IsPlural:        s.pluralize.IsPlural(Normalize(c.Name.Name.O)),
	}

	switch c.Tp.GetType() {
	case mysql.TypeVarchar, mysql.TypeString, mysql.TypeVarString, mysql.TypeJSON,
		mysql.TypeBlob, mysql.TypeTinyBlob, mysql.TypeMediumBlob, mysql.TypeLongBlob:
		column.GoType = "string"
	case mysql.TypeBit:
		column.GoType = "byte"
	case mysql.TypeTiny:
		if column.Len == 1 {
			column.GoType = "bool"
		} else {
			column.GoType = "int8"
		}
	case mysql.TypeShort, mysql.TypeInt24:
		column.GoType = "int"
	case mysql.TypeLong, mysql.TypeLonglong:
		column.GoType = "int64"
	case mysql.TypeFloat:
		column.GoType = "float32"
	case mysql.TypeDouble, mysql.TypeNewDecimal:
		column.GoType = "float64"
	case mysql.TypeDate, mysql.TypeDatetime, mysql.TypeTimestamp, mysql.TypeYear, mysql.TypeNewDate:
		column.GoType = "time.Time"
	case mysql.TypeDuration:
		column.GoType = "time.Duration"
	default:
		column.GoType = "any"
	}

	column.Comment, column.Enums = s.parseEnum(c)

	return column
}

func (s TidbParser) parseEnum(c *ast.ColumnDef) (string, map[string]string) {
	var comment string
	for _, o := range c.Options {
		if o.Tp != ast.ColumnOptionComment {
			continue
		}

		ve, ok := o.Expr.(*test_driver.ValueExpr)
		if !ok {
			continue
		}

		comment = ve.Datum.GetString()
	}

	enums := make(map[string]string)
	if comment == "" {
		return "", enums
	}

	//TODO:enums

	return comment, enums
}
