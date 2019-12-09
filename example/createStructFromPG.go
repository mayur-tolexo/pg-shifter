package main

import (
	"github.com/mayur-tolexo/contour/adapter/psql"
	shifter "github.com/mayur-tolexo/pg-shifter"
)

func main() {
	if conn, err := psql.Conn(true); err == nil {
		//this will create the test_address go struct in filepath
		//as filepath is not given so it will be created in
		//pwd/log/TestAddress/TestAddress.go
		shifter.NewShifter().CreateStruct(conn, "test_address", "")
	}
}
