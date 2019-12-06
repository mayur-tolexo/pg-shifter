package shifter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/structs"
	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/contour/adapter/psql"
	"github.com/mayur-tolexo/pg-shifter/model"
	"github.com/mayur-tolexo/pg-shifter/util"
)

//Alter Table
func (s *Shifter) alterTable(tx *pg.Tx, tableName string, skipPrompt bool) (err error) {

	var (
		columnSchema    []model.ColSchema
		constraint      []model.ColSchema
		uniqueKeySchema []model.UKSchema
	)
	_, isValid := s.table[tableName]

	if isValid == true {
		if columnSchema, err = util.GetColumnSchema(tx, tableName); err == nil {
			if constraint, err = util.GetConstraint(tx, tableName); err == nil {
				tSchema := MergeColumnConstraint(tableName, columnSchema, constraint)
				sSchema := s.GetStructSchema(tableName)

				if s.hisExists, err = util.IsAfterUpdateTriggerExists(tx, tableName); err == nil {

					if err = s.compareSchema(tx, tSchema, sSchema, skipPrompt); err == nil {
						if uniqueKeySchema, err = util.GetCompositeUniqueKey(tx, tableName); err == nil &&
							len(uniqueKeySchema) > 0 {
							defer func() { psql.LogMode(false) }()
							if s.Verbrose {
								psql.LogMode(true)
							}
							err = s.checkUniqueKeyToAlter(tx, tableName, uniqueKeySchema)
						}
					}
				}
				// printSchema(tSchema, sSchema)
			}
		}
	} else {
		msg := "Invalid Table Name: " + tableName
		fmt.Println(msg)
		err = errors.New(msg)
	}
	return
}

//compareSchema will compare then table and struct column scheam and change accordingly
func (s *Shifter) compareSchema(tx *pg.Tx, tSchema, sSchema map[string]model.ColSchema,
	skipPrompt bool) (err error) {

	var (
		added   bool
		removed bool
		modify  bool
	)

	defer func() { psql.LogMode(false) }()
	if s.Verbrose {
		psql.LogMode(true)
	}

	//adding column exists in struct but missing in db table
	if added, err = s.addRemoveCol(tx, sSchema, tSchema, Add, skipPrompt); err == nil {
		//removing column exists in db table but missing in struct
		if removed, err = s.addRemoveCol(tx, tSchema, sSchema, Drop, skipPrompt); err == nil {
			//TODO: modify column
			modify, err = s.modifyCol(tx, tSchema, sSchema, skipPrompt)
		}
	}
	if err == nil && (added || removed || modify) {
		if err = s.createAlterStructLog(tSchema); err == nil {

			if added || removed {
				tName := getTableName(sSchema)
				err = s.createTrigger(tx, tName)
			}
		}
	}
	return
}

//addRemoveCol will add/drop missing column which exists in a but not in b
func (s *Shifter) addRemoveCol(tx *pg.Tx, a, b map[string]model.ColSchema,
	op string, skipPrompt bool) (isAlter bool, err error) {

	for col, schema := range a {
		var curIsAlter bool
		if v, exists := b[col]; exists == false {
			if curIsAlter, err = s.alterCol(tx, schema, op, skipPrompt); err != nil {
				break
			}
		} else if v.StructColumnName == "" {
			v.StructColumnName = schema.StructColumnName
			b[col] = v
		}
		isAlter = isAlter || curIsAlter
	}
	return
}

//alterCol will add/drop column in table
func (s *Shifter) alterCol(tx *pg.Tx, schema model.ColSchema,
	op string, skipPrompt bool) (isAlter bool, err error) {

	switch op {
	case Add:
		isAlter, err = s.addCol(tx, schema, skipPrompt)
	case Drop:
		isAlter, err = s.dropCol(tx, schema, skipPrompt)
	}
	return
}

