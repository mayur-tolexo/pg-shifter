package util

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/flaw"
	"github.com/mayur-tolexo/pg-shifter/model"
)

//variables
var (
	QueryFp *os.File
	Y       = "y"
	Yes     = "yes"
)

//GetColumnSchema : Get Column Schema of given table
func GetColumnSchema(conn *pg.DB, tableName string) (columnSchema []model.DBCSchema, err error) {
	query := `SELECT column_name,column_default, data_type, 
	udt_name, is_nullable,character_maximum_length 
	FROM information_schema.columns WHERE table_name = ?;`
	_, err = conn.Query(&columnSchema, query, tableName)
	return
}

//GetConstraint : Get Constraint of table from database
func GetConstraint(conn *pg.DB, tableName string) (constraint []model.DBCSchema, err error) {
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
	_, err = conn.Query(&constraint, query, tableName, tableName)
	return
}

//GetCompositeUniqueKey : Get composite unique key name and columns
func GetCompositeUniqueKey(conn *pg.DB, tableName string) (ukSchema []model.UKSchema, err error) {
	query := `select string_agg(c.column_name,',') as col, pgc.conname 
	from pg_constraint as pgc join
	information_schema.table_constraints tc on pgc.conname = tc.constraint_name, 
	unnest(pgc.conkey::int[]) as colNo join information_schema.columns as c 
	on c.ordinal_position = colNo and c.table_name = ? 
	where array_length(pgc.conkey,1)>1 and pgc.contype='u'
	and pgc.conrelid=c.table_name::regclass::oid group by pgc.conname;`
	_, err = conn.Query(&ukSchema, query, tableName)
	return
}

//EnumExists : Check if Enum Type Exists in database
func EnumExists(tx *pg.Tx, enumName string) (flag bool) {
	var num int
	enumSQL := `SELECT 1 FROM pg_type WHERE typname = ?;`
	if _, err := tx.Query(pg.Scan(&num), enumSQL, enumName); err == nil && num == 1 {
		flag = true
	}
	return
}

//TableExists : Check if Table Exists in database
func TableExists(conn *pg.DB, tableName string) (flag bool) {
	var num int
	enumSQL := `SELECT 1 FROM pg_tables WHERE tablename = ?;`
	if _, err := conn.Query(pg.Scan(&num), enumSQL, tableName); err == nil && num == 1 {
		flag = true
	}
	return
}

//GetEnumValue enum values by enumType
func GetEnumValue(tx *pg.Tx, enumName string) (enumValue []string, err error) {
	enumSQL := `SELECT e.enumlabel as enum_value
	  FROM pg_enum e
	  JOIN pg_type t ON e.enumtypid = t.oid
	  WHERE t.typname = ?;`
	if _, err = tx.Query(&enumValue, enumSQL, enumName); err != nil {
		err = flaw.SelectError(err, "Enum", enumName)
	}
	return
}

//GetStructField will return struct fields
func GetStructField(model interface{}) (fields map[reflect.Value]reflect.StructField) {
	refObj := reflect.ValueOf(model)
	fields = make(map[reflect.Value]reflect.StructField)
	if refObj.Kind() == reflect.Ptr {
		refObj = refObj.Elem()
	}
	if refObj.IsValid() {
		for i := 0; i < refObj.NumField(); i++ {
			refField := refObj.Field(i)
			refType := refObj.Type().Field(i)
			if refType.Name[0] > 'Z' {
				continue
			}
			if refType.Anonymous && refField.Kind() == reflect.Struct {
				embdFields := GetStructField(refField.Interface())
				mergeMap(fields, embdFields)
			} else {
				if _, exists := refType.Tag.Lookup("sql"); exists == false {
					fmt.Println("No SQL tag in", refType.Name)
					panic("sql tag not fround")
				}
				fields[refField] = refType
			}
		}
	}
	return
}

func mergeMap(a, b map[reflect.Value]reflect.StructField) {
	for k, v := range b {
		a[k] = v
	}
}

//getSQLTag will return sql tag
func getSQLTag(refField reflect.StructField) (sqlTag string) {
	sqlTag = refField.Tag.Get("sql")
	sqlTag = strings.ToLower(sqlTag)
	return
}

//FieldType will return field type
func FieldType(refField reflect.StructField) (fType string) {
	sqlTag := getSQLTag(refField)
	vals := strings.Split(sqlTag, "type:")
	if len(vals) > 1 {
		fType = vals[1]
		fType = strings.Trim(strings.Split(fType, " ")[0], " ")
	}
	return
}

//RefTable will reutrn reference table
func RefTable(refField reflect.StructField) (refTable string) {
	sqlTag := getSQLTag(refField)
	refTag := strings.Split(sqlTag, "references")
	if len(refTag) > 1 {
		refTable = strings.Split(refTag[1], "(")[0]
		refTable = strings.Trim(refTable, " ")
	}
	return
}

//GetChoice will ask user choice
func GetChoice(sql string, skipPrompt bool) (choice string) {
	if skipPrompt {
		choice = Yes
	} else {
		fmt.Printf("%v\nWant to continue (y/n):", sql)
		fmt.Scan(&choice)
		choice = strings.ToLower(choice)
		if choice == Y {
			choice = Yes
		}
	}
	return
}
