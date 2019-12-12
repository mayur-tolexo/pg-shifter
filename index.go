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
	BtreeIndex  = "btree"
	GinIndex    = "gin"
	GistIndex   = "gist"
	HashIndex   = "hash"
	BrinIndex   = "brin"
	SPGistIndex = "sp-gist"
)

//Create index of given table
func (s *Shifter) createIndex(tx *pg.Tx, tableName string, skipPrompt bool) (err error) {
	var indexSQL string
	for index, idxType := range s.getIndexFromMethod(tableName) {
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
	indexDS = getIndexType(indexDS)
	constraintName := fmt.Sprintf("idx_%v_%v", tableName, strings.Replace(strings.Replace(column, " ", "", -1), ",", "_", -1))
	constraintName = util.GetStrByLen(constraintName, 64)
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %v ON %v USING %v (%v);\n",
		constraintName, tableName, indexDS, column)
}

//getIndexType will return index type to use
func getIndexType(iType string) (idxType string) {
	switch iType {
	case GinIndex:
		idxType = GinIndex
	case GistIndex:
		idxType = GistIndex
	case HashIndex:
		idxType = HashIndex
	case BrinIndex:
		idxType = BrinIndex
	case SPGistIndex:
		idxType = SPGistIndex
	default:
		idxType = BtreeIndex
	}
	return
}

//getIndexFromMethod will return index fields of struct from Index() method
func (s *Shifter) getIndexFromMethod(tableName string) (idx map[string]string) {
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