//alterCol will add column in table
func (s *Shifter) addCol(tx *pg.Tx, schema model.ColSchema,
	skipPrompt bool) (isAlter bool, err error) {

	dType := getAddColTypeSQL(schema)
	sql := getAddColSQL(schema.TableName, schema.ColumnName, dType)
	cSQL := getAddConstraintSQL(schema)

	if cSQL != "" {
		sql += "," + cSQL
	} else {
		sql += ";\n"
	}

	//checking history table exists
	if s.hisExists {
		hName := util.GetHistoryTableName(schema.TableName)
		dType = getStructDataType(schema)
		sql += getAddColSQL(hName, schema.ColumnName, dType)
	}
	//history alter sql end

	if isAlter, err = execByChoice(tx, sql, skipPrompt); err != nil {
		err = getWrapError(schema.TableName, "add column", sql, err)
	}
	return
}

//getAddColSQL will return add column sql
func getAddColSQL(tName, cName, dType string) (sql string) {
	sql = fmt.Sprintf("ALTER TABLE %v ADD %v %v", tName, cName, dType)
	return
}

//dropCol will drop column from table
func (s *Shifter) dropCol(tx *pg.Tx, schema model.ColSchema,
	skipPrompt bool) (isAlter bool, err error) {

	sql := getDropColSQL(schema.TableName, schema.ColumnName)
	//checking history table exists
	if s.hisExists {
		hName := util.GetHistoryTableName(schema.TableName)
		sql += ";\n" + getDropColSQL(hName, schema.ColumnName)
	}
	//history alter sql end

	if isAlter, err = execByChoice(tx, sql, skipPrompt); err != nil {
		err = getWrapError(schema.TableName, "drop column", sql, err)
	}
	return
}

//getAddColSQL will return add column sql
func getDropColSQL(tName, cName string) (sql string) {
	sql = fmt.Sprintf("ALTER TABLE %v DROP %v\n", tName, cName)
	return
}

//getFkName will return primary/unique/foreign key name
func getConstraintName(schema model.ColSchema) (keyName string) {
	var tag string
	switch schema.ConstraintType {
	case PrimaryKey:
		tag = PrimaryKeySuffix
	case Unique:
		tag = UniqueKeySuffix
	case ForeignKey:
		tag = ForeignKeySuffix
	}
	keyName = fmt.Sprintf("%v_%v_%v", schema.TableName, schema.ColumnName, tag)
	return
}

//getStructConstraintSQL will return constraint sql from scheam model
func getStructConstraintSQL(schema model.ColSchema) (sql string) {
	switch schema.ConstraintType {
	case PrimaryKey:
	case Unique:
	case ForeignKey:
		deleteTag := getConstraintTagByFlag(schema.DeleteType)
		updateTag := getConstraintTagByFlag(schema.UpdateType)
		sql = fmt.Sprintf(" REFERENCES %v(%v) ON DELETE %v ON UPDATE %v",
			schema.ForeignTableName, schema.ForeignColumnName, deleteTag, updateTag)
	}
	sql += getDefferSQL(schema)
	return
}

//getDefferSQL will reutrn deferable and initiall deferred sql
func getDefferSQL(schema model.ColSchema) (sql string) {
	if schema.IsDeferrable == Yes {
		sql = " " + Deferrable
		if schema.InitiallyDeferred == Yes {
			sql += " " + InitiallyDeferred
		} else {
			sql += " " + InitiallyImmediate
		}
	}
	return
}

//Get Constraint Tag by Flag
func getConstraintTagByFlag(flag string) (tag string) {
	switch flag {
	case "a":
		tag = NoAction
	case "r":
		tag = Restrict
	case "c":
		tag = Cascade
	case "n":
		tag = SetNull
	default:
		tag = SetDefault
	}
	return
}

