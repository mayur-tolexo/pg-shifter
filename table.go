package shifter

import (
	"fmt"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/mayur-tolexo/flaw"
	"github.com/mayur-tolexo/pg-shifter/model"
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
		err = s.createHistory(tx, tableName)
	} else {
		err = flaw.CreateError(err)
		fmt.Println("Table Error:", tableName, err.Error())
	}
	return
}

//dropTable will drop table
func (s *Shifter) dropTable(conn *pg.DB, tableName string, cascade bool) (err error) {
	var tx *pg.Tx
	if tx, err = conn.Begin(); err == nil {
		if err = execTableDrop(tx, tableName, cascade); err == nil {
			err = s.dropHistory(tx, tableName, cascade)
		}
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	} else {
		err = flaw.TxError(err, "Table", tableName)
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
		err = flaw.DropError(err)
		fmt.Println("Drop Error:", tableName, err.Error())
	}
	return
}

//Alter Table
func (s *Shifter) alterTable(tx *pg.Tx, tableName string) (err error) {
	// initStructTableMap()
	var (
		columnSchema []model.DBCSchema
		constraint   []model.DBCSchema
		// uniqueKeySchema []model.UniqueKeySchema
	)
	_, isValid := s.table[tableName]
	if isValid == true {
		if columnSchema, err = util.GetColumnSchema(tx, tableName); err == nil {
			if constraint, err = util.GetConstraint(tx, tableName); err == nil {
				tableSchema := util.MergeColumnConstraint(columnSchema, constraint)

				fmt.Println(tableSchema)

				// if err = checkTableToAlter(tx, tableSchema, tableModel, tableName); err == nil {
				// 	if uniqueKeySchema, err = util.GetCompositeUniqueKey(conn, tableName); err == nil {
				// 		if empty.IsEmptyInterface(uniqueKeySchema) == false {
				// 			if err = checkUniqueKeyToAlter(tx, uniqueKeySchema, tableName); err != nil {
				// 				return
				// 			}
				// 		}
				// 		tx.Commit()
				// 	} else {
				// 		fmt.Println("Composite unique key Fetch Error: ", tableName, err.Error())
				// 	}
				// }
			}
		}
	} else {
		fmt.Println("Invalid Table Name: ", tableName)
	}
	return
}
