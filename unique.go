package shifter

import (
	"fmt"
	"reflect"

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
			for i, ukFields := range val {
				ukName := fmt.Sprintf("uk_%v_%d", tName, i+1)
				uk[ukName] = ukFields
			}
		}
	}
	return
}

//Check unique key constraint to alter
func (s *Shifter) checkUniqueKeyToAlter(tx *pg.Tx, tName string,
	uniqueKeySchema []model.UKSchema) (err error) {

	uk := s.GetUniqueKey(tName)
	for _, curUK := range uniqueKeySchema {
		if _, exists := uk[curUK.ConstraintName]; exists {
			//TODO: check unique key diff
			delete(uk, curUK.ConstraintName)
		} else {
			sql := getDropConstraintSQL(tName, curUK.ConstraintName)
			if _, err = execByChoice(tx, sql, true); err != nil {
				err = getWrapError(tName, "composite unique key drop", sql, err)
				break
			}
		}
	}
	if len(uk) > 0 {
		sql := ""
		for ukName, ukFields := range uk {
			sql += getUniqueKeyQuery(tName, ukName, ukFields)
		}
		if _, err = execByChoice(tx, sql, true); err != nil {
			err = getWrapError(tName, "composite unique key alter", sql, err)
		}
	}
	return
}

//Get unique key query by tablename, unique key constraing name and table columns
func getUniqueKeyQuery(tableName string, constraintName string,
	column string) (uniqueKeyQuery string) {
	return fmt.Sprintf("ALTER TABLE %v DROP CONSTRAINT IF EXISTS %v;\nALTER TABLE %v ADD CONSTRAINT %v UNIQUE (%v);\n",
		tableName, constraintName, tableName, constraintName, column)
}
