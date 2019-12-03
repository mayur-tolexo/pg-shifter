package shifter

import (
	"testing"

	"github.com/mayur-tolexo/contour/adapter/psql"
)

func TestAlterTable(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		tx, _ := conn.Begin()
		s := NewShifter()
		s.SetTableModel(&TestAddress{})
		s.alterTable(tx, "test_address", true)
		tx.Commit()
	}
}
