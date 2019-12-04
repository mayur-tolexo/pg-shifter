package shifter

import (
	"fmt"

	"github.com/fatih/structs"
	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/pg-shifter/model"
	"github.com/mayur-tolexo/pg-shifter/util"
)

//compareSchema will compare then table and struct column scheam and change accordingly
func (s *Shifter) compareSchema(tx *pg.Tx, tSchema, sSchema map[string]model.ColSchema, skipPromt bool) (err error) {

	var (
		added   bool
		removed bool
	)
	// psql.StartLogging = true
	//adding column exists in struct but missing in db table
	if added, err = addRemoveCol(tx, sSchema, tSchema, Add, skipPromt); err == nil {
		//removing column exists in db table but missing in struct
		if removed, err = addRemoveCol(tx, tSchema, sSchema, Drop, skipPromt); err == nil {
			//TODO: modify column
		}
	}
	if added || removed {
		err = s.createAlterStructLog(tSchema)
	}
	return
}

//addRemoveCol will add/drop missing column which exists in a but not in b
func addRemoveCol(tx *pg.Tx, a, b map[string]model.ColSchema,
	op string, skipPrompt bool) (isAlter bool, err error) {

	for col, schema := range a {
		if v, exists := b[col]; exists == false {
			isAlter = true
			if err = alterCol(tx, schema, op, skipPrompt); err != nil {
				break
			}
		} else if v.StructColumnName == "" {
			v.StructColumnName = schema.StructColumnName
			b[col] = v
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
	sql := fmt.Sprintf("ALTER TABLE %v ADD %v %v",
		schema.TableName, schema.ColumnName, dType)

	sql += getAddFkConstraint(schema)
	choice := util.GetChoice(sql, skipPrompt)
	if choice == util.Yes {
		_, err = tx.Exec(sql)
	}
	return
}

//getAddFkConstraint will return fk constraint while adding column
func getAddFkConstraint(schema model.ColSchema) (sql string) {
	var tag string
	if schema.ConstraintType == ForeignKey {
		tag, sql = getConstraintSQL(schema)
		sql = `, ADD CONSTRAINT ` + tag + " " + ForeignKey + " " + sql
	}
	return
}

//addFk will add fk in column if exists in schema
func addFk(tx *pg.Tx, schema model.ColSchema, skipPrompt bool) (err error) {
	if schema.ConstraintType == ForeignKey {

		sql := getAddConstraintSQL(schema)
		choice := util.GetChoice(sql, skipPrompt)
		if choice == util.Yes {
			_, err = tx.Exec(sql)
		}
	}
	return
}

//getAddConstraintSQL will return add constraint sql
func getAddConstraintSQL(schema model.ColSchema) (sql string) {
	if schema.ConstraintType != "" {
		tag, constraintSQL := getConstraintSQL(schema)
		sql = fmt.Sprintf("ALTER TABLE %v ADD CONSTRAINT %v_%v_%v %v %v",
			schema.TableName, schema.TableName, schema.ColumnName, tag, schema.ConstraintType, constraintSQL)
	}
	return
}

//getConstraintSQL will return constraing sql
func getConstraintSQL(schema model.ColSchema) (tag, sql string) {
	switch schema.ConstraintType {
	case PrimaryKey:
		tag = "pkey"
		// sql = PrimaryKey
	case Unique:
		tag = "key"
		// sql = Unique
	case ForeignKey:
		tag = "fkey"
		deleteTag := getConstraintTagByFlag(schema.DeleteType)
		updateTag := getConstraintTagByFlag(schema.UpdateType)
		sql = fmt.Sprintf("(%v) REFERENCES %v(%v) ON DELETE %v ON UPDATE %v",
			schema.ColumnName, schema.ForeignTableName, schema.ForeignColumnName, deleteTag, updateTag)
	}
	if schema.IsDeferrable == Yes {
		sql += " DEFERRABLE"
	}
	if schema.InitiallyDeferred == Yes {
		sql += " INITIALLY DEFERRED"
	}
	return
}

//Get Constraint Tag by Flag
func getConstraintTagByFlag(flag string) (tag string) {
	switch flag {
	case "a":
		tag = "NO ACTION"
	case "r":
		tag = "RESTRICT"
	case "c":
		tag = "CASCADE"
	case "n":
		tag = "SET NULL"
	default:
		tag = "SET DEFAULT"
	}
	return
}

//getDataTypeByStruct will return data type from struct
func getDataTypeByStruct(schema model.ColSchema) (dType string) {
	dType = getStructDataType(schema)
	dType += getUniqueDTypeSQL(schema.ConstraintType)
	dType += getNullDTypeSQL(schema.IsNullable)
	dType += getDefaultDTypeSQL(schema)

	return
}

//getStructDataType will return data type from schema
func getStructDataType(schema model.ColSchema) (dType string) {
	var exists bool

	if schema.SeqName != "" {
		dType = getSerialType(schema.SeqDataType)
	} else {
		if dType, exists = rDataAlias[schema.DataType]; exists == false {
			dType = schema.DataType
		}
	}
	if schema.CharMaxLen != "" {
		dType += "(" + schema.CharMaxLen + ")"
	}
	return
}

//getSerialType will return data type for serial
func getSerialType(seqDataType string) (dType string) {
	switch seqDataType {
	case "bigint":
		dType = "bigserial"
	case "smallint":
		dType = "smallserial"
	default:
		dType = "serial"
	}
	return
}

//getDefaultDTypeSQL will return default constraint string if exists in column
func getDefaultDTypeSQL(schema model.ColSchema) (str string) {

	if schema.ColumnDefault != "" && schema.SeqName == "" {
		str = " DEFAULT " + schema.ColumnDefault
	}
	return
}

//getNullDTypeSQL will return null/not null constraint string if exists in column
func getNullDTypeSQL(isNullable string) (str string) {
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
	if constraintType == Unique {
		str = " " + Unique
	}
	return
}

//getStructConstraintSQL will return constraint sql from scheam model
func getStructConstraintSQL(schema model.ColSchema) (str string) {
	switch schema.ConstraintType {
	case PrimaryKey:
		str = " " + PrimaryKey
	case Unique:
		str = getUniqueDTypeSQL(Unique)
	case ForeignKey:
		deleteTag := getConstraintTagByFlag(schema.DeleteType)
		updateTag := getConstraintTagByFlag(schema.UpdateType)
		str = fmt.Sprintf(" REFERENCES %v(%v) ON DELETE %v ON UPDATE %v",
			schema.ForeignTableName, schema.ForeignColumnName, deleteTag, updateTag)
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
