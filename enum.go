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

//dropAllEnum will drop all enum associated to table
func (s *Shifter) dropAllEnum(tx *pg.Tx, tableName string, skipPrompt bool) (
	err error) {

	tableModel := s.table[tableName]
	fields := util.GetStructField(tableModel)

	for _, refFeild := range fields {
		fType := util.FieldType(refFeild)
		// fmt.Println(tableName, fType, s.isEnum(tableName, fType))
		// enm := s.getEnumFromMethod(tableName)
		// fmt.Println(enm, enm[fType])
		if s.isEnum(tableName, fType) {
			// fmt.Println("IN for ", fType)
			if _, err = dropEnum(tx, tableName, fType, skipPrompt); err != nil {
				break
			}
		}
	}
	return
}

//upsertEnum will create/update enum of the given table
func (s *Shifter) upsertEnum(tx *pg.Tx, tableName string,
	enumName string) (err error) {

	var sEnumValue []string
	if sEnumValue, err = s.getEnum(tableName, enumName); err == nil {
		if _, created := enumCreated[enumName]; created == false {
			if enumSQL, enumExists := getEnumQuery(tx, enumName, sEnumValue); enumExists == false {
				err = s.createEnum(tx, tableName, enumName, enumSQL)
			} else {
				err = s.updateEnum(tx, tableName, enumName, sEnumValue)
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
func (s *Shifter) updateEnum(tx *pg.Tx, tableName,
	enumName string, sEnumValue []string) (err error) {

	var tEnumValue []string
	if tEnumValue, err = getDBEnumValue(tx, enumName); err == nil {

		if _, err = addRemoveEnum(tx, tableName, enumName,
			sEnumValue, tEnumValue, add); err == nil {

			// _, err = addRemoveEnum(tx, tableName, enumName,
			// 	tEnumValue, sEnumValue, drop)
		}
	}
	return
}

//addRemoveEnum will add or remove enum which exists in a but not in b
func addRemoveEnum(tx *pg.Tx, tableName, enumName string,
	a, b []string, op string) (isAlter bool, err error) {

	var enumValueMap = make(map[string]struct{})
	for _, curEnumVal := range b {
		enumValueMap[curEnumVal] = struct{}{}
	}

	for _, curEnumVal := range a {
		var curIsAlter bool
		if _, exists := enumValueMap[curEnumVal]; exists == false {
			switch op {
			case add:
				curIsAlter, err = addEnumVal(tx, tableName, enumName, curEnumVal)
			case drop:
				curIsAlter, err = dropEnumVal(tx, tableName, enumName, curEnumVal)
			}
			if err != nil {
				break
			}
			isAlter = isAlter || curIsAlter
		}
	}
	return
}

//addEnumVal will add enum value
func addEnumVal(tx *pg.Tx, tableName, enumName string, value string) (
	isAlter bool, err error) {

	sql := getEnumAddValSQL(enumName, value)
	if isAlter, err = execByChoice(tx, sql, false); err != nil {
		err = getWrapError(tableName, "add enum value", sql, err)
	}

	return
}

//dropEnumVal will drop enum value
func dropEnumVal(tx *pg.Tx, tableName, enumName string, value string) (
	isAlter bool, err error) {

	sql := getEnumDropValSQL(enumName, value)
	if isAlter, err = execByChoice(tx, sql, false); err != nil {
		err = getWrapError(tableName, "drop enum value", sql, err)
	}

	return
}

//dropEnum will drop enum
func dropEnum(tx *pg.Tx, tableName, enumName string, skipPrompt bool) (
	isAlter bool, err error) {

	sql := fmt.Sprintf("DROP TYPE IF EXISTS %v;", enumName)
	if isAlter, err = execByChoice(tx, sql, skipPrompt); err != nil {
		err = getWrapError(tableName, "drop enum", sql, err)
	} else if isAlter {
		fmt.Printf("Enum Dropped if exists: %v\n", enumName)
	}

	return
}

//getEnumAddValSQL will return enum add new value sql
func getEnumAddValSQL(enumName string, value string) (sql string) {
	sql = fmt.Sprintf("ALTER type %v ADD VALUE IF NOT EXISTS '%v';", enumName, value)
	return
}

//getEnumDropValSQL will return enum drop value sql
func getEnumDropValSQL(enumName string, value string) (sql string) {
	sql = fmt.Sprintf("ALTER type %v DROP VALUE IF EXISTS '%v';", enumName, value)
	return
}

//Create Enum Query for given table
func getEnumQuery(tx *pg.Tx, enumName string, enumValue []string) (
	query string, enumExists bool) {

	if enumExists = dbEnumExists(tx, enumName); enumExists == false {
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
func (s *Shifter) isEnum(tableName, enumName string) (flag bool) {

	enum := s.getEnumFromMethod(tableName)

	//checking table local enum list
	if _, flag = enum[enumName]; flag == false {
		//checking global enum list
		_, flag = s.enumList[enumName]
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
