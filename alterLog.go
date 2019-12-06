package shifter

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"text/template"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/mayur-tolexo/pg-shifter/model"
)

//sLog : structure log model
type sLog struct {
	StructName  string
	TableName   string
	Data        map[string]model.ColSchema
	Date        string
	importedPkg map[string]struct{}
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
	"time without time zone":      timePkgTime,
	"time with time zone":         timePkgTime,
	"timestamp without time zone": timePkgTime,
	"timestamp with time zone":    timePkgTime,
}

//createAlterStructLog will create alter struct log
func (s *Shifter) createAlterStructLog(schema map[string]model.ColSchema) (err error) {

	var (
		fp     *os.File
		logDir string
		logStr string
	)
	tName := getTableName(schema)
	sName := tName
	if model, exists := s.table[sName]; exists {
		sName = getTableNameFromStruct(model)
	}

	sTime := time.Now().UTC()
	sNameWithTime := fmt.Sprintf("%v%v", sName, sTime.Unix())

	log := sLog{
		StructName:  sNameWithTime,
		TableName:   tName,
		Data:        schema,
		Date:        sTime.Format("Mon _2 Jan 2006 15:04:05"),
		importedPkg: make(map[string]struct{}),
	}

	if logStr, err = execLogTmpl(log); err == nil {
		if logDir, err = s.makeStructLogDir(sName); err == nil {
			file := logDir + "/" + sNameWithTime + ".go"
			if fp, err = os.Create(file); err == nil {
				fp.WriteString(getWarning())
				fp.WriteString("package " + sName + "\n")

				if len(log.importedPkg) > 0 {
					fp.WriteString("import (" + getImportPkg(log.importedPkg) + ")")
				}

				fp.WriteString(logStr)
				exec.Command("gofmt", "-w", file).Run()
			}
		}
	}
	if err != nil {
		err = errors.New("Log Creation Error: " + err.Error())
	}
	return
}

//getImportPkg will append all imported packeges
func getImportPkg(pkg map[string]struct{}) (impPkg string) {
	for k := range pkg {
		impPkg += "\"" + k + "\"\n"
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
	dType += getStructConstraintSQL(schema)
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
	if sType == timePkgTime {
		l.importedPkg[timePkg] = struct{}{}
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
//{{ .StructName }} : {{ .TableName }} table model [As on {{ .Date }} UTC]
type {{ .StructName }} struct {
	tableName struct{} ` + "`sql:\"{{ .TableName }}\"`" + `
{{- range $key, $value := .Data}}
	{{ if eq $value.StructColumnName "" -}}
		{{ Title $value.ColumnName -}}
	{{ else -}}
		{{ $value.StructColumnName -}}
	{{ end -}}
	{{print " "}} {{ $.GetStructFieldType $value.DataType }}` + " `sql:\"{{ .ColumnName }},type:{{ getSQLTag $value }}\"`" + `
{{- end }}
}`
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
----- DO NOT EDIT THIS----
*/

`
}
