package shifter

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/mayur-tolexo/pg-shifter/model"
)

//SLog : structure log model
type SLog struct {
	Name string
	Data map[string]model.ColSchema
	Date string
}

//pgToStructType to golang type mapping
var pgToGoType = map[string]string{
	"bigint":            "int",
	"bigserial":         "int",
	"varbit":            "[]bytes",
	"boolean":           "bool",
	"character":         "string",
	"character varying": "string",
	"double precision":  "float64",
	"integer":           "int",
	"numeric":           "float64",
	"real":              "float64",
	"smallint":          "int",
	"smallserial":       "int",
	"serial":            "int",
}

//createAlterStructLog will create alter struct log
func (s *Shifter) createAlterStructLog(schema map[string]model.ColSchema) (err error) {
	var (
		tmpl *template.Template
		buf  bytes.Buffer
	)
	tmplStr := getLogTmpl()
	tName := getTableName(schema)
	if model, exists := s.table[tName]; exists {
		tName = getTableNameFromStruct(model)
	}

	sTime := time.Now().UTC().Format("Mon _2 Jan 2006 15:04:05")
	log := SLog{Name: tName, Data: schema, Date: sTime}
	if tmpl, err = template.New("template").
		Funcs(getLogTmplFunc()).
		Parse(tmplStr); err == nil {
		if err = tmpl.Execute(&buf, log); err == nil {
			fmt.Println(buf.String())
		}
	}
	return
}

//getTmplFunc will return template functions
func getLogTmplFunc() template.FuncMap {
	return template.FuncMap{
		"Title":              getFieldName,
		"getStructFieldType": getStructFieldType,
		"getSQLTag":          getSQLTag,
	}
}

//getSQLTag will return struct sql tag from schema struct
func getSQLTag(schema model.ColSchema) (dType string) {
	dType = getStructDataType(schema)
	dType += getNullDTypeSQL(schema.IsNullable)
	dType += getDefaultDTypeSQL(schema)
	dType += getStructConstraintSQL(schema)
	return
}

//getTableNameFromStruct will return table name from struct
func getTableNameFromStruct(model interface{}) string {
	return reflect.TypeOf(model).Elem().Name()
}

//getStructFieldType will return struct field type from schema datatype
func getStructFieldType(dataType string) (sType string) {
	var exists bool
	if sType, exists = pgToGoType[dataType]; exists == false {
		sType = "interface{}"
	}
	return
}

//getTableName will return table name from schema
func getTableName(schema map[string]model.ColSchema) (tName string) {
	for _, v := range schema {
		tName = v.TableName
		break
	}
	return
}

//getFieldName will return field name in Camel Case
func getFieldName(k string) (f string) {
	f = strcase.ToCamel(k)
	return
}

func getLogTmpl() (tmplStr string) {

	tmplStr = `
//{{ .Name }} model as on {{ .Date }} UTC
type {{ .Name }} struct {
{{- range $key, $value := .Data}}
	{{ if eq $value.StructColumnName "" -}}
		{{ Title $value.ColumnName -}}
	{{ else -}}
		{{ $value.StructColumnName -}}
	{{ end -}}
	{{print " "}} {{ getStructFieldType $value.DataType }}` + " `sql:\"{{ .ColumnName }},type:{{ getSQLTag $value }}\"`" + `
{{- end }}
}`
	return
}
