package shifter

import (
	"strings"

	"github.com/mayur-tolexo/pg-shifter/model"
	"gitlab.com/tolexo/plib/migrator/util"
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

//rDataAlias is reverse data alias
var rDataAlias = map[string]string{
	"bit varying":                 "varbit",
	"boolean":                     "bool",
	"character":                   "char",
	"character varying":           "varchar",
	"double precision":            "float8",
	"integer":                     "int",
	"time without time zone":      "time",
	"time with time zone":         "timetz",
	"timestamp without time zone": "timestamp",
	"timestamp with time zone":    "timestamptz",
}

//GetStructSchema will return struct schema
func (s *Shifter) GetStructSchema(tableName string) (sSchema map[string]model.ColSchema) {
	tModel, isValid := s.table[tableName]
	sSchema = make(map[string]model.ColSchema)
	if isValid {
		fields := util.GetStructField(tModel)

		for _, field := range fields {
			var schema model.ColSchema
			tag := strings.ToLower(field.Tag.Get("sql"))
			schema.TableName = tableName
			schema.StructColumnName = field.Name
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
	if strings.Contains(tag, "not null") ||
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
func setColConstraint(schema *model.ColSchema, tag string) {
	cSet := false
	if strings.Contains(tag, "primary key") {
		cSet = true
		schema.ConstraintType = PrimaryKey
		//in case of primary key reference table is itself
		schema.ForeignTableName = schema.TableName
		schema.ForeignColumnName = schema.ColumnName
	} else if strings.Contains(tag, "unique") {
		cSet = true
		schema.ConstraintType = Unique
		//in case of unique key reference table is itself
		schema.ForeignTableName = schema.TableName
	} else if strings.Contains(tag, references) {
		cSet = true
		schema.ConstraintType = ForeignKey
		referenceCheck := strings.Split(tag, references)

		//setting reference table and on cascade flags
		if len(referenceCheck) > 1 {
			schema.ForeignTableName, schema.ForeignColumnName = getFkDetail(referenceCheck[1])
			schema.DeleteType = getConstraintFlagByKey(referenceCheck[1], "delete")
			schema.UpdateType = getConstraintFlagByKey(referenceCheck[1], "update")
		}
	}
	if cSet {
		schema.IsDeferrable = No
		if strings.Contains(tag, "deferrable") {
			schema.IsDeferrable = Yes
		}
		schema.InitiallyDeferred = No
		if strings.Contains(tag, "initially deferred") {
			schema.InitiallyDeferred = Yes
		}
	}
}

//Get FK table and column name
func getFkDetail(refCheck string) (table, column string) {
	refDetail := strings.Split(strings.Trim(refCheck, " "), " ")
	if len(refDetail) > 0 {
		if strings.Contains(refDetail[0], "(") == true {
			refTable := strings.Trim(refDetail[0], " ")
			tableDetail := strings.Split(refTable, "(")
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
