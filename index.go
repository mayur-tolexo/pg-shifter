package shifter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/pg-shifter/model"
	"github.com/mayur-tolexo/pg-shifter/util"
)

//index type const
const (
	BtreeIndex  = "btree"   //btree index type
	GinIndex    = "gin"     //gin index type
	GistIndex   = "gist"    //gist index type
	HashIndex   = "hash"    //hash index type
	BrinIndex   = "brin"    //brin index type
	SPGistIndex = "sp-gist" //sp-gist index type
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
				err = getWrapError(tableName, "create index", indexSQL, err)
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

//getDBIndex : Get index of table from database
func getDBIndex(tx *pg.Tx, tableName string) (idx []model.Index, err error) {
	query := `
	with idx as (
		select
		--    t.relname as table_name
		    i.relname as index_name
		    , c.column_name
		    , am.amname
		    , array_position(ix.indkey::int[],c.ordinal_position::int) as position
		from
			pg_index ix
			join pg_class t on  t.oid = ix.indrelid
		    join pg_class i on i.oid = ix.indexrelid
		    JOIN pg_am am ON am.oid = i.relam
		    join unnest(ix.indkey::int[]) as colNo on true 
		    join information_schema.columns as c 
			on c.ordinal_position = colNo and c.table_name = t.relname
		where
		    t.relkind = 'r'
		    and ix.indisunique = false
		    and t.relname = ?
		   order by i.relname, position
	)
	select index_name 
	, string_agg(distinct amname,',') as itype
	, string_agg(column_name,',') as col
	from idx
	group by index_name;`
	_, err = tx.Query(&idx, query, tableName)
	return
}
