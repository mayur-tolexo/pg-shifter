package db

import "time"

//TestUser Table structure as in DB
type TestUser struct {
	tableName        struct{}               `sql:"test_user"`
	UserID           int                    `json:"user_id" sql:"user_id,type:serial PRIMARY KEY"`
	Username         string                 `json:"username" sql:"username,type:varchar(255)"`
	Password         string                 `json:"-" sql:"password,type:varchar(255) NULL"`
	PassSalt         string                 `json:"-" sql:"pass_salt,type:varchar(255) NULL"`
	Email            string                 `json:"email" sql:"email,type:varchar(255) UNIQUE"`
	Name             string                 `json:"name" sql:"name,type:varchar(255)"`
	AltContactNo     string                 `json:"alt_contact_no" default:"true" sql:"alt_contact_no,type:varchar(20)"`
	AltPhoneCode     string                 `json:"alt_phonecode" default:"true" sql:"alt_phonecode,type:varchar(20)"`
	Landline         string                 `json:"landline" default:"null" sql:"landline,type:text NULL DEFAULT NULL"`
	Department       string                 `json:"department" default:"null" sql:"department,type:varchar(100)"`
	Designation      string                 `json:"designation" default:"null" sql:"designation,type:varchar(100)"`
	EmailVerified    string                 `json:"email_verified,omitempty" sql:"email_verified,type:user_yesno_type NOT NULL DEFAULT 'no'"`
	PhoneVerified    string                 `json:"phone_verified,omitempty" sql:"phone_verified,type:user_yesno_type NOT NULL DEFAULT 'no'"`
	WhatsappVerified string                 `json:"whatsapp_verified,omitempty" sql:"whatsapp_verified,type:user_yesno_type NOT NULL DEFAULT 'no'"`
	Attribute        map[string]interface{} `json:"attribute,omitempty" sql:"attribute,type:jsonb NOT NULL DEFAULT '{}'::jsonb"`
	Status           string                 `json:"status" sql:"status,type:test_user_status_type DEFAULT 'notverified'"`
	LastLogin        time.Time              `json:"last_login" sql:"last_login,type:timestamp"`
	CreatedBy        int                    `json:"-" sql:"created_by,type:int NOT NULL REFERENCES test_user(user_id) ON DELETE RESTRICT ON UPDATE CASCADE DEFERRABLE INITIALLY DEFERRED"`
	CreatedAt        time.Time              `json:"-" sql:"created_at,type:timestamp NOT NULL DEFAULT NOW()"`
	UpdatedAt        time.Time              `json:"-" sql:"updated_at,type:timestamp NOT NULL DEFAULT NOW()"`
}

//Index of the table. For composite index use ,
//Default index type is btree. For gin index use gin value
func (TestUser) Index() map[string]string {
	idx := map[string]string{
		"username": "",
		"email":    "",
		"name":     "",
		"status":   "",
	}
	return idx
}

//UniqueKey of the table. This is for composite unique keys
func (TestUser) UniqueKey() []string {
	uk := []string{
		"username,status",
	}
	return uk
}

//Enum of the table.
func (TestUser) Enum() map[string][]string {
	enm := map[string][]string{
		"test_user_status_type": {"enable", "disable", "notverified"},
		"user_yesno_type":       {"yes", "no"},
	}
	return enm
}
