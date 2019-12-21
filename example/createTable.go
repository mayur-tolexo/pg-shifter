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

		//1. Create table by table struct
		err := s.CreateTable(conn, &TestUser{})
		logIfError(err)

		//2. Create table by table name
		err = s.SetTableModel(&TestUser{})
		logIfError(err)
		err = s.CreateTable(conn, "test_user")
		logIfError(err)
	}
}

//TestUser Table structure as in DB
type TestUser struct {
	tableName        struct{}               `sql:"test_user"`
	UserID           int                    `json:"user_id" sql:"user_id,type:serial PRIMARY KEY"`
	Username         string                 `json:"username" sql:"username,type:varchar(255)"`
	Password         string                 `json:"-" sql:"password,type:varchar(255) NULL"`
	PassSalt         string                 `json:"-" sql:"pass_salt,type:varchar(255) NULL"`
	CountryID        int                    `json:"country_id" sql:"fk_country_id,type:int NULL REFERENCES country(country_id) ON DELETE RESTRICT ON UPDATE CASCADE"`
	Email            string                 `json:"email" sql:"email,type:varchar(255) UNIQUE"`
	Name             string                 `json:"name" sql:"name,type:varchar(255)"`
	AltContactNo     string                 `json:"alt_contact_no" default:"true" sql:"alt_contact_no,type:varchar(20)"`
	AltPhoneCode     string                 `json:"alt_phonecode" default:"true" sql:"alt_phonecode,type:varchar(20)"`
	Landline         string                 `json:"landline" default:"null" sql:"landline,type:text NULL DEFAULT NULL"`
	Department       string                 `json:"department" default:"null" sql:"department,type:varchar(100)"`
	Designation      string                 `json:"designation" default:"null" sql:"designation,type:varchar(100)"`
	EmailVerified    string                 `json:"email_verified,omitempty" sql:"email_verified,type:yesno_type NOT NULL DEFAULT 'no'"`
	PhoneVerified    string                 `json:"phone_verified,omitempty" sql:"phone_verified,type:yesno_type NOT NULL DEFAULT 'no'"`
	WhatsappVerified string                 `json:"whatsapp_verified,omitempty" sql:"whatsapp_verified,type:yesno_type NOT NULL DEFAULT 'no'"`
	Attribute        map[string]interface{} `json:"attribute,omitempty" sql:"attribute,type:jsonb NOT NULL DEFAULT '{}'::jsonb"`
	LastLogin        time.Time              `json:"last_login" sql:"last_login,type:timestamp"`
	CreatedBy        int                    `json:"-" sql:"created_by,type:int NOT NULL REFERENCES pra_user(user_id) ON DELETE RESTRICT ON UPDATE CASCADE DEFERRABLE INITIALLY DEFERRED"`
	CreatedAt        time.Time              `json:"-" sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt        time.Time              `json:"-" sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

func logIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}
