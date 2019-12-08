package shifter

import (
	"testing"

	"github.com/mayur-tolexo/contour/adapter/psql"
	"github.com/mayur-tolexo/pg-shifter/db"
	"github.com/stretchr/testify/assert"
)

func TestAlterTable(t *testing.T) {
	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()
		s.Verbose = true
		s.SetTableModel(&db.TestAddress{})

		assert := assert.New(t)
		err = s.AlterAllTable(conn, true)
		assert.NoError(err)
	}
}
