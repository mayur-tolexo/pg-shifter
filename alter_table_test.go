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

func getAllTypeStruct() []interface{} {
	diffStruct := []interface{}{
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:xml"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:uuid"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:txid_snapshot"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:tsvector"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:tsquery"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:timestamptz"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:timestamp"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:timetz"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:time"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:text"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:smallint"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:real"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:polygon"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:polygon"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:point"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:pg_lsn"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:path"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:numeric"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:money"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:macaddr8"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:macaddr"`
		}{},
		&struct {
			tableName struct{} `sql:"test_table"`
			City      string   `json:"city" sql:"city,type:lseg"`
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
