package shifter

import (
	"fmt"
	"strings"

	"github.com/fatih/structs"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/mayur-tolexo/flaw"
	"github.com/mayur-tolexo/pg-shifter/model"
	util "github.com/mayur-tolexo/pg-shifter/util.go"
)

//DataAlias alias to table value mapping
var DataAlias = map[string]string{
	"int8":        "bigint",
	"serial8":     "bigserial",
	"varbit":      "bit varying",
	"bool":        "boolean",
	"char":        "character",
	"varchar":     "character varying",
	"float8":      "double precision",
	"int":         "integer",
	"int4":        "integer",
	"decimal":     "numeric",
	"float4":      "real",
	"int2":        "smallint",
	"serial2":     "smallserial",
	"serial4":     "serial",
	"time":        "time without time zone",
	"timetz":      "time with time zone",
	"timestamp":   "timestamp without time zone",
	"timestamptz": "timestamp with time zone",
}

//Create Table in database
func (s *Shifter) createTable(tx *pg.Tx, tableName string, withDependency int) (err error) {
	tableModel := s.table[tableName]
	if _, alreadyCreated := tableCreated[tableModel]; alreadyCreated == false {
		tableCreated[tableModel] = true
		err = s.upsertEnum(tx, tableName)
		if err == nil {
			if withDependency == 1 {
				err = s.createTableDependencies(tx, tableModel)
			}
			if err == nil {
				err = s.execTableCreation(tx, tableName)
			}
		}
	}
	return
}

//Create all Tables if not exists whose Fk present in table Model
func (s *Shifter) createTableDependencies(tx *pg.Tx, tableModel interface{}) (err error) {
	fields := util.GetStructField(tableModel)
	for _, curField := range fields {
		refTable := util.RefTable(curField)
		if len(refTable) > 0 {
			if refTableModel, isValid := s.table[refTable]; isValid == true {
				if _, alreadyCreated := tableCreated[refTableModel]; alreadyCreated == false {

					//creating ref table dep tables
					tableCreated[refTableModel] = true
					//create/update enum
					if err = s.upsertEnum(tx, refTable); err == nil {
						//creating dependent table
						if err = s.createTableDependencies(tx, refTableModel); err == nil {
							//executin table creatin sql
							err = s.execTableCreation(tx, refTable)
						}
					}
					if err != nil {
						break
					}
				}
			}
		}
	}
	return
}

//execTableCreation will execute table creation
func (s *Shifter) execTableCreation(tx *pg.Tx, tableName string) (err error) {
	tableModel := s.table[tableName]
	if err = tx.CreateTable(tableModel,
		&orm.CreateTableOptions{IfNotExists: true}); err == nil {
		fmt.Println("Table Created if not exists: ", tableName)
		err = s.createHistory(tx, tableName)
	} else {
		err = flaw.CreateError(err)
		fmt.Println("Table Error:", tableName, err.Error())
	}
	return
}

//dropTable will drop table
func (s *Shifter) dropTable(conn *pg.DB, tableName string, cascade bool) (err error) {
	var tx *pg.Tx
	if tx, err = conn.Begin(); err == nil {
		if err = execTableDrop(tx, tableName, cascade); err == nil {
			err = s.dropHistory(tx, tableName, cascade)
		}
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	} else {
		err = flaw.TxError(err, "Table", tableName)
	}
	return
}

//execTableDrop will execute table drop
func execTableDrop(tx *pg.Tx, tableName string, cascade bool) (err error) {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %v", tableName)
	if cascade {
		sql += " CASCADE"
	}
	if _, err = tx.Exec(sql); err == nil {
		fmt.Println("Table Dropped if exists: ", tableName)
	} else {
		err = flaw.DropError(err)
		fmt.Println("Drop Error:", tableName, err.Error())
	}
	return
}

//Alter Table
func (s *Shifter) alterTable(tx *pg.Tx, tableName string) (err error) {
	// initStructTableMap()
	var (
		columnSchema []model.DBCSchema
		constraint   []model.DBCSchema
		// uniqueKeySchema []model.UniqueKeySchema
	)
	_, isValid := s.table[tableName]
	if isValid == true {
		if columnSchema, err = util.GetColumnSchema(tx, tableName); err == nil {
			if constraint, err = util.GetConstraint(tx, tableName); err == nil {
				tSchema := util.MergeColumnConstraint(columnSchema, constraint)
				sSchema := s.GetStructSchema(tableName)
				printSchema(tSchema, sSchema)

				// if err = checkTableToAlter(tx, tableSchema, tableModel, tableName); err == nil {
				// 	if uniqueKeySchema, err = util.GetCompositeUniqueKey(conn, tableName); err == nil {
				// 		if empty.IsEmptyInterface(uniqueKeySchema) == false {
				// 			if err = checkUniqueKeyToAlter(tx, uniqueKeySchema, tableName); err != nil {
				// 				return
				// 			}
				// 		}
				// 		tx.Commit()
				// 	} else {
				// 		fmt.Println("Composite unique key Fetch Error: ", tableName, err.Error())
				// 	}
				// }
			}
		}
	} else {
		fmt.Println("Invalid Table Name: ", tableName)
	}
	return
}

