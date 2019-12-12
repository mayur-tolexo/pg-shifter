package shifter

import (
	"strings"

	"github.com/mayur-tolexo/pg-shifter/model"
	"github.com/mayur-tolexo/pg-shifter/util"
)

//pgAlias is postgresql alias to actual type
var pgAlias = map[string]string{
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

//rPGAlias is reverse postgresql alias
var rPGAlias = map[string]string{
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
			schema.ColumnDefault, schema.DefaultExists = getColDefault(tag)
			schema.DataType, schema.CharMaxLen = getColType(tag)
			schema.IsNullable = getColIsNullable(tag)
			s.setColConstraint(&schema, tag)
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
func getColDefault(tag string) (def string, exists bool) {
	val := strings.Split(tag, "default ")
	if len(val) > 1 {
		def = strings.Split(val[1], " ")[0]
		exists = true
	}
	return
}

//getColType will return col type from struct tag
func getColType(tag string) (cType string, maxLen string) {
	val := strings.Split(tag, "type:")
	if len(val) > 1 {
		val[1] = strings.TrimSpace(val[1])
		cType = strings.Split(val[1], " ")[0]
		maxLen = getColMaxChar(cType)
		cType = strings.Split(cType, "(")[0]
		if _, exists := pgAlias[cType]; exists {
			cType = pgAlias[cType]
		}
	}
	return
}

//getColIsNullable will return col nullable allowed from struct tag
func getColIsNullable(tag string) (nullable string) {
	nullable = yes
	if strings.Contains(tag, notNullTag) ||
		strings.Contains(tag, primaryKeyTag) {
		nullable = no
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
func (s *Shifter) setColConstraint(schema *model.ColSchema, tag string) {
	cSet := false
	if strings.Contains(tag, primaryKeyTag) {
		cSet = true
		schema.ConstraintType = primaryKey
		//in case of primary key reference table is itself
		schema.ForeignTableName = schema.TableName
		schema.ForeignColumnName = schema.ColumnName
	} else if strings.Contains(tag, uniqueKeyTag) {
		cSet = true
		schema.ConstraintType = uniqueKey
		//in case of unique key reference table is itself
		schema.ForeignTableName = schema.TableName
	}
	if strings.Contains(tag, referencesTag) {
		cSet = true
		if schema.ConstraintType != "" {
			schema.IsFkUnique = true
		}
		schema.ConstraintType = foreignKey
		referenceCheck := strings.Split(tag, referencesTag)

		//setting reference table and on cascade flags
		if len(referenceCheck) > 1 {
			schema.ForeignTableName, schema.ForeignColumnName = getFkDetail(referenceCheck[1])
			schema.DeleteType = getConstraintFlagByKey(referenceCheck[1], deleteTag)
			schema.UpdateType = getConstraintFlagByKey(referenceCheck[1], updateTag)
		}
	}

	if cSet {
		schema.IsDeferrable = no
		if strings.Contains(tag, "deferrable") {
			schema.IsDeferrable = yes
		}
		schema.InitiallyDeferred = no
		if strings.Contains(tag, "initially deferred") {
			schema.InitiallyDeferred = yes
		}
	}
	s.addConstraintFromUkMap(schema)
}

//addConstraintFromUkMap will add constraint from unique key map defined on struct
func (s *Shifter) addConstraintFromUkMap(schema *model.ColSchema) {
	var colFound bool
	if uk := s.getUKFromMethod(schema.TableName); len(uk) > 0 {
		for _, fields := range uk {
			if fields == schema.ColumnName {
				colFound = true
				break
			}
		}
	}

	if colFound {
		if schema.ConstraintType == "" {
			schema.ConstraintType = uniqueKey
			schema.IsDeferrable = no
			schema.InitiallyDeferred = no
		} else if schema.ConstraintType == foreignKey {
			schema.IsFkUnique = true
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
	flag = "a"
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
	case restrictTag:
		flag = "r"
	case cascadeTag:
		flag = "c"
	case setNullTag:
		flag = "n"
	case setDefault:
		flag = "d"
	default:
		flag = "a"
	}
	return
}
