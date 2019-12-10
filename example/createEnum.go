package main

import (
	"log"

	"github.com/mayur-tolexo/contour/adapter/psql"
	shifter "github.com/mayur-tolexo/pg-shifter"
	"github.com/mayur-tolexo/pg-shifter/db"
)

func main() {
	if conn, err := psql.Conn(true); err == nil {
		s := shifter.NewShifter()

		//Create Enum
		//1.
		err := s.CreateEnum(conn, &db.TestAddress{}, "address_status")
		logIfError(err)

		//2.
		s.SetTableModel(&db.TestAddress{})
		err = s.CreateEnum(conn, "test_address", "address_status")
		logIfError(err)

		//Create All Enum
		//1.
		err = s.CreateAllEnum(conn, &db.TestAddress{})
		logIfError(err)

		//2.
		s.SetTableModel(&db.TestAddress{})
		err = s.CreateAllEnum(conn, "test_address")
		logIfError(err)
	}
}

func logIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}
