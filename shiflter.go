package shifter

import (
	"fmt"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/mayur-tolexo/flaw"
	util "github.com/mayur-tolexo/pg-shifter/util.go"
)

var (
	tableCreated = make(map[interface{}]bool)
	enumCreated  = make(map[interface{}]struct{})
)

//Create Table in database
func createTable(conn *pg.DB, tableName string, withDependency int) (err error) {
	tableModel := util.Table[tableName]
	if _, alreadyCreated := tableCreated[tableModel]; alreadyCreated == false {
		tableCreated[tableModel] = true
		createEnum(conn, tableName)
		// if withDependency == 1 {
		// 	createTableDependencies(conn, tableModel)
		// }
		if err = conn.CreateTable(tableModel,
			&orm.CreateTableOptions{IfNotExists: true}); err == nil {
			fmt.Println("Table Created if not exists: ", tableName)
			// createHistory(conn, tableName)
		} else {
			err = flaw.CreateError(err)
		}
	}
	return
}
