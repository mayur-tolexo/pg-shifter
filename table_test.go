package shifter

import (
	"fmt"
	"testing"

	"github.com/mayur-tolexo/contour/adapter/psql"
)

func TestAlterTable(t *testing.T) {
	// psql.StartLogging = true
	if conn, err := psql.Conn(true); err == nil {
		tx, _ := conn.Begin()
		s := NewShifter()
		s.SetTableModel(&TestAddress{})
		if err = s.alterTable(tx, "test_address", false); err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
			fmt.Println(err)
		}
	}
}
