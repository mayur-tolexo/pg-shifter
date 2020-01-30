package shifter

import (
	"testing"
	"time"

	"github.com/mayur-tolexo/contour/adapter/psql"
	"github.com/stretchr/testify/assert"
)

func TestLocalAlterTable(t *testing.T) {
	if tx, err := psql.Tx(); err == nil {
		s := NewShifter()
		assert := assert.New(t)

		//invalid table
		err = s.alterTable(tx, "invalid_table", true)
		assert.Error(err)

		tName := "local_user"
		s.SetTableModel(&localUser{})
		err = s.createTable(tx, tName, true)
		assert.NoError(err)
		s.SetTableModel(&localUserColRemoved{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)
		err = s.dropTable(tx, tName, true)
		assert.NoError(err)
	}
}

type localUser struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `json:"user_id" sql:"user_id,type:serial PRIMARY KEY"`
	CreatedAt time.Time `json:"-" sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `json:"-" sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserColRemoved struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `json:"user_id" sql:"user_id,type:serial PRIMARY KEY"`
	CreatedAt time.Time `json:"-" sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserColAdded struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `json:"user_id" sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `json:"-" sql:"password,type:varchar(255) NULL"`
	CreatedAt time.Time `json:"-" sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `json:"-" sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}