//getAddColTypeSQL will return add column type sql
func getAddColTypeSQL(schema model.ColSchema) (dType string) {
	dType = getStructDataType(schema)
	// dType += getUniqueDTypeSQL(schema.ConstraintType)
	dType += getNullDTypeSQL(schema.IsNullable)
	dType += getDefaultDTypeSQL(schema)
	return
}

//getStructDataType will return data type from schema
func getStructDataType(schema model.ColSchema) (dType string) {
	var exists bool

	if schema.SeqName != "" {
		dType = getSerialType(schema.SeqDataType)
	} else if schema.DataType == UserDefined {
		dType = schema.UdtName
	} else if dType, exists = rDataAlias[schema.DataType]; exists == false {
		dType = schema.DataType
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
		str = fmt.Sprintf(" %v %v", Default, schema.ColumnDefault)
	}
	return
}

//getNullDTypeSQL will return null/not null constraint string if exists in column
func getNullDTypeSQL(isNullable string) (str string) {
	if isNullable != "" {
		if isNullable == Yes {
			str = " " + Null
		} else {
			str = " " + NotNull
		}
	}
	return
}

//getUniqueDTypeSQL will return unique constraint string if exists in column
func getUniqueDTypeSQL(schema model.ColSchema) (str string) {
	if schema.ConstraintType == Unique || schema.IsFkUnique {
		str = " " + Unique
	}
	return
}

//modifyCol will modify column of table by comparing with struct
func (s *Shifter) modifyCol(tx *pg.Tx, tSchema, sSchema map[string]model.ColSchema,
	skipPrompt bool) (isAlter bool, err error) {

	for col, tcSchema := range tSchema {
		var curIsAlter bool
		if scSchema, exists := sSchema[col]; exists {

			//modify data type
			if curIsAlter, err = s.modifyDataType(tx, tcSchema, scSchema, skipPrompt); err != nil {
				break
			}
			isAlter = isAlter || curIsAlter

			//if data type is not modified then only modify default type
			if curIsAlter == false {
				if curIsAlter, err = s.modifyDefault(tx, tcSchema, scSchema, skipPrompt); err != nil {
					break
				}
			}
			isAlter = isAlter || curIsAlter

			//modify not null constraint
			if curIsAlter, err = s.modifyNotNullConstraint(tx, tcSchema, scSchema, skipPrompt); err != nil {
				break
			}
			isAlter = isAlter || curIsAlter

			//modify pk/uk/fk constraint
			if curIsAlter, err = s.modifyConstraint(tx, tcSchema, scSchema, skipPrompt); err != nil {
				break
			}
			isAlter = isAlter || curIsAlter
		}
	}

	return
}

//modifyNotNullConstraint will modify not null by comparing table and structure
func (s *Shifter) modifyNotNullConstraint(tx *pg.Tx, tSchema, sSchema model.ColSchema,
	skipPrompt bool) (isAlter bool, err error) {

	if tSchema.IsNullable != sSchema.IsNullable {
		option := Set
		if sSchema.IsNullable == Yes {
			option = Drop
		}
		sql := getNotNullColSQL(sSchema.TableName, sSchema.ColumnName, option)
		if isAlter, err = execByChoice(tx, sql, skipPrompt); err != nil {
			err = getWrapError(sSchema.TableName, "modify not null", sql, err)
		}
	}
	return
}

//getDropDefaultSQL will return set/drop not null constraint sql
func getNotNullColSQL(tName, cName, option string) (sql string) {
	sql = fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v %v NOT NULL",
		tName, cName, option)
	return
}

