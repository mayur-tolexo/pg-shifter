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
				err = execTableCreation(tx, tableName)
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
					//create/update enum
					if err = upsertEnum(tx, refTable); err == nil {
						//creating dependent table
						if err = createTableDependencies(tx, refTableModel); err == nil {
							//creating table
							err = execTableCreation(tx, refTable)
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
func execTableCreation(tx *pg.Tx, tableName string) (err error) {
	tableModel := util.Table[tableName]
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
