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

		//drop column
		s.SetTableModel(&localUserColRemoved{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		//add column
		s.SetTableModel(&localUser{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		//remove column nullable
		s.SetTableModel(&localUserNullRemoved{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		//add column nullable
		s.SetTableModel(&localUser{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		//remove column default
		s.SetTableModel(&localUserDefaultRemoved{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		//add column default
		s.SetTableModel(&localUser{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		err = s.dropTable(tx, tName, true)
		assert.NoError(err)
	}
}

type localUser struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL DEFAULT NULL"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserColRemoved struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL DEFAULT NULL"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserNullRemoved struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserDefaultRemoved struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}