//modifyDataType will modify column data type by comparing with structure
func (s *Shifter) modifyDataType(tx *pg.Tx, tSchema, sSchema model.ColSchema,
	skipPrompt bool) (isAlter bool, err error) {

	tDataType := getStructDataType(tSchema)
	sDataType := getStructDataType(sSchema)

	// fmt.Println("DType", tSchema.ColumnName, "T", tDataType, "S", sDataType)

	if tDataType != sDataType {
		//dropping default sql
		sql := getDropDefaultSQL(sSchema.TableName, sSchema.ColumnName)
		//modifying column type
		sql += getModifyColSQL(sSchema.TableName, sSchema.ColumnName, sDataType, sDataType)
		//adding back default sql
		sql += getSetDefaultSQL(sSchema.TableName, sSchema.ColumnName, sSchema.ColumnDefault)

		//checking history table exists
		if s.hisExists {
			hName := util.GetHistoryTableName(sSchema.TableName)
			sql += getModifyColSQL(hName, sSchema.ColumnName, sDataType, sDataType)
		}
		//history alter sql end

		if isAlter, err = execByChoice(tx, sql, skipPrompt); err != nil {
			err = getWrapError(sSchema.TableName, "modify datatype", sql, err)
		}
	}

	return
}

//getModifyColSQL will return modify column data type sql
func getModifyColSQL(tName, cName, dType, udtType string) (sql string) {

	sql = fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v TYPE %v USING (%v::text::%v);\n",
		tName, cName, dType, cName, udtType)
	return
}

//modifyDefault will modify default value by comparing table and structure
func (s *Shifter) modifyDefault(tx *pg.Tx, tSchema, sSchema model.ColSchema,
	skipPrompt bool) (isAlter bool, err error) {

	isSame := isSameDefault(tSchema, sSchema)

	//for primary key default is series so should remove it
	if tSchema.ConstraintType != PrimaryKey &&
		isSame == false {
		sql := ""
		if sSchema.ColumnDefault == "" {
			sql = getDropDefaultSQL(sSchema.TableName, sSchema.ColumnName)
		} else {
			sql = getSetDefaultSQL(sSchema.TableName, sSchema.ColumnName, sSchema.ColumnDefault)
		}
		if isAlter, err = execByChoice(tx, sql, skipPrompt); err != nil {
			err = getWrapError(sSchema.TableName, "modify default", sql, err)
		}
	}
	return
}

//isSameDefault will check table and struct default values are same or not
func isSameDefault(tSchema, sSchema model.ColSchema) (isSame bool) {
	tDefault := tSchema.ColumnDefault
	sDefault := sSchema.ColumnDefault

	if tDefault == "" {
		if tSchema.IsNullable == Yes && sDefault != "" {
			tDefault = Null
		}
		tDefault = strings.ToLower(tDefault)
		isSame = (tDefault == sDefault)
	} else if tSchema.ConstraintType != PrimaryKey && tSchema.SeqName == "" {

		if strings.Contains(tDefault, "::") &&
			strings.Contains(sDefault, "::") == false {
			dataType, exists := DataAlias[sSchema.DataType]
			if exists == false {
				dataType = sSchema.DataType
			}
			sDefault += "::" + dataType
			// tDefault = strings.Split(tDefault, "::")[0]
		}

		if hasQuote(sDefault) && hasQuote(tDefault) == false {
			tDefault = addQuote(tDefault)
		} else if hasQuote(tDefault) && hasQuote(sDefault) == false {
			sDefault = addQuote(sDefault)
		}

		tDefault = strings.ToLower(tDefault)
		// fmt.Println(tSchema.ColumnName, "T", tDefault, "S", sDefault)

		if tDefault == sDefault {
			isSame = true
		}
	}
	return
}

//addQuote will add single quote
func addQuote(a string) (qa string) {
	prefix := ""
	if strings.Contains(a, "::") {
		data := strings.Split(a, "::")
		if len(data) > 1 {
			a = data[0]
			prefix = "::" + strings.Join(data[1:], "")
		}
	}

	qa = "'" + a + "'" + prefix
	return
}

//hasQuote will check data have single quote
func hasQuote(a string) (hasQuote bool) {
	if strings.Contains(a, "::") {
		a = strings.Split(a, "::")[0]
	}
	if strings.HasPrefix(a, "'") &&
		strings.HasSuffix(a, "'") {
		hasQuote = true
	}
	return
}

