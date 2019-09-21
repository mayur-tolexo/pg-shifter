package shifter

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/flaw"
	util "github.com/mayur-tolexo/pg-shifter/util.go"
)

//upsertEnum Enum in database
func (s *Shifter) upsertEnum(tx *pg.Tx, tableName string) (err error) {
	tableModel := s.table[tableName]
	fields := util.GetStructField(tableModel)
	for _, refFeild := range fields {
		fType := util.FieldType(refFeild)
		if _, exists := s.enumList[fType]; exists {
			if err = s.createEnumByName(tx, tableName, fType); err != nil {
				break
			}
		}
	}
	return
}

//Create Enum in database
func (s *Shifter) createEnumByName(tx *pg.Tx, tableName, enumName string) (err error) {
	if _, created := enumCreated[enumName]; created == false {
		if enumValue, exists := s.enumList[enumName]; exists {
			if enumSQL, enumExists := getEnumQuery(tx, enumName, enumValue); enumExists == false {
				err = s.createEnum(tx, tableName, enumName, enumSQL)
			} else {
				err = s.updateEnum(tx, tableName, enumName)
			}
		} else {
			msg := fmt.Sprintf("Table: %v Enum: %v not found", tableName, enumName)
			err = flaw.CustomError(msg)
			fmt.Println("Enum Error:", msg)
		}
	}
	return
}

//createEnum will create enum
func (s *Shifter) createEnum(tx *pg.Tx, tableName, enumName, enumSQL string) (err error) {
	if _, err = tx.Exec(enumSQL); err == nil {
		enumCreated[enumName] = struct{}{}
		fmt.Printf("Enum %v created\n", enumName)
	} else {
		msg := fmt.Sprintf("Table: %v Enum: %v", tableName, enumName)
		err = flaw.CreateError(err, msg)
		fmt.Println("Enum Error:", msg, err.Error())
	}
	return
}

//updateEnum will update enum if changed in enum map
func (s *Shifter) updateEnum(tx *pg.Tx, tableName, enumName string) (err error) {
	var dbEnumVal []string
	enumValue := s.enumList[enumName]

	if dbEnumVal, err = util.GetEnumValue(tx, enumName); err == nil {
		//comparing old and new enum values
		if newValue := compareEnumValue(dbEnumVal, enumValue); len(newValue) > 0 {

			enumAlterSQL := getEnumAlterQuery(enumName, newValue)
			choice := util.GetChoice(enumAlterSQL)

			if choice == util.Yes {
				if _, err = tx.Exec(enumAlterSQL); err == nil {
				} else {
					msg := fmt.Sprintf("Table: %v Enum: %v", tableName, enumName)
					err = flaw.UpdateError(err, msg)
					fmt.Println("Enum Alter Error:", msg, err.Error())
				}
			}
		}
	}
	return
}

//Compare enum values
func compareEnumValue(dbEnumVal, structEnumValue []string) (newValue []string) {
	var enumValueMap = make(map[string]struct{})
	for _, curEnumVal := range dbEnumVal {
		enumValueMap[curEnumVal] = struct{}{}
	}
	for _, curEnumVal := range structEnumValue {
		if _, exists := enumValueMap[curEnumVal]; exists == false {
			newValue = append(newValue, curEnumVal)
		}
	}
	return
}

//Create enum alter query by enumType and new values
func getEnumAlterQuery(enumName string, newValue []string) (enumAlterSQL string) {
	for _, curValue := range newValue {
		enumAlterSQL += fmt.Sprintf("ALTER type %v ADD VALUE '%v'; ", enumName, curValue)
	}
	return
}

//Create Enum Query for given table
func getEnumQuery(tx *pg.Tx, enumName string, enumValue []string) (
	query string, enumExists bool) {

	if enumExists = util.EnumExists(tx, enumName); enumExists == false {
		query += fmt.Sprintf("CREATE type %v AS ENUM('%v'); ",
			enumName, strings.Join(enumValue, "','"))
	}
	return
}
