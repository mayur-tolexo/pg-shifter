package shifter

import (
	"github.com/go-pg/pg"
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

//CreateAllTable will create all tables
func (s *Shifter) CreateAllTable(conn *pg.DB) (err error) {
	for tableName := range s.table {
		var tx *pg.Tx
		if tx, err = conn.Begin(); err == nil {
			err = s.CreateTable(tx, tableName)
		} else {
			err = flaw.TxError(err)
			break
		}
	}
	return
}
