package shifter

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/flaw"
	util "github.com/mayur-tolexo/pg-shifter/util.go"
)

//Create Enum in database
func createEnum(conn *pg.DB, tableName string) {
	tableModel := util.Table[tableName]
	fields := util.GetStructField(tableModel)
	for _, refFeild := range fields {
		fType := util.FieldType(refFeild)
		if _, exists := util.EnumList[fType]; exists {
			createEnumByName(conn, tableName, fType)
		}
	}
}

//Create Enum in database
func createEnumByName(conn *pg.DB, tableName, enumName string) (err error) {

	// var dbEnumVal []string
	if _, created := enumCreated[enumName]; created == false {
		if enumValue, exists := util.EnumList[enumName]; exists {
			if enumSQL, enumExists := getEnumQuery(conn, enumName, enumValue); enumExists == false {
				if _, err = conn.Exec(enumSQL); err == nil {
					enumCreated[enumName] = struct{}{}
					fmt.Printf("Enum %v created\n", enumName)
				} else {
					msg := fmt.Sprintf("Table: %v Enum: %v", tableName, enumName)
					err = flaw.CreateError(err, msg)
				}
			} else {
				// if dbEnumVal, err = getEnumValue(conn, enumName); err == nil {
				// 	//comparing old and new enum values
				// 	if newValue := compareEnumValue(dbEnumVal, enumValue); len(newValue) > 0 {
				// 		var choice string
				// 		enumAlterSQL := getEnumAlterQuery(enumName, newValue)
				// 		fmt.Printf("%v\nWant to continue (y/n): ", enumAlterSQL)
				// 		fmt.Scan(&choice)
				// 		if choice == util.YES_CHOICE {
				// 			if _, err = conn.Exec(enumAlterSQL); err == nil {
				// 				util.QueryFp.WriteString(fmt.Sprintf("-- ALTER ENUM\n%v\n", enumAlterSQL))
				// 				log.Println(fmt.Sprintf("----ALTER TABLE: %v", tableName))
				// 				log.Println(fmt.Sprintf("ENUM TYPE MODIFIED:\t%v\nPREV VALUE:\t%v\nNEW VALUE:\t%v\n",
				// 					enumName, dbEnumVal, enumValue))
				// 			} else {
				// 				fmt.Println("createEnumByName: Enum Alter Error: ", tableName, err.Error())
				// 			}
				// 		}
				// 	}
				// } else {
				// 	fmt.Println("createEnumByName: Enum Value fetch Error: ", tableName, err.Error())
				// }
			}
		} else {
			msg := fmt.Sprintf("Table: %v Enum: %v", tableName, enumName)
			err = flaw.CustomError(msg)
		}
	}
	return
}

//Create Enum Query for given table
func getEnumQuery(conn *pg.DB, enumName string, enumValue []string) (
	query string, enumExists bool) {

	if enumExists = util.EnumExists(conn, enumName); enumExists == false {
		query += fmt.Sprintf("CREATE type %v AS ENUM('%v'); ",
			enumName, strings.Join(enumValue, "','"))
	}
	return
}
