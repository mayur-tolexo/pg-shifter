package shifter

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"text/template"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/mayur-tolexo/pg-shifter/model"
)

//sLog : structure log model
type sLog struct {
	StructNameWT string
	StructName   string
	TableName    string
	Data         []model.ColSchema
	Unique       []model.UKSchema
	Index        []model.Index
	Date         string
	importedPkg  map[string]struct{}
}

//pgToStructType to golang type mapping
var pgToGoType = map[string]string{
	"bigint":                      "int",
	"bigserial":                   "int",
	"varbit":                      "[]bytes",
	"boolean":                     "bool",
	"character":                   "string",
	"character varying":           "string",
	"double precision":            "float64",
	"integer":                     "int",
	"numeric":                     "float64",
	"real":                        "float64",
	"smallint":                    "int",
	"smallserial":                 "int",
	"serial":                      "int",
	"text":                        "string",
	"citext":                      "string",
	"time without time zone":      "time.Time",
	"time with time zone":         "time.Time",
	"timestamp without time zone": "time.Time",
	"timestamp with time zone":    "time.Time",
}

//goPkg to import when using that type in struct
var goPkg = map[string]string{
	"time.Time": "time",
}

//createAlterStructLog will create alter struct log
func (s *Shifter) createAlterStructLog(schema map[string]model.ColSchema,
	ukSchema []model.UKSchema, idx []model.Index) (err error) {

	var logStr string
	log := s.getSLogModel(schema, ukSchema, idx)
	if logStr, err = execLogTmpl(log); err == nil {
		err = s.logTableChange(logStr, log)
	}
	if err != nil {
		err = errors.New("Log Creation Error: " + err.Error())
	}
	return
}

//logTableChange will log table change in file
func (s *Shifter) logTableChange(logStr string, log sLog) (err error) {
	var (
		fp     *os.File
		logDir string
	)
	if logDir, err = s.makeStructLogDir(log.StructName); err == nil {

		file := logDir + "/" + log.StructNameWT + ".go"
		if fp, err = os.Create(file); err == nil {

			fp.WriteString(getWarning())
			fp.WriteString("package " + log.StructName + "\n")
			fp.WriteString(getImportPkg(log.importedPkg))
			fp.WriteString(logStr)
			exec.Command("gofmt", "-w", file).Run()
		}
	}
	return
}

//getSLogModel will return slog model
func (s *Shifter) getSLogModel(schema map[string]model.ColSchema,
	ukSchema []model.UKSchema, idx []model.Index) (log sLog) {

	tName := getTableName(schema)
	sName := tName
	if model, exists := s.table[sName]; exists {
		sName = getTableNameFromStruct(model)
	}

	sTime := time.Now().UTC()
	sNameWithTime := fmt.Sprintf("%v%v", sName, sTime.Unix())

	log = sLog{
		StructNameWT: sNameWithTime,
		StructName:   sName,
		TableName:    tName,
		Data:         getLogData(schema),
		Unique:       ukSchema,
		Index:        idx,
		Date:         sTime.Format("Mon _2 Jan 2006 15:04:05"),
		importedPkg:  make(map[string]struct{}),
	}
	return
}

//getImportPkg will append all imported packeges
func getImportPkg(pkg map[string]struct{}) (impPkg string) {
	if len(pkg) > 0 {
		impPkg += "import (\n"
		for k := range pkg {
			impPkg += "\"" + k + "\"\n"
		}
		impPkg += ")\n"
	}
	return
}

//makeStructLogDir will create struct log dir if not exists
func (s *Shifter) makeStructLogDir(structName string) (logDir string, err error) {
	if s.LogPath == "" {
		s.LogPath, err = os.Getwd()
		s.LogPath += "/log/"
	}
	if err == nil {
		logDir = s.LogPath + "/" + structName
		if _, err = os.Stat(logDir); os.IsNotExist(err) {
			err = os.MkdirAll(logDir, os.ModePerm)
		}
	}
	return
}

