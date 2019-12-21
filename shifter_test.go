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

func TestCreateAllTable(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		addAllTables(s)
		assert := assert.New(t)
		err = s.CreateAllTable(conn)
		assert.NoError(err)
	}
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

		err = s.SetTableModel(&db.TestUser{})
		assert.NoError(err)
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

		err = s.SetTableModel(&db.TestUser{})
		assert.NoError(err)
		err = s.CreateEnum(conn, "test_user", "user_yesno_type")
		assert.NoError(err)
	}
}

func TestCreateAllEnum(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		err := s.CreateAllEnum(conn, &db.TestAddress{})
		assert := assert.New(t)
		assert.NoError(err)

		err = s.SetTableModel(&db.TestUser{})
		assert.NoError(err)
		err = s.CreateAllEnum(conn, "test_user")
		assert.NoError(err)
	}
}

func TestUpsertEnum(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		err := s.UpsertEnum(conn, &db.TestAddress{}, "address_status")
		assert := assert.New(t)
		assert.NoError(err)

		err = s.SetTableModel(&db.TestUser{})
		assert.NoError(err)
		err = s.UpsertEnum(conn, "test_user", "user_yesno_type")
		assert.NoError(err)
	}
}

func TestUpsertAllEnum(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		err := s.UpsertAllEnum(conn, &db.TestAddress{})
		assert := assert.New(t)
		assert.NoError(err)

		err = s.SetTableModel(&db.TestUser{})
		assert.NoError(err)
		err = s.UpsertAllEnum(conn, "test_user")
		assert.NoError(err)
	}
}
func TestCreateAllIndex(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		err := s.CreateAllIndex(conn, &db.TestAddress{}, true)
		assert := assert.New(t)
		assert.NoError(err)

		err = s.SetTableModel(&db.TestUser{})
		assert.NoError(err)
		err = s.CreateAllIndex(conn, "test_user", true)
		assert.NoError(err)
	}
}

func TestCreateAllUniqueKey(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		err := s.CreateAllUniqueKey(conn, &db.TestAddress{}, true)
		assert := assert.New(t)
		assert.NoError(err)

		err = s.SetTableModel(&db.TestUser{})
		assert.NoError(err)
		err = s.CreateAllUniqueKey(conn, "test_user", true)
		assert.NoError(err)
	}
}

func TestUpsertAllUniqueKey(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		err := s.UpsertAllUniqueKey(conn, &db.TestAddress{}, true)
		assert := assert.New(t)
		assert.NoError(err)

		err = s.SetTableModel(&db.TestUser{})
		assert.NoError(err)
		err = s.UpsertAllUniqueKey(conn, "test_user", true)
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

func TestDropTable(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		err := s.DropTable(conn, &db.TestAddress{}, true)
		assert := assert.New(t)
		assert.NoError(err)

		err = s.DropTable(conn, "test_user", true)
		assert.NoError(err)
	}
}

func TestDropAllTable(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		addAllTables(s)
		err := s.DropAllTable(conn, true)
		assert := assert.New(t)
		assert.NoError(err)
	}
}

func TestDropAllEnum(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		assert := assert.New(t)
		s := NewShifter()
		err := s.DropAllEnum(conn, &db.TestAddress{}, true)
		assert.NoError(err)
		err = s.DropAllEnum(conn, &db.TestUser{}, true)
		assert.NoError(err)
	}
}
