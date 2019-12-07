package shifter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/pg-shifter/model"
)

//GetUniqueKey will return unique key fields of struct
func (s *Shifter) GetUniqueKey(tName string) (uk map[string]string) {
	dbModel := s.table[tName]
	refObj := reflect.ValueOf(dbModel)
	m := refObj.MethodByName("UniqueKey")
	uk = make(map[string]string)
	if m.IsValid() {
		out := m.Call([]reflect.Value{})
		if len(out) > 0 && out[0].Kind() == reflect.Slice {
			val := out[0].Interface().([]string)
			for _, ukFields := range val {
				fName := strings.Replace(ukFields, ",", "_", -1)
				ukName := fmt.Sprintf("%v_%v_%v", tName, fName, UniqueKeySuffix)
				if len(ukName) > 64 {
					ukName = ukName[:63]
				}
				uk[ukName] = ukFields
			}
		}
	}
	return
}

//Check unique key constraint to alter
func (s *Shifter) checkUniqueKeyToAlter(tx *pg.Tx, tName string,
	tUK []model.UKSchema, sUK map[string]string) (isAlter bool, err error) {

	if isAlter, err = dropCompositeUK(tx, tName, tUK, sUK); err == nil {
		var curAlter bool
		curAlter, err = addCompositeUK(tx, tName, sUK)
		isAlter = isAlter || curAlter
	}

	return
}

//addCompositeUK will add composite unique key which is not in table
func addCompositeUK(tx *pg.Tx, tName string,
	sUK map[string]string) (isAlter bool, err error) {

	if len(sUK) > 0 {
		sql := ""
		for ukName, ukFields := range sUK {
			//only for more than one fields
			if isCompositeUk(ukFields) {
				sql += getUniqueKeyQuery(tName, ukName, ukFields)
			}
		}
		if sql != "" {
			if isAlter, err = execByChoice(tx, sql, true); err != nil {
				err = getWrapError(tName, "add composite unique key", sql, err)
			}
		}
	}
	return
}

//dropCompositeUK will drop composite unique key if not exists in struct
func dropCompositeUK(tx *pg.Tx, tName string, tUK []model.UKSchema,
	sUK map[string]string) (isAlter bool, err error) {

	for _, curTableUK := range tUK {
		var curAlter bool
		if _, exists := sUK[curTableUK.ConstraintName]; exists {
			//TODO: check unique key diff
			delete(sUK, curTableUK.ConstraintName)
		} else {
			sql := getDropConstraintSQL(tName, curTableUK.ConstraintName)
			if curAlter, err = execByChoice(tx, sql, true); err != nil {
				err = getWrapError(tName, "drop composite unique key", sql, err)
				break
			}
		}
		isAlter = isAlter || curAlter
	}
	return
}

//isCompositeUk will check unique is composite or not
func isCompositeUk(fields string) (isComposite bool) {
	if strings.Contains(fields, ",") {
		isComposite = true
	}
	return
}

//Get unique key query by tablename, unique key constraing name and table columns
func getUniqueKeyQuery(tableName string, constraintName string,
	column string) (uniqueKeyQuery string) {
	return fmt.Sprintf("ALTER TABLE %v DROP CONSTRAINT IF EXISTS %v;\nALTER TABLE %v ADD CONSTRAINT %v UNIQUE (%v);\n",
		tableName, constraintName, tableName, constraintName, column)
}
