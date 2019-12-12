package shifter

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/flaw"
	m "github.com/mayur-tolexo/pg-shifter/model"
	"github.com/mayur-tolexo/pg-shifter/util"
)

var (
	tableCreated = make(map[interface{}]bool)
	enumCreated  = make(map[interface{}]struct{})
)

//Shifter model contains all the methods to migrate go struct to postgresql
type Shifter struct {
	table     map[string]interface{}
	enumList  map[string][]string
	hisExists bool
	logSQL    bool
	verbose   bool
	logPath   string
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

// SetLogPath will set logpath where alter struct log will be created.
//
//deafult path is pwd/log/
func (s *Shifter) SetLogPath(logPath string) *Shifter {
	s.logPath = logPath
	return s
}

//Verbose will enable executed sql printing in console
func (s *Shifter) Verbose(enable bool) *Shifter {
	s.verbose = enable
	return s
}

// CreateTable will create table if not exists.
//
// Parameters
//  conn: postgresql connection
//  model: struct pointer or string (table name)
// if model is table name then need to set shifter SetTableModel() before calling CreateTable()
func (s *Shifter) CreateTable(conn *pg.DB, model interface{}) (err error) {
	var (
		tx        *pg.Tx
		tableName string
	)
	if tableName, err = s.getTableName(model); err == nil {
		if tx, err = conn.Begin(); err == nil {
			err = s.createTable(tx, tableName, true)
			commitIfNil(tx, err)
		} else {
			err = flaw.TxError(err)
		}
	}
	return
}

// CreateEnum will create enum by enum name.
//
// Parameters
//  conn: postgresql connection
//  model: struct pointer or string (table name)
//  enumName: enum which you want to create
// if model is table name then need to set shifter SetTableModel() before calling CreateEnum()
func (s *Shifter) CreateEnum(conn *pg.DB, model interface{}, enumName string) (err error) {
	var (
		tx        *pg.Tx
		tableName string
	)
	if tableName, err = s.getTableName(model); err == nil {
		if tx, err = conn.Begin(); err == nil {
			err = s.createEnumByName(tx, tableName, enumName)
			commitIfNil(tx, err)
		} else {
			err = flaw.TxError(err)
		}
	}
	return
}

// CreateAllEnum will create all enums of the given table.
//
// Parameters
//  conn: postgresql connection
//  model: struct pointer or string (table name)
// if model is table name then need to set shifter SetTableModel() before calling CreateAllEnum()
func (s *Shifter) CreateAllEnum(conn *pg.DB, model interface{}) (err error) {
	var (
		tx        *pg.Tx
		tableName string
	)
	if tableName, err = s.getTableName(model); err == nil {
		if tx, err = conn.Begin(); err == nil {
			for enumName := range s.getEnumFromMethod(tableName) {
				if err = s.createEnumByName(tx, tableName, enumName); err != nil {
					break
				}
			}
			commitIfNil(tx, err)
		} else {
			err = flaw.TxError(err)
		}
	}
	return
}

// UpsertEnum will create/update enum by enum name.
//
// Parameters
//  conn: postgresql connection
//  model: struct pointer or string (table name)
//  enumName: enum which you want to upsert
// if model is table name then need to set shifter SetTableModel() before calling UpsertEnum()
func (s *Shifter) UpsertEnum(conn *pg.DB, model interface{}, enumName string) (err error) {
	var (
		tx        *pg.Tx
		tableName string
	)
	if tableName, err = s.getTableName(model); err == nil {
		if tx, err = conn.Begin(); err == nil {
			err = s.updateEnum(tx, tableName, enumName)
			commitIfNil(tx, err)
		} else {
			err = flaw.TxError(err)
		}
	}
	return
}

// UpsertAllEnum will create/update all enums of the given table.
//
// Parameters
//  conn: postgresql connection
//  model: struct pointer or string (table name)
// if model is table name then need to set shifter SetTableModel() before calling UpsertAllEnum()
func (s *Shifter) UpsertAllEnum(conn *pg.DB, model interface{}) (err error) {
	var (
		tx        *pg.Tx
		tableName string
	)
	if tableName, err = s.getTableName(model); err == nil {
		if tx, err = conn.Begin(); err == nil {
			err = s.upsertAllEnum(tx, tableName)
			commitIfNil(tx, err)
		} else {
			err = flaw.TxError(err)
		}
	}
	return
}

// CreateAllIndex will create all index of the given table.
//
// Parameters
//  conn: postgresql connection
//  model: struct pointer or string (table name)
//  skipPrompt: bool (default false | if false then before execution sql it will prompt for confirmation)
// if model is table name then need to set shifter SetTableModel() before calling CreateAllIndex()
func (s *Shifter) CreateAllIndex(conn *pg.DB, model interface{}, skipPrompt ...bool) (err error) {
	var (
		tx        *pg.Tx
		tableName string
	)
	if tableName, err = s.getTableName(model); err == nil {
		if tx, err = conn.Begin(); err == nil {
			err = s.createIndex(tx, tableName, getSP(skipPrompt))
			commitIfNil(tx, err)
		} else {
			err = flaw.TxError(err)
		}
	}
	return
}

// CreateAllUniqueKey will create table all composite unique key.
//
// Parameters
//  conn: postgresql connection
//  model: struct pointer or string (table name)
//  skipPrompt: bool (default false | if false then before execution sql it will prompt for confirmation)
// if model is table name then need to set shifter SetTableModel() before calling CreateAllUniqueKey()
func (s *Shifter) CreateAllUniqueKey(conn *pg.DB, model interface{}, skipPrompt ...bool) (err error) {
	var (
		tx        *pg.Tx
		tableName string
	)
	if tableName, err = s.getTableName(model); err == nil {
		if tx, err = conn.Begin(); err == nil {
			uk := s.getUKFromMethod(tableName)
			_, err = addCompositeUK(tx, tableName, uk, getSP(skipPrompt))
			commitIfNil(tx, err)
		} else {
			err = flaw.TxError(err)
		}
	}
	return
}

// UpsertAllUniqueKey will create/alter/drop composite unique keys of table.
//
// Parameters
//  conn: postgresql connection
//  model: struct pointer or string (table name)
//  skipPrompt: bool (default false | if false then before execution sql it will prompt for confirmation)
// If model is table name then need to set shifter SetTableModel() before calling CreateAllUniqueKey().
// If composite unique key is modified then also it will update.
// If composite unique key exists in table but doesn't exists in struct UniqueKey method
// then that will be dropped.
func (s *Shifter) UpsertAllUniqueKey(conn *pg.DB, model interface{}, skipPrompt ...bool) (err error) {
	var (
		tx        *pg.Tx
		tUK       []m.UKSchema
		tableName string
	)
	if tableName, err = s.getTableName(model); err == nil {
		if tx, err = conn.Begin(); err == nil {

			if tUK, err = util.GetCompositeUniqueKey(tx, tableName); err == nil {
				sUK := s.getUKFromMethod(tableName)
				if len(tUK) > 0 || len(sUK) > 0 {
					if _, err = dropCompositeUK(tx, tableName, tUK, sUK, getSP(skipPrompt)); err == nil {
						_, err = addCompositeUK(tx, tableName, sUK, getSP(skipPrompt))
					}
				}
			}

			commitIfNil(tx, err)
		} else {
			err = flaw.TxError(err)
		}
	}
	return
}

//CreateAllTable will create all tables
//before calling it you need to set the table model in shifter using SetTableModels()
func (s *Shifter) CreateAllTable(conn *pg.DB) (err error) {
	for tableName := range s.table {
		var tx *pg.Tx
		if tx, err = conn.Begin(); err == nil {
			if err = s.createTable(tx, tableName, true); err == nil {
				if err = s.createIndex(tx, tableName, true); err == nil {
					uk := s.getUKFromMethod(tableName)
					_, err = addCompositeUK(tx, tableName, uk, true)
				}
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

//DropTable will drop table if exists
//before calling it you need to set the table model in shifter using SetTableModel()
func (s *Shifter) DropTable(conn *pg.DB, tableName string, cascade bool) (err error) {
	err = s.dropTable(conn, tableName, cascade)
	return
}

//AlterAllTable will alter all tables
//before calling it you need to set the table model in shifter using SetTableModels()
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
		tUK     []m.UKSchema
		idx     []m.Index
		tSchema map[string]m.ColSchema
	)
	if tx, err = conn.Begin(); err == nil {

		if tSchema, err = s.getTableSchema(tx, tableName); err == nil {
			if tUK, err = util.GetCompositeUniqueKey(tx, tableName); err == nil {
				if idx, err = util.GetIndex(tx, tableName); err == nil {
					curLogPath := s.logPath
					s.logPath = filePath
					err = s.createAlterStructLog(tSchema, tUK, idx, false)
					s.logPath = curLogPath
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
//before calling it you need to set all the table models in shifter using SetTableModels()
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
//before calling it you need to set the table model in shifter using SetTableModel()
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
