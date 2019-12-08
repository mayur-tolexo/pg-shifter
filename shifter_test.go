package shifter

import (
	"testing"

	"github.com/mayur-tolexo/contour/adapter/psql"
	"github.com/mayur-tolexo/pg-shifter/db"
	"github.com/stretchr/testify/assert"
)

func addAllTables(s *Shifter) {
	s.SetTableModel(&db.TestAddress{})
	s.SetTableModel(&db.TestAdminUser{})
	s.SetTableModel(&db.TestUser{})
}

func TestCreateStruct(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		filePath := ""
		err = s.CreateStruct(conn, "test_address", filePath)
		assert := assert.New(t)
		assert.NoError(err)
	}
}

func TestCreateAllTable(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		addAllTables(s)
		assert := assert.New(t)
		err = s.CreateAllTable(conn)
		assert.NoError(err)
	}
}

func TestAlterTable(t *testing.T) {
	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		s.Verbose = true
		addAllTables(s)
		assert := assert.New(t)
		err = s.AlterAllTable(conn, true)
		assert.NoError(err)
	}
}