//getDropDefaultSQL will return drop default constraint sql
func getDropDefaultSQL(tName, cName string) (sql string) {
	sql = fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v DROP DEFAULT;\n",
		tName, cName)
	return
}

//getSetDefaultSQL will return default column sql
func getSetDefaultSQL(tName, cName, dVal string) (sql string) {
	if dVal != "" {
		sql = fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v SET DEFAULT %v;\n",
			tName, cName, dVal)
	}
	return
}

//modifyConstraint will modify primary key/ unique key/ foreign key constraints by comparing table and structure
func (s *Shifter) modifyConstraint(tx *pg.Tx, tSchema, sSchema model.ColSchema,
	skipPrompt bool) (isAlter bool, err error) {

	// fmt.Println(sSchema.ColumnName, "T", tSchema.IsFkUnique, tSchema.ConstraintType, "S", sSchema.IsFkUnique, sSchema.ConstraintType)
	//if table and struct constraint doesn't match
	if tSchema.ConstraintType != sSchema.ConstraintType {
		if sSchema.ConstraintType == "" {
			isAlter, err = dropColAllConstraints(tx, tSchema, sSchema, skipPrompt)
		} else if tSchema.ConstraintType == "" {
			isAlter, err = addColAllConstraints(tx, tSchema, sSchema, skipPrompt)
		} else {
			isAlter, err = dropAndCreateConstraint(tx, tSchema, sSchema, skipPrompt)
		}
	} else if tSchema.ConstraintType == ForeignKey {
		isAlter, err = modifyFkAllConstraint(tx, tSchema, sSchema, skipPrompt)
	}

	if err == nil && isAlter == false {
		isAlter, err = modifyDeferrable(tx, tSchema, sSchema, skipPrompt)
	}
	return
}

//modifyFkAllConstraint will modify foreign key all constraints
func modifyFkAllConstraint(tx *pg.Tx, tSchema, sSchema model.ColSchema, skipPrompt bool) (
	isAlter bool, err error) {
	if isAlter, err = modifyFkUniqueConstraint(tx, tSchema, sSchema, skipPrompt); err == nil {
		var curAlter bool
		curAlter, err = modifyFkConstraint(tx, tSchema, sSchema, skipPrompt)
		isAlter = isAlter || curAlter
	}
	return
}

//modifyFkConstraint will modify foreign key of column if changed
func modifyFkConstraint(tx *pg.Tx, tSchema, sSchema model.ColSchema, skipPrompt bool) (
	isAlter bool, err error) {
	//if foreign table or column changed
	if tSchema.ForeignTableName != sSchema.ForeignTableName ||
		tSchema.ForeignColumnName != sSchema.ForeignColumnName ||
		tSchema.UpdateType != sSchema.UpdateType ||
		tSchema.DeleteType != sSchema.DeleteType {

		isAlter, err = dropAndCreateConstraint(tx, tSchema, sSchema, skipPrompt)
	}
	return
}

//dropAndCreateConstraint will drop current constraint and create new one
func dropAndCreateConstraint(tx *pg.Tx, tSchema, sSchema model.ColSchema, skipPrompt bool) (
	isAlter bool, err error) {
	fmt.Println("---dropping old and creating new constraint---")
	if isAlter, err = dropColAllConstraints(tx, tSchema, sSchema, skipPrompt); err == nil {
		var curAlter bool
		curAlter, err = addColAllConstraints(tx, tSchema, sSchema, skipPrompt)
		isAlter = isAlter || curAlter
	}
	return
}