//GetStructSchema will return struct schema
func (s *Shifter) GetStructSchema(tableName string) (sSchema map[string]model.SCSchema) {
	tModel, isValid := s.table[tableName]
	sSchema = make(map[string]model.SCSchema)
	if isValid {
		fields := util.GetStructField(tModel)

		for _, field := range fields {
			var schema model.SCSchema
			tag := strings.ToLower(field.Tag.Get("sql"))
			schema.TableName = tableName
			schema.ColumnName = getColName(tag)
			schema.ColumnDefault = getColDefault(tag)
			schema.DataType, schema.CharMaxLen = getColType(tag)
			schema.IsNullable = getColIsNullable(tag)
			setColConstraint(&schema, tag)
			sSchema[schema.ColumnName] = schema
		}
	}
	return
}

//getColName will return col name from struct tag
func getColName(tag string) string {
	return strings.Split(tag, ",")[0]
}

//getColDefault will return col default value from struct tag
func getColDefault(tag string) (def string) {
	val := strings.Split(tag, "default ")
	if len(val) > 1 {
		def = strings.Split(val[1], " ")[0]
	}
	return
}

//getColType will return col type from struct tag
func getColType(tag string) (cType string, maxLen string) {
	val := strings.Split(tag, "type:")
	if len(val) > 1 {
		cType = strings.Split(val[1], " ")[0]
		maxLen = getColMaxChar(cType)
		cType = strings.Split(cType, "(")[0]
		if _, exists := DataAlias[cType]; exists {
			cType = DataAlias[cType]
		}
	}
	return
}

//getColIsNullable will return col nullable allowed from struct tag
func getColIsNullable(tag string) (nullable string) {
	nullable = Yes
	if strings.Contains(tag, " not null ") ||
		strings.Contains(tag, "primary key") {
		nullable = No
	}
	return
}

//getColMaxChar will return column max char type from struct tag type
func getColMaxChar(cType string) (maxLen string) {
	val := strings.Split(cType, "(")
	if len(val) > 1 {
		maxLen = strings.Split(val[1], ")")[0]
	} else if strings.Contains(cType, "varchar") || strings.Contains(cType, "char") {
		maxLen = "1"
	}
	return
}

//setColConstraint will set column constraints
//here we are setting the pk,uk or fk and deferrable and initially defered constraings
func setColConstraint(schema *model.SCSchema, tag string) {
	if strings.Contains(tag, "primary key") {

		schema.ConstraintType = PrimaryKey
		//in case of primary key reference table is itself
		schema.ForiegnTableName = schema.TableName
	} else if strings.Contains(tag, "unique") {

		schema.ConstraintType = Unique
		//in case of unique key reference table is itself
		schema.ForiegnTableName = schema.TableName
	} else if strings.Contains(tag, references) {

		schema.ConstraintType = ForeignKey
		referenceCheck := strings.Split(tag, references)

		//setting reference table and on cascade flags
		if len(referenceCheck) > 1 {
			schema.ForiegnTableName, schema.ForeignColumnName =
				getFkDetail(referenceCheck[1])
			schema.DeleteType = getConstraintFlagByKey(referenceCheck[1], "delete")
			schema.UpdateType = getConstraintFlagByKey(referenceCheck[1], "update")
		}
	}

	schema.IsDeferrable = No
	if strings.Contains(tag, "deferrable") {
		schema.IsDeferrable = Yes
	}
	schema.InitiallyDeferred = No
	if strings.Contains(tag, "initially deferred") {
		schema.InitiallyDeferred = Yes
	}
}

//Get FK table and column name
func getFkDetail(refCheck string) (table, column string) {
	refDetail := strings.Split(strings.Trim(refCheck, " "), " ")
	if len(refDetail) > 0 {
		if strings.Contains(refDetail[0], "(") == true {
			tableDetail := strings.Split(refDetail[0], "(")
			if len(tableDetail) > 1 {
				table = tableDetail[0]
				column = strings.Trim(tableDetail[1], ")")
			}
		}
	}
	return
}

//Get FK constraint Flag by key i.e. delete/update
func getConstraintFlagByKey(refCheck string, key string) (flag string) {
	flag = "d"
	if strings.Contains(refCheck, key) == true {
		keyCheck := strings.Split(refCheck, key)
		if len(keyCheck) > 1 {
			keyDetail := strings.Split(strings.Trim(keyCheck[1], " "), " ")
			keyDetailLen := len(keyDetail)
			if keyDetailLen > 0 {
				key = keyDetail[0]
				if (key == "set" || key == "no") && keyDetailLen > 1 {
					key += keyDetail[1]
				}
				flag = getConstraintFlag(key)
			}
		}
	}
	return
}

//Get FK constraint falg
func getConstraintFlag(key string) (flag string) {
	switch key {
	case "noaction":
		flag = "a"
	case "restrict":
		flag = "r"
	case "cascade":
		flag = "c"
	case "setnull":
		flag = "n"
	default:
		flag = "d"
	}
	return
}

func printSchema(tSchema map[string]model.DBCSchema, sSchema map[string]model.SCSchema) {
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
