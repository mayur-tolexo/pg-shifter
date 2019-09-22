package shifter

import (
	"reflect"

	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/contour/adapter/psql"
	"github.com/mayur-tolexo/flaw"
)

var (
	tableCreated = make(map[interface{}]bool)
	enumCreated  = make(map[interface{}]struct{})
)

//Shifter model
type Shifter struct {
	table    map[string]interface{}
	enumList map[string][]string
}

//NewShifter will return shifter model
func NewShifter() *Shifter {
	return &Shifter{
		table:    make(map[string]interface{}),
		enumList: make(map[string][]string),
	}
}

//CreateTable will create table if not exists
func (s *Shifter) CreateTable(tx *pg.Tx, tableName string) (err error) {
	err = s.createTable(tx, tableName, 1)
	return
}

//CreateAllIndex will create table all index if not exists
func (s *Shifter) CreateAllIndex(tx *pg.Tx, tableName string, skipPrompt bool) (err error) {
	err = s.createIndex(tx, tableName, skipPrompt)
	return
}

//CreateAllTable will create all tables
func (s *Shifter) CreateAllTable(conn *pg.DB) (err error) {
	for tableName := range s.table {
		psql.StartLogging = true
		var tx *pg.Tx
		if tx, err = conn.Begin(); err == nil {
			if err = s.CreateTable(tx, tableName); err == nil {
				err = s.CreateAllIndex(tx, tableName, true)
			}
			if err == nil {
				tx.Commit()
			} else {
				tx.Rollback()
			}
		} else {
			err = flaw.TxError(err)
			break
		}
	}
	return
}

//CreateEnum will create enum by enum name
func (s *Shifter) CreateEnum(conn *pg.DB, tableName, enumName string) (err error) {
	var tx *pg.Tx
	if tx, err = conn.Begin(); err == nil {
		if err = s.createEnumByName(tx, tableName, enumName); err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	} else {
		err = flaw.TxError(err)
	}
	return
}

//DropTable will drop table if exists
func (s *Shifter) DropTable(conn *pg.DB, tableName string, cascade bool) (err error) {
	err = s.dropTable(conn, tableName, cascade)
	return
}

//SetTableModel will set table model
func (s *Shifter) SetTableModel(table interface{}) (err error) {
	refObj := reflect.ValueOf(table)
	if refObj.Kind() != reflect.Ptr || refObj.Elem().Kind() != reflect.Struct {
		err = flaw.CustomError("Invalid struct pointer")
	} else {
		refObj = refObj.Elem()
		if field, exists := refObj.Type().FieldByName("tableName"); exists {
			tableName := field.Tag.Get("sql")
			s.table[tableName] = table
		} else {
			err = flaw.CustomError("tableName field not found")
		}
	}
	return
}

//SetTableModels will set table models
func (s *Shifter) SetTableModels(tables []interface{}) (err error) {
	for _, table := range tables {
		if err = s.SetTableModel(table); err != nil {
			break
		}
	}
	return
}
