package shifter

import (
	"fmt"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/mayur-tolexo/flaw"
	util "github.com/mayur-tolexo/pg-shifter/util.go"
)

//Create Table in database
func createTable(tx *pg.Tx, tableName string, withDependency int) (err error) {
	tableModel := util.Table[tableName]
	if _, alreadyCreated := tableCreated[tableModel]; alreadyCreated == false {
		tableCreated[tableModel] = true
		err = upsertEnum(tx, tableName)
		if err == nil {
			if withDependency == 1 {
				err = createTableDependencies(tx, tableModel)
			}
			if err == nil {
				if err = tx.CreateTable(tableModel,
					&orm.CreateTableOptions{IfNotExists: true}); err == nil {
					fmt.Println("Table Created if not exists: ", tableName)
					// createHistory(tx, tableName)
				} else {
					err = flaw.CreateError(err)
					fmt.Println("Table Creation Error:", tableName, err.Error())
				}
			}
		}
	}
	return
}

//Create all Tables if not exists whose Fk present in table Model
func createTableDependencies(tx *pg.Tx, tableModel interface{}) (err error) {
	fields := util.GetStructField(tableModel)
	for _, curField := range fields {
		refTable := util.RefTable(curField)
		if len(refTable) > 0 {
			if refTableModel, isValid := util.Table[refTable]; isValid == true {
				if _, alreadyCreated := tableCreated[refTableModel]; alreadyCreated == false {

					//creating ref table dep tables
					tableCreated[refTableModel] = true
					if err = upsertEnum(tx, refTable); err == nil {
						err = createTableDependencies(tx, refTableModel)
					}

					//creating ref table
					if err == nil {
						if err = tx.CreateTable(refTableModel,
							&orm.CreateTableOptions{IfNotExists: true}); err == nil {
							fmt.Println("Table Created if not exists: ", refTable)
							// createHistory(tx, refTable)
						} else {
							err = flaw.CreateError(err)
							fmt.Println("Dependency Creation Error:", refTable, err.Error())
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
