package main

import (
	"log"
	"time"

	"github.com/mayur-tolexo/contour/adapter/psql"
	shifter "github.com/mayur-tolexo/pg-shifter"
)

func main() {
	if conn, err := psql.Conn(true); err == nil {
		s := shifter.NewShifter()

		//1. Create All Index by passing table struct
		//as skip prompt is true hence it will execute without confirmation
		err = s.CreateAllIndex(conn, &TestAddress{}, true)
		logIfError(err)

		//2. Create All Index by passing table name
		s.SetTableModel(&TestAddress{})
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

//UniqueKey of the table. This is for composite unique keys
func (TestAddress) UniqueKey() []string {
	uk := []string{
		"address_id,status,city",
	}
	return uk
}

//Index of the table. For composite index use ,
//Default index type is btree. For gin index use gin
func (TestAddress) Index() map[string]string {
	idx := map[string]string{
		"status":            shifter.BtreeIndex,
		"address_id,status": shifter.BtreeIndex,
	}
	return idx
}

//Enum of the table.
func (TestAddress) Enum() map[string][]string {
	enm := map[string][]string{
		"address_status": {"enable", "disable"},
	}
	return enm
}

func logIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}
