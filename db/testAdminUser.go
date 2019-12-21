package db

import "time"

//TestAdminUser Table structure as in DB
type TestAdminUser struct {
	tableName   struct{}  `sql:"test_admin_user"`
	AdminUserID int       `json:"admin_user_id" sql:"admin_user_id,type:serial PRIMARY KEY"`
	FkUserID    int       `json:"user_id" sql:"fk_user_id,type:int NOT NULL REFERENCES test_user(user_id) ON DELETE RESTRICT ON UPDATE CASCADE"`
	CreatedBy   int       `json:"-" sql:"created_by,type:int NOT NULL REFERENCES test_user(user_id) ON DELETE RESTRICT ON UPDATE CASCADE"`
	CreatedAt   time.Time `json:"-" sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt   time.Time `json:"-" sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
	MyAdminID   int       `json:"-" sql:"mysql_admin_id,type:int NULL UNIQUE"`
}

//Index of the table. For composite index use ,
//Default index type is btree. For gin index use gin value
func (TestAdminUser) Index() map[string]string {
	idx := map[string]string{
		"fk_user_id": "",
	}
	return idx
}

//UniqueKey of the table. This is for composite unique keys
func (TestAdminUser) UniqueKey() []string {
	uk := []string{
		"fk_user_id",
	}
	return uk
}