//dropColConstraints will drop column all constraints
func dropColAllConstraints(tx *pg.Tx, tSchema, sSchema model.ColSchema, skipPrompt bool) (
	isAlter bool, err error) {

	if isAlter, err = dropConstraint(tx, tSchema, skipPrompt); err == nil {
		//TODO: also drop unique constraint if exists in table
		//with foreign key
		if tSchema.IsFkUnique && sSchema.IsFkUnique == false {
			var curAtler bool
			tSchema.ConstraintName = tSchema.FkUniqueName
			curAtler, err = dropConstraint(tx, tSchema, skipPrompt)
			isAlter = isAlter || curAtler
		}
	}
	return
}

//modifyFkUniqueConstraint will modify unique key constraint
//if exists with foreign key on same column
func modifyFkUniqueConstraint(tx *pg.Tx, tSchema, sSchema model.ColSchema,
	skipPrompt bool) (isAlter bool, err error) {
	if tSchema.IsFkUnique != sSchema.IsFkUnique {
		if sSchema.IsFkUnique {
			//adding unique constraint in table
			//as its exists with foreign key
			sSchema.ConstraintType = Unique
			isAlter, err = addConstraint(tx, sSchema, skipPrompt)
		} else if tSchema.IsFkUnique {
			//droping unique constraint from table
			//as its not exists with foreign key in struct anymore
			tSchema.ConstraintName = tSchema.FkUniqueName
			isAlter, err = dropConstraint(tx, tSchema, skipPrompt)
		}
	}
	return
}

//dropConstraint will drop constraint from table
func dropConstraint(tx *pg.Tx, tSchema model.ColSchema, skipPrompt bool) (isAlter bool, err error) {
	sql := getDropConstraintSQL(tSchema.TableName, tSchema.ConstraintName)
	if isAlter, err = execByChoice(tx, sql, skipPrompt); err != nil {
		err = getWrapError(tSchema.TableName, "drop constraint", sql, err)
	}
	return
}

//getDropConstraintSQL will return drop constraint sql
func getDropConstraintSQL(tName, constraintName string) (sql string) {
	sql = fmt.Sprintf("ALTER TABLE %v DROP CONSTRAINT %v;\n", tName, constraintName)
	return
}

//addColAllConstraints will add column all constraints
func addColAllConstraints(tx *pg.Tx, tSchema, sSchema model.ColSchema, skipPrompt bool) (
	isAlter bool, err error) {

	if isAlter, err = addConstraint(tx, sSchema, skipPrompt); err == nil {
		//TODO: also adding unique constraint if exists in struct
		//with foreign key
		if sSchema.IsFkUnique && tSchema.IsFkUnique == false {
			var curAtler bool
			sSchema.ConstraintType = Unique
			curAtler, err = addConstraint(tx, sSchema, skipPrompt)
			isAlter = isAlter || curAtler
		}
	}
	return
}

//addConstraint will add constraint on table column
func addConstraint(tx *pg.Tx, schema model.ColSchema, skipPrompt bool) (
	isAlter bool, err error) {

	sql := getAlterAddConstraintSQL(schema)
	if isAlter, err = execByChoice(tx, sql, skipPrompt); err != nil {
		err = getWrapError(schema.TableName, "add constraint", sql, err)
	}
	return
}

//getAlterAddConstraintSQL will return add constraint with alter table
func getAlterAddConstraintSQL(schema model.ColSchema) (sql string) {
	sql = getAddConstraintSQL(schema)
	sql = fmt.Sprintf("ALTER TABLE %v %v", schema.TableName, sql)
	return
}

//getAddConstraintSQL will return add constraint sql
func getAddConstraintSQL(schema model.ColSchema) (sql string) {
	if schema.ConstraintType != "" {
		sql = getStructConstraintSQL(schema)
		fkName := getConstraintName(schema)
		sql = fmt.Sprintf("ADD CONSTRAINT %v %v (%v) %v;\n",
			fkName, schema.ConstraintType, schema.ColumnName, sql)
	}
	return
}

