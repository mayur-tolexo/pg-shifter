package shifter

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-pg/pg"
)

//GetStructTableName will return table name from table struct
func (s *Shifter) GetStructTableName(table interface{}) (
	tableName string, err error) {

	refObj := reflect.ValueOf(table)
	if refObj.Kind() != reflect.Ptr || refObj.Elem().Kind() != reflect.Struct {
		msg := fmt.Sprintf("Expected struct pointer of struct but found %v", refObj.Kind().String())
		err = errors.New(msg)
	} else {
		if field, exists := getStructTableNameField(table); exists {
			tableName = field.Tag.Get("sql")
		} else {
			msg := "tableName struct{} field not found in given struct"
			err = errors.New(msg)
		}
	}
	return
}

//SetTableModel will set table model
func (s *Shifter) SetTableModel(table interface{}) (err error) {
	var tableName string
	if tableName, err = s.GetStructTableName(table); err == nil {
		s.table[tableName] = table
	}
	return
}

//SetTableModels will set table models
func (s *Shifter) SetTableModels(tables []interface{}) (err error) {
	for _, table := range tables {
		if err = s.SetTableModel(table); err != nil {
			break
		}
	}
	return
}

//getTableTriggersTag will return trigger tag which need to create on table
func (s *Shifter) getTableTriggersTag(tableName string) (tag []string) {
	tag = make([]string, 0)
	if model, exists := s.table[tableName]; exists {
		if field, exists := getStructTableNameField(model); exists {
			trigger, exists := field.Tag.Lookup(TriggerTag)
			if exists == true {
				for _, v := range strings.Split(trigger, ",") {
					tag = append(tag, strings.TrimSpace(v))
				}
			} else {
				tag = []string{afterInsertTrigger, afterUpdateTrigger, afterDeleteTrigger, beforeUpdateTrigger}
			}
		}
	}
	return
}

//getStructTableNameField will return struct tableName field
func getStructTableNameField(model interface{}) (field reflect.StructField, exists bool) {
	refObj := reflect.ValueOf(model)
	if refObj.Kind() == reflect.Ptr {
		field, exists = refObj.Elem().Type().FieldByName("tableName")
	}
	return
}

//getTableName will check model is struct of string
//and return table name based on that
func (s *Shifter) getTableName(model interface{}) (
	tableName string, err error) {

	if reflect.TypeOf(model).Kind() == reflect.String {
		tableName = model.(string)
	} else {
		if err = s.SetTableModel(model); err == nil {
			tableName, err = s.GetStructTableName(model)
		}
	}

	return
}

//getSP will return skip prompt value
func getSP(val []bool) (skipPrompt bool) {
	if len(val) > 0 {
		skipPrompt = val[0]
	}
	return
}

//commitIfNil will commit transation if error is nil
func commitIfNil(tx *pg.Tx, err error) {
	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
}
