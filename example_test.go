package shifter

import (
	"log"
	"time"

	"github.com/mayur-tolexo/contour/adapter/psql"
)

func ExampleCreateTable() {
	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()

		//1. Create table by table struct
		err := s.CreateTable(conn, &TestAddress{})
		logIfError(err)

		//2. Create table by table name
		err = s.SetTableModel(&TestAddress{})
		logIfError(err)
		err = s.CreateTable(conn, "test_address")
		logIfError(err)
	}
}

func ExampleCreateAllIndex() {
	if conn, err := psql.Conn(true); err == nil {
		s := NewShifter()

		//1. Create All Index by passing table struct
		//as skip prompt is true hence it will execute without confirmation
		err = s.CreateAllIndex(conn, &TestAddress{}, true)
		logIfError(err)

		//2. Create All Index by passing table name
		err = s.SetTableModel(&TestAddress{})
		logIfError(err)
		err = s.CreateAllIndex(conn, "test_address")
		logIfError(err)
	}
}

//TestAddress Table structure as in DB
type TestAddress struct {
	tableName struct{}  `sql:"test_address"`
	AddressID int       `json:"address_id,omitempty" sql:"address_id,type:serial PRIMARY KEY"`
	City      string    `json:"city" sql:"city,type:varchar(25) UNIQUE"`
	Status    string    `json:"status,omitempty" sql:"status,type:address_status"`
	CreatedBy int       `json:"created_by" sql:"created_by,type:int NOT NULL UNIQUE REFERENCES test_user(user_id)"`
	CreatedAt time.Time `json:"-" sql:"created_at,type:time NOT NULL DEFAULT NOW()"`
	UpdatedAt time.Time `json:"-" sql:"updated_at,type:timetz NOT NULL DEFAULT NOW()"`
}

//Index of the table. For composite index use ,
//Default index type is btree. For gin index use gin
func (TestAddress) Index() map[string]string {
	idx := map[string]string{
		"status":            "",
		"address_id,status": "",
	}
	return idx
}

func logIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}
