package shifter

import (
	"fmt"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/mayur-tolexo/flaw"
	"github.com/mayur-tolexo/pg-shifter/util"
)

//Create Table in database
func (s *Shifter) createTable(tx *pg.Tx, tableName string, withDependency bool) (err error) {
	tableModel := s.table[tableName]
	if _, alreadyCreated := tableCreated[tableModel]; alreadyCreated == false {
		tableCreated[tableModel] = true
		err = s.upsertAllEnum(tx, tableName)
		if err == nil {
			if withDependency {
				err = s.createTableDependencies(tx, tableModel)
			}
			if err == nil {
				err = s.execTableCreation(tx, tableName)
			}
		}
	}
	return
}

//Create all Tables if not exists whose Fk present in table Model
func (s *Shifter) createTableDependencies(tx *pg.Tx, tableModel interface{}) (err error) {
	fields := util.GetStructField(tableModel)
	for _, curField := range fields {
		refTable := util.RefTable(curField)
		if len(refTable) > 0 {
			if refTableModel, isValid := s.table[refTable]; isValid == true {
				if _, alreadyCreated := tableCreated[refTableModel]; alreadyCreated == false {

					//creating ref table dep tables
					tableCreated[refTableModel] = true
					//create/update enum
					if err = s.upsertAllEnum(tx, refTable); err == nil {
						//creating dependent table
						if err = s.createTableDependencies(tx, refTableModel); err == nil {
							//executin table creatin sql
							err = s.execTableCreation(tx, refTable)
						}
					}
					if err != nil {
						break
					}
				}
			}
		}
	}
	return
}

//execTableCreation will execute table creation
func (s *Shifter) execTableCreation(tx *pg.Tx, tableName string) (err error) {
	tableModel := s.table[tableName]
	if err = tx.CreateTable(tableModel,
		&orm.CreateTableOptions{IfNotExists: true}); err == nil {
		fmt.Println("Table Created if not exists: ", tableName)
		err = s.createHistory(tx, tableName)
	} else {
		err = flaw.CreateError(err)
		fmt.Println("Table Error:", tableName, err.Error())
	}
	return
}

//dropTable will drop table
func (s *Shifter) dropTable(tx *pg.Tx, tableName string, cascade bool) (err error) {
	var (
		log    sLog
		fData  []byte
		exists bool
	)
	if log, fData, exists, err = s.generateTableStructSchema(tx, tableName, true); err == nil &&
		exists {
		if err = execTableDrop(tx, tableName, cascade); err == nil {
			if err = s.dropHistory(tx, tableName, cascade); err == nil {
				err = s.logTableChange(log, fData)
			}
		}
	}
	return
}

//execTableDrop will execute table drop
func execTableDrop(tx *pg.Tx, tableName string, cascade bool) (err error) {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %v", tableName)
	if cascade {
		sql += " CASCADE"
	}
	if _, err = tx.Exec(sql); err == nil {
		fmt.Println("Table Dropped if exists: ", tableName)
	} else {
		err = getWrapError(tableName, "drop table", sql, err)
	}
	return
}