//modifyDeferrable will modify add/drop constraint deferrable
func modifyDeferrable(tx *pg.Tx, tSchema, sSchema model.ColSchema, skipPrompt bool) (
	isAlter bool, err error) {

	// fmt.Println(tSchema.ColumnName, "T", tSchema.IsDeferrable, "S", sSchema.IsDeferrable)
	if tSchema.IsDeferrable != sSchema.IsDeferrable ||
		tSchema.InitiallyDeferred != sSchema.InitiallyDeferred {

		sSchema.ConstraintName = tSchema.ConstraintName
		sql := getDeferrableSQL(sSchema)
		if isAlter, err = execByChoice(tx, sql, skipPrompt); err != nil {
			err = getWrapError(tSchema.TableName, "modify deferrable", sql, err)
		}
	}
	return
}

//getDeferrableSQL will return deferrable sql
func getDeferrableSQL(schema model.ColSchema) (sql string) {

	sql = fmt.Sprintf("ALTER TABLE %v ALTER CONSTRAINT %v ", schema.TableName, schema.ConstraintName)

	//if deferrable then checking its initially deffered or initially immediate
	if schema.IsDeferrable == Yes {
		sql += Deferrable
		if schema.InitiallyDeferred == Yes {
			sql += " " + InitiallyDeferred
		} else {
			sql += " " + InitiallyImmediate
		}
	} else {
		sql += NotDeferrable
	}
	return
}

//getWrapError will return wrapped error for better debugging
func getWrapError(tName, op string, sql string, err error) (werr error) {
	msg := fmt.Sprintf("%v %v error %v\nSQL: %v",
		tName, op, err.Error(), sql)
	werr = errors.New(msg)
	return
}

//execByChoice will execute by choice
func execByChoice(tx *pg.Tx, sql string, skipPrompt bool) (
	isAlter bool, err error) {

	choice := util.GetChoice(sql, skipPrompt)
	if choice == util.Yes {
		isAlter = true
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

//MergeColumnConstraint : Merge Table Schema with Constraint
func MergeColumnConstraint(tName string, columnSchema,
	constraint []model.ColSchema) map[string]model.ColSchema {

	constraintMap := make(map[string]model.ColSchema)
	ColSchema := make(map[string]model.ColSchema)
	for _, curConstraint := range constraint {
		if v, exists := constraintMap[curConstraint.ColumnName]; exists {
			//if curent column is unique as foreign key as well
			if v.ConstraintType == Unique && curConstraint.ConstraintType == ForeignKey {
				curConstraint.FkUniqueName = v.ConstraintName
				v = curConstraint
				v.IsFkUnique = true
			} else if v.ConstraintType == ForeignKey && curConstraint.ConstraintType == Unique {
				v.FkUniqueName = curConstraint.ConstraintName
				v.IsFkUnique = true
			}
			constraintMap[curConstraint.ColumnName] = v
		} else {
			constraintMap[curConstraint.ColumnName] = curConstraint
		}
	}
	for _, curColumnSchema := range columnSchema {
		if curConstraint, exists :=
			constraintMap[curColumnSchema.ColumnName]; exists == true {
			curColumnSchema.ConstraintType = curConstraint.ConstraintType
			curColumnSchema.ConstraintName = curConstraint.ConstraintName
			curColumnSchema.IsDeferrable = curConstraint.IsDeferrable
			curColumnSchema.InitiallyDeferred = curConstraint.InitiallyDeferred
			curColumnSchema.ForeignTableName = curConstraint.ForeignTableName
			curColumnSchema.ForeignColumnName = curConstraint.ForeignColumnName
			curColumnSchema.UpdateType = curConstraint.UpdateType
			curColumnSchema.DeleteType = curConstraint.DeleteType
			curColumnSchema.IsFkUnique = curConstraint.IsFkUnique
			curColumnSchema.FkUniqueName = curConstraint.FkUniqueName
		}
		curColumnSchema.TableName = tName
		ColSchema[curColumnSchema.ColumnName] = curColumnSchema
	}
	return ColSchema
}
