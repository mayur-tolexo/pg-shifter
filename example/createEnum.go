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

		//1. Create enum by passing table model
		err = s.CreateEnum(conn, &TestAddress{}, "address_status")
		logIfError(err)

		//2. Create enum by passing table name
		err = s.SetTableModel(&TestAddress{})
		logIfError(err)
		err = s.CreateEnum(conn, "test_address", "address_status")
		logIfError(err)

		//1. Create all enum by passing table model
		err = s.CreateAllEnum(conn, &TestAddress{})
		logIfError(err)

		//2. Create all enum by pasing table name
		err = s.SetTableModel(&TestAddress{})
		logIfError(err)
		err = s.CreateAllEnum(conn, "test_address")
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
