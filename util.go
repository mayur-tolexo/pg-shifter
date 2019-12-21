package shifter

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/pg-shifter/model"
)

//getStructTableName will return table name from table struct
func (s *Shifter) getStructTableName(table interface{}) (
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

//SetTableModel will set table struct pointer to shifter
func (s *Shifter) SetTableModel(table interface{}) (err error) {
	var tableName string
	if tableName, err = s.getStructTableName(table); err == nil {
		s.table[tableName] = table
	}
	return
}

// SetTableModels will set multiple table struct pointer to shifter.
// You can set all the table struct pointers and then perform operation by table name only
func (s *Shifter) SetTableModels(tables []interface{}) (err error) {
	for _, table := range tables {
		if err = s.SetTableModel(table); err != nil {
			break
		}
	}
	return
}

//SetEnum will set global enum list
func (s *Shifter) SetEnum(enum map[string][]string) (err error) {
	s.enumList = enum
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
			tableName, err = s.getStructTableName(model)
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

//getConstraint : Get Constraint of table from database
func getConstraint(tx *pg.Tx, tableName string) (constraint []model.ColSchema, err error) {
	query := `SELECT tc.constraint_type,
    tc.constraint_name, tc.is_deferrable, tc.initially_deferred, 
    kcu.column_name AS column_name, ccu.table_name AS foreign_table_name, 
    ccu.column_name AS foreign_column_name, pgc.confupdtype, pgc.confdeltype  
    FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu 
    ON tc.constraint_name = kcu.constraint_name 
    JOIN information_schema.constraint_column_usage AS ccu 
    ON ccu.constraint_name = tc.constraint_name 
    JOIN pg_constraint AS pgc ON pgc.conname = tc.constraint_name AND 
    conrelid=?::regclass::oid WHERE tc.constraint_type 
    IN('FOREIGN KEY','PRIMARY KEY','UNIQUE') AND tc.table_name = ?
    AND array_length(pgc.conkey,1) = 1;`
	if _, err = tx.Query(&constraint, query, tableName, tableName); err != nil {
		err = getWrapError(tableName, "table constraint", query, err)
	}
	return
}

//getColumnSchema : Get Column Schema of given table
func getColumnSchema(tx *pg.Tx, tableName string) (columnSchema []model.ColSchema, err error) {
	query := `SELECT col.column_name, col.column_default, col.data_type,
	col.ordinal_position as position,
	col.udt_name, col.is_nullable, col.character_maximum_length 
	, sq.sequence_name AS seq_name
	, sq.data_type AS seq_data_type
	FROM information_schema.columns col
	left join information_schema.sequences sq
	ON concat(sq.sequence_schema,'.',sq.sequence_name) = pg_get_serial_sequence(table_name, column_name)
	WHERE col.table_name = ?;`
	if _, err = tx.Query(&columnSchema, query, tableName); err != nil {
		err = getWrapError(tableName, "column schema", query, err)
	}
	return
}

//tableExists : Check if table exists in database
func tableExists(tx *pg.Tx, tableName string) (flag bool) {
	var num int
	sql := `SELECT 1 FROM pg_tables WHERE tablename = ?;`
	if _, err := tx.Query(pg.Scan(&num), sql, tableName); err != nil {
		fmt.Println("Table exists check error", err)
	} else if num == 1 {
		flag = true
	}
	return
}

//getDBEnumValue enum values by enumType from database
func getDBEnumValue(tx *pg.Tx, enumName string) (enumValue []string, err error) {
	query := `SELECT e.enumlabel as enum_value
	  FROM pg_enum e
	  JOIN pg_type t ON e.enumtypid = t.oid
	  WHERE t.typname = ?;`
	if _, err = tx.Query(&enumValue, query, enumName); err != nil {
		err = getWrapError(enumName, "enum type", query, err)
	}
	return
}

//dbEnumExists : Check if Enum Type Exists in database
func dbEnumExists(tx *pg.Tx, enumName string) (flag bool) {
	var num int
	enumSQL := `SELECT 1 FROM pg_type WHERE typname = ?;`
	if _, err := tx.Query(pg.Scan(&num), enumSQL, enumName); err == nil && num == 1 {
		flag = true
	}
	return
}
