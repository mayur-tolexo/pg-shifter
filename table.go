package shifter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/mayur-tolexo/flaw"
	"github.com/mayur-tolexo/pg-shifter/model"
	"github.com/mayur-tolexo/pg-shifter/util"
)

//Create Table in database
func (s *Shifter) createTable(tx *pg.Tx, tableName string, withDependency bool) (err error) {
	tableModel := s.table[tableName]
	if _, alreadyCreated := tableCreated[tableModel]; alreadyCreated == false {
		tableCreated[tableModel] = true
		err = s.upsertAllEnum(tx, tableName)
		if err == nil {
			if withDependency {
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
					if err = s.upsertAllEnum(tx, refTable); err == nil {
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

	exists := false
	if tableExists(tx, tableName) {
		exists = true
	}

	if exists == false {
		if err = tx.CreateTable(tableModel,
			&orm.CreateTableOptions{IfNotExists: true}); err == nil {

			if err = s.createHistory(tx, tableName); err == nil {
				if sql := s.getPostCreateSQLFromMethod(tableName); sql != "" {
					if _, err = tx.Exec(sql); err != nil {
						err = getWrapError(tableName, "Post Table Create SQL", sql, err)
					}
				}
			}

			if err == nil {
				fmt.Println("Table created: ", tableName)
			}
		} else {
			err = flaw.CreateError(err)
			fmt.Println("Table Error:", tableName, err.Error())
		}
	} else {
		fmt.Println("Table already exists: ", tableName)
	}
	return
}

//dropTable will drop table
func (s *Shifter) dropTable(tx *pg.Tx, tableName string, cascade bool) (err error) {
	var (
		log    sLog
		fData  []byte
		exists bool
	)
	if log, fData, exists, err = s.generateTableStructSchema(tx, tableName, true); err == nil &&
		exists {
		if err = execTableDrop(tx, tableName, cascade); err == nil {
			if err = s.dropHistory(tx, tableName, cascade); err == nil {
				err = s.logTableChange(log, fData)
			}
		}
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
		err = getWrapError(tableName, "drop table", sql, err)
	}
	return
}

//getPostCreateSQLFromMethod will return post table creation sql need to executed
//as defined in PostCreateSQL() method
func (s *Shifter) getPostCreateSQLFromMethod(tName string) (sql string) {
	dbModel := s.table[tName]
	refObj := reflect.ValueOf(dbModel)
	m := refObj.MethodByName("PostCreateSQL")
	if m.IsValid() {
		out := m.Call([]reflect.Value{})
		if len(out) > 0 && out[0].Kind() == reflect.String {
			sql = out[0].Interface().(string)
			strings.TrimSpace(sql)
		}
	}
	return
}

//getConstraint : Get Constraint of table from database
func getConstraint(tx *pg.Tx, tableName string) (constraint []model.ColSchema, err error) {
	query := `SELECT tc.constraint_type,
    tc.constraint_name, tc.is_deferrable, tc.initially_deferred, 
    kcu.column_name AS column_name, ccu.table_name AS foreign_table_name, 
    ccu.column_name AS foreign_column_name, pgc.confupdtype, pgc.confdeltype  
    FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu 
    ON tc.constraint_name = kcu.constraint_name 
    JOIN information_schema.constraint_column_usage AS ccu 
    ON ccu.constraint_name = tc.constraint_name 
    JOIN pg_constraint AS pgc ON pgc.conname = tc.constraint_name AND 
    conrelid=?::regclass::oid WHERE tc.constraint_type 
    IN('FOREIGN KEY','PRIMARY KEY','UNIQUE') AND tc.table_name = ?
    AND array_length(pgc.conkey,1) = 1;`
	if _, err = tx.Query(&constraint, query, tableName, tableName); err != nil {
		err = getWrapError(tableName, "table constraint", query, err)
	}
	return
}

//getColumnSchema : Get Column Schema of given table
func getColumnSchema(tx *pg.Tx, tableName string) (columnSchema []model.ColSchema, err error) {
	query := `SELECT col.column_name, col.column_default, col.data_type,
	col.ordinal_position as position,
	col.udt_name, col.is_nullable, col.character_maximum_length 
	, sq.sequence_name AS seq_name
	, sq.data_type AS seq_data_type
	FROM information_schema.columns col
	left join information_schema.sequences sq
	ON concat(sq.sequence_schema,'.',sq.sequence_name) = pg_get_serial_sequence(table_name, column_name)
	WHERE col.table_name = ?;`
	if _, err = tx.Query(&columnSchema, query, tableName); err != nil {
		err = getWrapError(tableName, "column schema", query, err)
	}
	return
}

//tableExists : Check if table exists in database
func tableExists(tx *pg.Tx, tableName string) (flag bool) {
	var num int
	sql := `SELECT 1 FROM pg_tables WHERE tablename = ?;`
	if _, err := tx.Query(pg.Scan(&num), sql, tableName); err != nil {
		fmt.Println("Table exists check error", err)
	} else if num == 1 {
		flag = true
	}
	return
}
