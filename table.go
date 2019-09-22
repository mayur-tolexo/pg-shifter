package shifter

import (
	"fmt"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/mayur-tolexo/flaw"
	util "github.com/mayur-tolexo/pg-shifter/util.go"
)

//Create Table in database
func (s *Shifter) createTable(tx *pg.Tx, tableName string, withDependency int) (err error) {
	tableModel := s.table[tableName]
	if _, alreadyCreated := tableCreated[tableModel]; alreadyCreated == false {
		tableCreated[tableModel] = true
		err = s.upsertEnum(tx, tableName)
		if err == nil {
			if withDependency == 1 {
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
					if err = s.upsertEnum(tx, refTable); err == nil {
						//creating dependent table
						if err = s.createTableDependencies(tx, refTableModel); err == nil {
							//creating table
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
		// createHistory(tx, refTable)
	} else {
		err = flaw.CreateError(err)
		fmt.Println("Table Error:", tableName, err.Error())
	}
	return
}

//dropTable will drop table
func (s *Shifter) dropTable(conn *pg.DB, tableName string, cascade bool) (err error) {
	if tableModel, exists := s.table[tableName]; exists {
		if err = conn.DropTable(tableModel,
			&orm.DropTableOptions{IfExists: true, Cascade: cascade}); err == nil {
			fmt.Println("Table Dropped if exists: ", tableName)
		} else {
			err = flaw.DropError(err)
			fmt.Println("Drop Error:", tableName, err.Error())
		}
	}
	return
}
