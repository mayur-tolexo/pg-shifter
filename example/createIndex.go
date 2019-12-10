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

		//1. Create All Index by passing table struct
		//as skip prompt is true hence it will execute without confirmation
		err = s.CreateAllIndex(conn, &db.TestAddress{}, true)
		logIfError(err)

		//2. Create All Index by passing table name
		s.SetTableModel(&db.TestAddress{})
		err = s.CreateAllIndex(conn, "test_address")
		logIfError(err)
	}
}

func logIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}
