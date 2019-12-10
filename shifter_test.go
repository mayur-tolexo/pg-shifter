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
		err = s.CreateStruct(conn, "test_user", filePath)
		assert := assert.New(t)
		assert.NoError(err)
	}
}

func TestCreateTable(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		err := s.CreateTable(conn, &db.TestAddress{})
		assert := assert.New(t)
		assert.NoError(err)

		s.SetTableModel(&db.TestUser{})
		err = s.CreateTable(conn, "test_user")
		assert.NoError(err)
	}
}

func TestCreateEnum(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		err := s.CreateEnum(conn, &db.TestAddress{}, "address_status")
		assert := assert.New(t)
		assert.NoError(err)

		s.SetTableModel(&db.TestUser{})
		err = s.CreateEnum(conn, "test_user", "yesno_type")
		assert.NoError(err)
	}
}

func TestCreateAllEnum(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		err := s.CreateAllEnum(conn, &db.TestAddress{})
		assert := assert.New(t)
		assert.NoError(err)

		s.SetTableModel(&db.TestUser{})
		err = s.CreateAllEnum(conn, "test_user")
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
		s.Verbose(true)
		addAllTables(s)
		assert := assert.New(t)
		err = s.AlterAllTable(conn, true)
		assert.NoError(err)
	}
}

func TestCreateTrigger(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		// s.Verbose(true)
		addAllTables(s)
		err = s.CreateTrigger(conn, "test_user")
		assert := assert.New(t)
		assert.NoError(err)
	}
}
