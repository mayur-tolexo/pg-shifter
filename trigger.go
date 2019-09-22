package shifter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/flaw"
	util "github.com/mayur-tolexo/pg-shifter/util.go"
)

//Create trigger
func (s *Shifter) createTrigger(tx *pg.Tx, tableName string) (err error) {
	if s.IsSkip(tableName) == false {
		trigger := s.GetTrigger(tableName)
		// fmt.Println(trigger)
		if _, err = tx.Exec(trigger); err != nil {
			msg := fmt.Sprintf("Table: %v", tableName)
			err = flaw.ExecError(err, msg)
			fmt.Println("Trigger Error:", msg, err)
		}
	}
	return
}

//GetTrigger : Get triggers by table name
func (s *Shifter) GetTrigger(tableName string) (trigger string) {
	var (
		aInsertTrigger string
		bUpdateTrigger string
		aUpdateTrigger string
		aDeleteTrigger string
	)
	// aInsertTrigger := getInsertTrigger(tableName)
	bUpdateTrigger, aUpdateTrigger = s.getUpdateTrigger(tableName)
	aDeleteTrigger = s.getDeleteTrigger(tableName)
	trigger = aInsertTrigger + bUpdateTrigger + aUpdateTrigger + aDeleteTrigger
	return
}

//Get after insert trigger
func (s *Shifter) getInsertTrigger(tableName string) (aInsertTrigger string) {
	if dbModel, valid := s.table[tableName]; valid == true {
		if fields, values, _, _, err := s.getHistoryFields(dbModel, "NEW", "insert"); err == nil {
			aInsertTrigger = s.getAfterInsertTrigger(tableName, fields, values)
		} else {
			fmt.Println("getInsertTrigger: ", err.Error())
		}
	}
	return
}

//Get after insert trigger function and trigger by table name
func (s *Shifter) getAfterInsertTrigger(tableName, fields, values string) (
	aInsertTrigger string) {

	historyTable := util.GetHistoryTableName(tableName)
	afterInsertTable := util.GetAfterInsertTriggerName(tableName)
	fnQuery := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %v()
		RETURNS trigger AS
		$$
	    	BEGIN
	        	INSERT INTO %v (%v) 
	        	VALUES(%v);
	        	RETURN NEW;
	    	END;
		$$
		LANGUAGE 'plpgsql';
		`, afterInsertTable, historyTable, fields, values)
	triggerQuery := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %v ON %v;
		CREATE TRIGGER %v
		AFTER INSERT ON %v 
		FOR EACH ROW
		EXECUTE PROCEDURE %v();
		`, afterInsertTable, tableName, afterInsertTable, tableName, afterInsertTable)
	aInsertTrigger = fnQuery + triggerQuery
	return
}

//Get before and after update triggers
func (s *Shifter) getUpdateTrigger(tableName string) (bUpdateTrigger, aUpdateTrigger string) {
	if dbModel, valid := s.table[tableName]; valid == true {
		if fields, values, updateCondition, updatedAt, err :=
			s.getHistoryFields(dbModel, "OLD", "update"); err == nil {
			if aUpdateTrigger = s.getAfterUpdateTrigger(tableName, fields,
				values, updateCondition); updatedAt == true {
				bUpdateTrigger = s.getBeforeUpdateTrigger(tableName)
			}
		} else {
			fmt.Println("getUpdateTrigger: ", err.Error())
		}
	}
	return
}

//Get after update trigger function and trigger by table name
func (s *Shifter) getAfterUpdateTrigger(tableName, fields, values,
	updateCondition string) (aUpdateTrigger string) {

	historyTable := util.GetHistoryTableName(tableName)
	afterUpdateTable := util.GetAfterUpdateTriggerName(tableName)

	fnQuery := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %v()
		RETURNS trigger AS
		$$
	    	BEGIN
	    		IF %v THEN
		        	INSERT INTO %v (%v) 
		        	VALUES(%v);
	        	END IF;
	        	RETURN NEW;
	    	END;
		$$
		LANGUAGE 'plpgsql';
		`, afterUpdateTable, updateCondition, historyTable, fields, values)
	triggerQuery := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %v ON %v;
		CREATE TRIGGER %v
		AFTER UPDATE ON %v 
		FOR EACH ROW
		EXECUTE PROCEDURE %v();
		`, afterUpdateTable, tableName, afterUpdateTable, tableName, afterUpdateTable)
	aUpdateTrigger = fnQuery + triggerQuery
	return
}

