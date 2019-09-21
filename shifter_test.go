package shifter

import (
	"fmt"
	"testing"

	"github.com/mayur-tolexo/contour/adapter/psql"
)

//TestAddress Table structure as in DB
type TestAddress struct {
	tableName struct{} `sql:"test_address"`
	AddressID int      `json:"address_id,omitempty" sql:"address_id,type:serial PRIMARY KEY"`
	Address   string   `json:"address" sql:"address,type:text NOT NULL DEFAULT ''"`
	Type      string   `json:"type" sql:"type,type:address_type NOT NULL DEFAULT 'billing'"`
	Landmark  string   `json:"landmark" default:"null" sql:"landmark,type:varchar(255)"`
	Pincode   string   `json:"pincode" sql:"pincode,type:varchar(20)"`
	City      string   `json:"city" sql:"city,type:varchar(255)"`
}

func TestCreateAllTable(t *testing.T) {

	conn, err := psql.Conn(true)
	if err == nil {
		// psql.StartLogging = true
		s := NewShifter()
		s.SetTableModel(&TestAddress{})
		err = s.CreateAllTable(conn)
		if err != nil {
			fmt.Println(err)
		}
	}
}
