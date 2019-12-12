package shifter

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/pg-shifter/util"
)

//upsertAllEnum will create/update all enum of the given table
func (s *Shifter) upsertAllEnum(tx *pg.Tx, tableName string) (err error) {

	tableModel := s.table[tableName]
	fields := util.GetStructField(tableModel)

	for _, refFeild := range fields {
		fType := util.FieldType(refFeild)
		if s.isEnum(tableName, fType) {
			if err = s.upsertEnum(tx, tableName, fType); err != nil {
				break
			}
		}
	}
	return
}

//upsertEnum will create/update enum of the given table
func (s *Shifter) upsertEnum(tx *pg.Tx, tableName string,
	enumName string) (err error) {

	var enumValue []string
	if enumValue, err = s.getEnum(tableName, enumName); err == nil {
		if _, created := enumCreated[enumName]; created == false {
			if enumSQL, enumExists := getEnumQuery(tx, enumName, enumValue); enumExists == false {
				err = s.createEnum(tx, tableName, enumName, enumSQL)
			} else {
				err = s.updateEnum(tx, tableName, enumName)
			}
		}
	}
	return
}

//Create Enum in database only if not exists
func (s *Shifter) createEnumByName(tx *pg.Tx, tableName, enumName string) (err error) {

	var enumValue []string
	if enumValue, err = s.getEnum(tableName, enumName); err == nil {
		if _, created := enumCreated[enumName]; created == false {
			if enumSQL, enumExists := getEnumQuery(tx, enumName, enumValue); enumExists == false {
				err = s.createEnum(tx, tableName, enumName, enumSQL)
			}
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
		err = getWrapError(tableName, "create enum", enumSQL, err)
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
			choice := util.GetChoice(enumAlterSQL, false)

			if choice == util.Yes {
				if _, err = tx.Exec(enumAlterSQL); err == nil {
				} else {
					err = getWrapError(tableName, "update enum", enumAlterSQL, err)
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

//getEnum will return enum values from enum name
func (s *Shifter) getEnum(tableName, enumName string) (
	enumValue []string, err error) {

	var exists bool
	enum := s.getEnumFromMethod(tableName)

	//checking table local enum list
	if enumValue, exists = enum[enumName]; exists == false {
		//checking global enum list
		enumValue, exists = s.enumList[enumName]
	} else {
		enum[enumName] = enumValue
	}

	if exists == false {
		msg := fmt.Sprintf("Table: %v Enum: %v not found", tableName, enumName)
		err = errors.New(msg)
	}
	return
}

//isEnum will check given type is enum or not
func (s *Shifter) isEnum(tableName, fType string) (flag bool) {

	var exists bool
	enum := s.getEnumFromMethod(tableName)
	//checking table local enum list
	if _, flag = enum[fType]; exists == false {
		//checking global enum list
		_, flag = s.enumList[fType]
	}
	return
}

//getEnumFromMethod will return table enum from Enum() method associted to table structure
func (s *Shifter) getEnumFromMethod(tableName string) (enum map[string][]string) {

	enum = make(map[string][]string)

	if dbModel, exists := s.table[tableName]; exists {
		refObj := reflect.ValueOf(dbModel)
		m := refObj.MethodByName("Enum")

		if m.IsValid() {
			out := m.Call([]reflect.Value{})
			if len(out) > 0 && out[0].Kind() == reflect.Map {
				if ev, ok := out[0].Interface().(map[string][]string); ok {
					enum = ev
				}
			}
		}
	}
	return
}