//execLogTmpl will execute log template
func execLogTmpl(log sLog) (logStr string, err error) {
	var (
		tmpl *template.Template
		buf  bytes.Buffer
	)
	tmplStr := getLogTmpl()
	if tmpl, err = template.New("template").
		Funcs(getLogTmplFunc()).
		Parse(tmplStr); err == nil {
		if err = tmpl.Execute(&buf, &log); err == nil {
			// fmt.Println(buf.String())
			logStr = buf.String()
		}
	}
	return
}

//getTmplFunc will return template functions
func getLogTmplFunc() template.FuncMap {
	return template.FuncMap{
		"Title":     getFieldName,
		"getSQLTag": getSQLTag,
	}
}

//getSQLTag will return struct sql tag from schema struct
func getSQLTag(schema model.ColSchema) (dType string) {
	dType = getStructDataType(schema)
	dType += getNullDTypeSQL(schema.IsNullable)
	dType += getDefaultDTypeSQL(schema)
	dType += getUniqueDTypeSQL(schema)
	dType += getConstraintTagSQL(schema)
	return
}

//getConstraintTagSQL will return sql tag constraint
func getConstraintTagSQL(schema model.ColSchema) (sql string) {
	switch schema.ConstraintType {
	case PrimaryKey:
		sql = " " + PrimaryKey
	case ForeignKey:
		sql = getStructConstraintSQL(schema)
	}
	return
}

//getTableNameFromStruct will return table name from struct
func getTableNameFromStruct(model interface{}) string {
	return reflect.TypeOf(model).Elem().Name()
}

//GetStructFieldType will return struct field type from schema datatype
func (l *sLog) GetStructFieldType(dataType string) (sType string) {
	var exists bool
	if sType, exists = pgToGoType[dataType]; exists == false {
		sType = "interface{}"
	}
	if pkg, exists := goPkg[sType]; exists {
		l.importedPkg[pkg] = struct{}{}
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

//getLogData will return data based on position of column in table
func getLogData(schema map[string]model.ColSchema) (data []model.ColSchema) {
	for _, v := range schema {
		data = append(data, v)
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Position < data[j].Position
	})
	return
}

//getLogTmpl will return struct log template
func getLogTmpl() (tmplStr string) {

	tmplStr = `
//{{ .StructNameWT }} : {{ .TableName }} table model [As on {{ .Date }} UTC]
type {{ .StructNameWT }} struct {
	tableName struct{} ` + "`sql:\"{{ .TableName }}\"`" + `
{{- range $key, $value := .Data}}
	{{ if eq $value.StructColumnName "" -}}
		{{ Title $value.ColumnName -}}
	{{ else -}}
		{{ $value.StructColumnName -}}
	{{ end -}}
	{{print " "}} {{ $.GetStructFieldType $value.DataType }}` + " `sql:\"{{ .ColumnName }},type:{{ getSQLTag $value }}\"`" + `
{{- end }}
}

{{ $length := len .Unique }} {{ if gt $length 0 }}
//UniqueKey of the table. This is for composite unique keys
func ({{ .StructNameWT }}) UniqueKey() []string {
	uk := []string{
		{{- range $key, $value := .Unique}}
			"{{ $value.Columns }}", //{{ $value.ConstraintName }}
		{{- end }}
	}
	return uk
}
{{ end }}

{{ $length := len .Index }} {{ if gt $length 0 }}
//Index of the table. For composite index use ,
//Default index type is btree. For gin index use gin value
func ({{ .StructNameWT }}) Index() map[string]string {
	idx := map[string]string{
		{{- range $key, $value := .Index}}
			"{{ $value.Columns }}": "{{ $value.IType }}", //{{ $value.IdxName }}
		{{- end }}
	}
	return idx
}
{{ end }}

`
	return
}

//getWarning will return warning string
func getWarning() string {
	return `
/*
@uthor Mayur Das<mayur.das4@gmail.com>
https://www.linkedin.com/in/mayurdaeron/

This is a history of the table before altering it.
We will use this to revert back to the previous state if needed.
----- DO NOT EDIT THIS -----
*/

`
}
