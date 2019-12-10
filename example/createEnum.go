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

		//1.
		err := s.CreateEnum(conn, &db.TestAddress{}, "status")
		logIfError(err)

		//2.
		s.SetTableModel(&db.TestAddress{})
		err = s.CreateEnum(conn, "test_address", "status")
		logIfError(err)
	}
}

func logIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}
