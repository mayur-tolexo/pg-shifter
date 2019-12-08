package shifter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/flaw"
	"github.com/mayur-tolexo/pg-shifter/util"
)

//index type const
const (
	GinIndex   = "gin"
	BtreeIndex = "btree"
)

//Create index of given table
func (s *Shifter) createIndex(tx *pg.Tx, tableName string, skipPrompt bool) (err error) {
	var indexSQL string
	for index, idxType := range s.GetIndex(tableName) {
		indexSQL += getIndexQuery(tableName, idxType, index)
	}
	if indexSQL != "" {
		choice := util.GetChoice("INDEX:\n"+indexSQL, skipPrompt)
		if choice == util.Yes {
			if _, err = tx.Exec(indexSQL); err != nil {
				msg := fmt.Sprintf("Table: %v Index: %v", tableName)
				err = flaw.CreateError(err, msg)
				fmt.Println("Index Error:", msg, err.Error())
			}
		}
	}
	return
}

//Get index query by tablename and table columns
func getIndexQuery(tableName string, indexDS string, column string) (uniqueKeyQuery string) {
	if strings.HasPrefix(indexDS, GinIndex) == true {
		indexDS = GinIndex
	} else {
		indexDS = BtreeIndex
	}

	constraintName := fmt.Sprintf("idx_%v_%v", tableName, strings.Replace(strings.Replace(column, " ", "", -1), ",", "_", -1))
	constraintName = util.GetStrByLen(constraintName, 64)
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %v ON %v USING %v (%v);\n",
		constraintName, tableName, indexDS, column)
}

//GetIndex will return index fields of struct
func (s *Shifter) GetIndex(tableName string) (idx map[string]string) {
	dbModel := s.table[tableName]
	refObj := reflect.ValueOf(dbModel)
	m := refObj.MethodByName("Index")
	if m.IsValid() {
		out := m.Call([]reflect.Value{})
		if len(out) > 0 && out[0].Kind() == reflect.Map {
			idx, _ = out[0].Interface().(map[string]string)
		}
	}
	return
}
