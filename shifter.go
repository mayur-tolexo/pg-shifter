package shifter

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/flaw"
	"github.com/mayur-tolexo/pg-shifter/model"
	"github.com/mayur-tolexo/pg-shifter/util"
)

var (
	tableCreated = make(map[interface{}]bool)
	enumCreated  = make(map[interface{}]struct{})
)

//Shifter model
type Shifter struct {
	table     map[string]interface{}
	enumList  map[string][]string
	hisExists bool
	logSQL    bool
	verbose   bool
	LogPath   string
}

func (s *Shifter) logMode(enable bool) {
	s.logSQL = enable
}

//NewShifter will return shifter model
func NewShifter(tables ...interface{}) *Shifter {
	s := &Shifter{
		table:    make(map[string]interface{}),
		enumList: make(map[string][]string),
	}
	if len(tables) > 0 {
		if err := s.SetTableModels(tables); err != nil {
			log.Fatalln(err)
		}
	}
	return s
}

//Verbose will enable executed sql printing in console
func (s *Shifter) Verbose(enable bool) *Shifter {
	s.verbose = enable
	return s
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
		// psql.StartLogging = true
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

//AlterAllTable will alter all tables
func (s *Shifter) AlterAllTable(conn *pg.DB, skipPromt bool) (err error) {

	s.Debug(conn)
	var tx *pg.Tx
	if tx, err = conn.Begin(); err == nil {
		for tableName := range s.table {
			if err = s.alterTable(tx, tableName, skipPromt); err != nil {
				break
			}
		}
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	} else {
		err = flaw.TxError(err)
	}
	return
}

//CreateStruct will create golang structure from postgresql table
func (s *Shifter) CreateStruct(conn *pg.DB, tableName string,
	filePath string) (err error) {

	var (
		tx      *pg.Tx
		tUK     []model.UKSchema
		idx     []model.Index
		tSchema map[string]model.ColSchema
	)
	if tx, err = conn.Begin(); err == nil {

		if tSchema, err = s.getTableSchema(tx, tableName); err == nil {
			if tUK, err = util.GetCompositeUniqueKey(tx, tableName); err == nil {
				if idx, err = util.GetIndex(tx, tableName); err == nil {
					s.LogPath = filePath
					err = s.createAlterStructLog(tSchema, tUK, idx, false)
				}
			}
		}

		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}
	return
}

//CreateStructFromStruct will create structure from shifter structures
//which are set in shifter map
func (s *Shifter) CreateStructFromStruct(conn *pg.DB, filePath string) (
	err error) {
	for tName := range s.table {
		if err = s.CreateStruct(conn, tName, filePath); err != nil {
			break
		} else if s.verbose {
			fmt.Print("Struct created: ")
			d := color.New(color.FgBlue, color.Bold)
			d.Println(tName)
		}
	}
	return
}

//CreateTrigger will create triggers mentioned on struct
func (s *Shifter) CreateTrigger(conn *pg.DB, tableName string) (err error) {
	var tx *pg.Tx
	s.Debug(conn)
	if tx, err = conn.Begin(); err == nil {
		err = s.createTrigger(tx, tableName)
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}
	return
}
