package shifter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-pg/pg"
	"github.com/mayur-tolexo/pg-shifter/util"
)

//Create trigger
func (s *Shifter) createTrigger(tx *pg.Tx, tableName string) (err error) {
	if s.isSkip(tableName) == false {
		defer s.logMode(false)
		trigger := s.GetTrigger(tableName)
		s.logMode(s.verbose)
		if _, err = tx.Exec(trigger); err != nil {
			err = getWrapError(tableName, "create trigger", trigger, err)
		}
	}
	return
}

//GetTrigger : Get triggers by table name
func (s *Shifter) GetTrigger(tableName string) (trigger string) {
	var (
		// aInsertTrigger string
		bUpdateTrigger string
		aUpdateTrigger string
		// aDeleteTrigger string
	)
	// aInsertTrigger := getInsertTrigger(tableName)
	bUpdateTrigger, aUpdateTrigger = s.getUpdateTrigger(tableName)
	// aDeleteTrigger = s.getDeleteTrigger(tableName)

	for _, curTag := range s.getTableTriggersTag(tableName) {
		switch curTag {
		case afterInsertTrigger:
			trigger += s.getInsertTrigger(tableName)
		case afterUpdateTrigger:
			trigger += aUpdateTrigger
		case afterDeleteTrigger:
			trigger += s.getDeleteTrigger(tableName)
		case beforeUpdateTrigger:
			trigger += bUpdateTrigger
		}
	}

	// trigger = aInsertTrigger + bUpdateTrigger + aUpdateTrigger + aDeleteTrigger
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
	delimiter := `
	------------------------- AFTER INSERT TRIGGER -------------------------`

	fnQuery := fmt.Sprintf(delimiter+`
	CREATE OR REPLACE FUNCTION %v()
	RETURNS trigger AS
	$$
    	BEGIN
        	INSERT INTO %v (
        		%v
        	) VALUES (
        		%v
        	);
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
	EXECUTE PROCEDURE %v();`+delimiter,
		afterInsertTable, tableName, afterInsertTable, tableName, afterInsertTable)
	aInsertTrigger = fnQuery + triggerQuery + "\n"
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
	delimiter := `
	------------------------- AFTER UPDATE TRIGGER -------------------------`

	fnQuery := fmt.Sprintf(delimiter+`
	CREATE OR REPLACE FUNCTION %v()
	RETURNS trigger AS
	$$
		BEGIN
			IF %v
			THEN
        	INSERT INTO %v (
        		%v
        	) VALUES(
        		%v
        	);
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
	EXECUTE PROCEDURE %v();`+delimiter,
		afterUpdateTable, tableName, afterUpdateTable, tableName, afterUpdateTable)
	aUpdateTrigger = fnQuery + triggerQuery + "\n"
	return
}

//Get before update trigger function and trigger by table name
func (s *Shifter) getBeforeUpdateTrigger(tableName string) (bUpdateTrigger string) {

	beforeUpdateTable := util.GetBeforeInsertTriggerName(tableName)
	delimiter := `
	------------------------- BEFORE UPDATE TRIGGER -------------------------`

	fnQuery := fmt.Sprintf(delimiter+`
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
	EXECUTE PROCEDURE %v();`+delimiter,
		beforeUpdateTable, tableName, beforeUpdateTable, tableName, beforeUpdateTable)
	bUpdateTrigger = fnQuery + triggerQuery + "\n"
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
	delimiter := `
	------------------------- AFTER DELETE TRIGGER -------------------------`

	fnQuery := fmt.Sprintf(delimiter+`
	CREATE OR REPLACE FUNCTION %v()
	RETURNS trigger AS
	$$
	BEGIN
		INSERT INTO %v (
			%v
		) VALUES (
			%v
		);
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
	EXECUTE PROCEDURE %v();`+delimiter,
		afterDeleteTable, tableName, afterDeleteTable, tableName, afterDeleteTable)
	aDeleteTrigger = fnQuery + triggerQuery + "\n"
	return
}

//Get history table fields from struct model of database
func (s *Shifter) getHistoryFields(dbModel interface{}, dataTag, action string) (
	fields string, values string, updateCondition string, updatedAt bool, err error) {

	fieldMap := util.GetStructField(dbModel)
	fCount, uCount := 0, 0
	for _, inputField := range fieldMap {
		if tagValue, exists := inputField.Tag.Lookup("sql"); exists == true {

			curField := strings.Split(tagValue, ",")
			updatedAtExists := strings.Contains(curField[0], "updated_at")

			if len(curField) > 0 && updatedAtExists == false {
				fCount++
				fields += curField[0] + "," + getNewline(fCount)
				if curField[0] == "created_at" {
					values += "NOW()," + getNewline(fCount)
				} else {
					uCount++
					values += dataTag + "." + curField[0] + "," + getNewline(fCount)
					updateCondition += " OLD." + curField[0] + " <> NEW." + curField[0] + " OR" + getNewline(uCount)
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
	updateCondition = strings.TrimSuffix(updateCondition, "OR"+getNewline(uCount))
	return
}

func getNewline(count int) (sep string) {
	if count%4 == 0 {
		sep = "\n\t\t\t"
	}
	return
}
