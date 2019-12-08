package shifter

import (
	"testing"

	"github.com/mayur-tolexo/contour/adapter/psql"
	"github.com/mayur-tolexo/pg-shifter/db"
	"github.com/stretchr/testify/assert"
)

func TestCreateAllTable(t *testing.T) {

	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()

		s.SetTableModel(&db.TestAddress{})
		s.SetTableModel(&db.TestAdminUser{})
		s.SetTableModel(&db.TestUser{})

		assert := assert.New(t)
		err = s.CreateAllTable(conn)
		assert.NoError(err)
	}
}
