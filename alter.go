package shifter

import (
	"fmt"

	"github.com/fatih/structs"
	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/pg-shifter/model"
	"github.com/mayur-tolexo/pg-shifter/util"
)

//compareSchema will compare then table and struct column scheam and change accordingly
func compareSchema(tx *pg.Tx, tSchema, sSchema map[string]model.ColSchema, skipPromt bool) (err error) {

	//adding column exists in struct but missing in db table
	if err = addRemoveCol(tx, sSchema, tSchema, Add, skipPromt); err == nil {
		//removing column exists in db table but missing in struct
		err = addRemoveCol(tx, tSchema, sSchema, Drop, skipPromt)
	}
	return
}

//addRemoveCol will add/drop missing column which exists in a but not in b
func addRemoveCol(tx *pg.Tx, a, b map[string]model.ColSchema,
	op string, skipPrompt bool) (err error) {

	for col, schema := range a {
		if _, exists := b[col]; exists == false {
			if err = alterCol(tx, schema, op, skipPrompt); err != nil {
				break
			}
		}
	}
	return
}

//alterCol will add/drop column in table
func alterCol(tx *pg.Tx, schema model.ColSchema,
	op string, skipPrompt bool) (err error) {

	switch op {
	case Add:
		err = addCol(tx, schema, skipPrompt)
	case Drop:
		err = dropCol(tx, schema, skipPrompt)
	}
	return
}

//alterCol will add column in table
func addCol(tx *pg.Tx, schema model.ColSchema, skipPrompt bool) (err error) {
	dType := getDataTypeByStruct(schema)
	sql := fmt.Sprintf("ALTER TABLE %v ADD %v %v;\n",
		schema.TableName, schema.ColumnName, dType)

	choice := util.GetChoice(sql, skipPrompt)
	if choice == util.Yes {
		_, err = tx.Exec(sql)
	}
	return
}

//getDataTypeByStruct will return data type from struct
func getDataTypeByStruct(schema model.ColSchema) (dType string) {
	var exists bool
	if dType, exists = rDataAlias[schema.DataType]; exists {
		if schema.CharMaxLen != "" {
			dType += "(" + schema.CharMaxLen + ")"
		}
	} else {
		dType = schema.DataType
	}

	dType += getUniqueDTypeSQL(schema.ConstraintType)
	dType += getNullDTypeSQL(schema.IsNullable)
	dType += getDefaultDTypeSQL(schema)

	return
}

//getDefaultDTypeSQL will return default constraint string if exists in column
func getDefaultDTypeSQL(schema model.ColSchema) (str string) {
	//checking default value of column
	if schema.ColumnDefault != "" {
		str = " DEFAULT " + schema.ColumnDefault
	}
	return
}

//getNullDTypeSQL will return null/not null constraint string if exists in column
func getNullDTypeSQL(isNullable string) (str string) {
	//checking null value allowed
	if isNullable != "" {
		if isNullable == Yes {
			str = " NULL"
		} else {
			str = " NOT NULL"
		}
	}
	return
}

//getUniqueDTypeSQL will return unique constraint string if exists in column
func getUniqueDTypeSQL(constraintType string) (str string) {
	//checking unique constraint of column
	if constraintType == Unique {
		str = " UNIQUE"
	}
	return
}

//dropCol will drop column in table
func dropCol(tx *pg.Tx, schema model.ColSchema, skipPrompt bool) (err error) {
	sql := fmt.Sprintf("ALTER TABLE %v DROP %v\n", schema.TableName, schema.ColumnName)
	choice := util.GetChoice(sql, skipPrompt)
	if choice == util.Yes {
		_, err = tx.Exec(sql)
	}
	return
}

//printSchema will print both schemas
func printSchema(tSchema, sSchema map[string]model.ColSchema) {
	for k, v1 := range tSchema {
		fmt.Println(k)
		if v2, exists := sSchema[k]; exists {
			tv := structs.Map(v1)
			sv := structs.Map(v2)
			for k, v := range tv {
				fmt.Println(k, v)
				fmt.Println(k, sv[k])
			}
			fmt.Println("---")
		}
	}
}
