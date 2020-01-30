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

		//col unique added
		s.SetTableModel(&localUserUnqCol{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		//col unique removed
		s.SetTableModel(&localUser{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		//col default foreign key added
		s.SetTableModel(&localUserDefaultFkCol{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		//col cascade foreign key altered
		s.SetTableModel(&localUserCascadeFkCol{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		//col restrict foreign key altered
		s.SetTableModel(&localUserRestrictFkCol{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		//col no action foreign key altered
		s.SetTableModel(&localUserNoActionFkCol{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		//col set null foreign key altered
		s.SetTableModel(&localUserSetNullFkCol{})
		err = s.alterTable(tx, tName, true)
		assert.NoError(err)

		// //unique foreign key alter
		// s.SetTableModel(&localUserUniqueDefaultFkCol{})
		// err = s.alterTable(tx, tName, true)
		// assert.NoError(err)

		// //unique removed from foreign key - alter table
		// s.SetTableModel(&localUserDefaultFkCol{})
		// err = s.alterTable(tx, tName, true)
		// assert.NoError(err)

		// //col set default foreign key altered
		// s.SetTableModel(&localUserSetDefaultFkCol{})
		// err = s.alterTable(tx, tName, true)
		// assert.NoError(err)

		//col default foreign key removed
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
	CreatedBy int       `sql:"created_by,type:int"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserColRemoved struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	CreatedBy int       `sql:"created_by,type:int"`
	Password  string    `sql:"password,type:varchar(255) NULL DEFAULT NULL"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserNullRemoved struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	CreatedBy int       `sql:"created_by,type:int"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserDefaultRemoved struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL"`
	CreatedBy int       `sql:"created_by,type:int"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserUnqCol struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL DEFAULT NULL UNIQUE"`
	CreatedBy int       `sql:"created_by,type:int"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserDefaultFkCol struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL DEFAULT NULL UNIQUE"`
	CreatedBy int       `sql:"created_by,type:int  NOT NULL REFERENCES local_user(user_id)"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserCascadeFkCol struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL DEFAULT NULL UNIQUE"`
	CreatedBy int       `sql:"created_by,type:int NOT NULL REFERENCES local_user(user_id) ON DELETE CASCADE ON UPDATE CASCADE"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserRestrictFkCol struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL DEFAULT NULL UNIQUE"`
	CreatedBy int       `sql:"created_by,type:int NOT NULL REFERENCES local_user(user_id)  ON DELETE RESTRICT ON UPDATE RESTRICT"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserNoActionFkCol struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL DEFAULT NULL UNIQUE"`
	CreatedBy int       `sql:"created_by,type:int NOT NULL REFERENCES local_user(user_id)  ON DELETE NO ACTION ON UPDATE NO ACTION"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserSetNullFkCol struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL DEFAULT NULL UNIQUE"`
	CreatedBy int       `sql:"created_by,type:int NOT NULL REFERENCES local_user(user_id)  ON DELETE SET NULL ON UPDATE SET NULL"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserSetDefaultFkCol struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL DEFAULT NULL UNIQUE"`
	CreatedBy int       `sql:"created_by,type:int NOT NULL REFERENCES local_user(user_id)  ON DELETE SET DEFAULT ON UPDATE SET DEFAULT"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

type localUserUniqueDefaultFkCol struct {
	tableName struct{}  `sql:"local_user"`
	UserID    int       `sql:"user_id,type:serial PRIMARY KEY"`
	Password  string    `sql:"password,type:varchar(255) NULL DEFAULT NULL UNIQUE"`
	CreatedBy int       `sql:"created_by,type:int NOT NULL UNIQUE REFERENCES local_user(user_id)"`
	CreatedAt time.Time `sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}
