package shifter

import (
	"testing"

	"github.com/mayur-tolexo/contour/adapter/psql"
	"github.com/stretchr/testify/assert"
)

var dataTypes = []string{
	"bigint", "bit", "varbit(32)", "boolean", "bool",
	"box", "bytea", "char(4)", "varchar(4)", "cidr", "circle",
	"date", "float8", "inet", "integer", "int", "int4",
	"json", "jsonb", "line", "lseg", "macaddr", "macaddr8",
	"money", "numeric", "path", "pg_lsn", "point", "polygon",
	"real", "smallint", "text", "time", "timetz", "timestamp",
	"timestamptz", "tsquery", "tsvector", "txid_snapshot",
	"uuid", "xml",
}

//TestAddress Table structure as in DB
type TestTable struct {
	tableName struct{} `sql:"test_table"`
	City      string   `json:"city" sql:"city,type:text"`
}

func getAllTypeStruct() []interface{} {
	diffStruct := []interface{}{
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:line"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:jsonb"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:json"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:int4"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:bigint"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:bit"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:varbit(32)"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:boolean"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:box"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:bytea"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:char(4)"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:varchar(4)"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:cidr"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:circle"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:date"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:float8"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:inet"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:integer"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:int"`
		}{},
	}
	return diffStruct
}

func TestAlterColumnTypeModify(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		assert := assert.New(t)
		err = s.CreateTable(conn, &TestTable{})
		for _, curStruct := range getAllTypeStruct() {
			err = s.AlterTable(conn, curStruct, true)
			assert.NoError(err)
		}
		err = s.DropTable(conn, &TestTable{}, true)
		assert.NoError(err)
	}
}
