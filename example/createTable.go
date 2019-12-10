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
		err := s.CreateTable(conn, &db.TestAddress{})
		logIfError(err)

		//2.
		s.SetTableModel(&db.TestUser{})
		err = s.CreateTable(conn, "test_user")
		logIfError(err)
	}
}

func logIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}