//Get before update trigger function and trigger by table name
func (s *Shifter) getBeforeUpdateTrigger(tableName string) (bUpdateTrigger string) {

	beforeUpdateTable := util.GetBeforeInsertTriggerName(tableName)

	fnQuery := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %v()
		RETURNS trigger AS
		$$
	    	BEGIN
	        	NEW.updated_at = now();
	        	RETURN NEW;
	    	END;
		$$
		LANGUAGE 'plpgsql';
		`, beforeUpdateTable)
	triggerQuery := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %v ON %v;
		CREATE TRIGGER %v
		BEFORE UPDATE ON %v 
		FOR EACH ROW
		EXECUTE PROCEDURE %v();
		`, beforeUpdateTable, tableName, beforeUpdateTable, tableName, beforeUpdateTable)
	bUpdateTrigger = fnQuery + triggerQuery
	return
}

//Get after delete trigger
func (s *Shifter) getDeleteTrigger(tableName string) (aDeleteTrigger string) {
	if dbModel, valid := s.table[tableName]; valid == true {
		if fields, values, _, _, err := s.getHistoryFields(dbModel, "OLD", "delete"); err == nil {
			aDeleteTrigger = s.getAfterDeleteTrigger(tableName, fields, values)
		} else {
			fmt.Println("getDeleteTrigger: ", err.Error())
		}
	}
	return
}

//Get after delete trigger function and trigger by table name
func (s *Shifter) getAfterDeleteTrigger(tableName, fields, values string) (aDeleteTrigger string) {

	historyTable := util.GetHistoryTableName(tableName)
	afterDeleteTable := util.GetAfterDeleteTriggerName(tableName)

	fnQuery := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %v()
		RETURNS trigger AS
		$$
	    	BEGIN
	        	INSERT INTO %v (%v) 
	        	VALUES(%v);
	        	RETURN OLD;
	    	END;
		$$
		LANGUAGE 'plpgsql';
		`, afterDeleteTable, historyTable, fields, values)
	triggerQuery := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %v ON %v;
		CREATE TRIGGER %v
		AFTER DELETE ON %v 
		FOR EACH ROW
		EXECUTE PROCEDURE %v();
		`, afterDeleteTable, tableName, afterDeleteTable, tableName, afterDeleteTable)
	aDeleteTrigger = fnQuery + triggerQuery
	return
}

//Get history table fields from struct model of database
func (s *Shifter) getHistoryFields(dbModel interface{}, dataTag, action string) (
	fields string, values string, updateCondition string, updatedAt bool, err error) {
	fieldMap := util.GetStructField(dbModel)
	for _, inputField := range fieldMap {
		if tagValue, exists := inputField.Tag.Lookup("sql"); exists == true {
			curField := strings.Split(tagValue, ",")
			updatedAtExists := strings.Contains(curField[0], "updated_at")
			if len(curField) > 0 && updatedAtExists == false {
				fields += curField[0] + ","
				if curField[0] == "created_at" {
					values += "NOW(),"
				} else {
					values += dataTag + "." + curField[0] + ","
					updateCondition += " OLD." + curField[0] + " <> NEW." + curField[0] + " OR"
				}
			} else if updatedAtExists == true {
				updatedAt = true
			}
		} else if exists == false {
			err = errors.New("sql tag is missing in database struct model " + inputField.Name)
			return
		}
	}
	fields += "action"
	values += "'" + action + "'"
	updateCondition = strings.TrimSuffix(updateCondition, "OR")
	return
}
