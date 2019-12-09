[![Godocs](https://img.shields.io/badge/golang-documentation-blue.svg)](https://www.godoc.org/github.com/mayur-tolexo/pg-shifter)
[![Go Report Card](https://goreportcard.com/badge/github.com/mayur-tolexo/pg-shifter)](https://goreportcard.com/report/github.com/mayur-tolexo/pg-shifter)
[![Open Source Helpers](https://www.codetriage.com/mayur-tolexo/sworker/badges/users.svg)](https://www.codetriage.com/mayur-tolexo/pg-shifter)
[![Release](https://img.shields.io/github/release/mayur-tolexo/sworker.svg?style=flat-square)](https://github.com/mayur-tolexo/pg-shifter/releases)

# pg-shifter
Golang struct to postgres table shifter.

### Features
- [Create go struct from postgresql table name](#create-go-struct-from-postgresql-table-name)
- [Create table from struct](#create-table-from-struct)
- [Create enum](#recovery)
- [Create history table with after update/delete triggers](#recovery)
- [Alter table](#recovery)
	- [Add New Column](#add-new-column)
	- [Remove existing column](#remove-existing-column)
	- [Modify existing column](#modify-existing-column)
		- [Modify datatype](#modify-datatype)
		- Modify data length (e.g. varchar(255) to varchar(100))
		- Add/Drop default value
		- Add/Drop Not Null Constraint
		- Add/Drop constraint (Unique/Foreign Key)
		- [Modify constraint](#modify-constraint)
			- Set constraint deferrable
				- Initially deferred
				- Initially immediate
			- Set constraint not deferrable
			- Add/Drop **ON DELETE** DEFAULT/NO ACTION/RESTRICT/CASCADE/SET NULL
			- Add/Drop **ON UPDATE** DEFAULT/NO ACTION/RESTRICT/CASCADE/SET NULL
			
### Create go struct from postgresql table name
CreateStruct(conn *pg.DB, tableName string, filePath string) (err error)
```
if conn, err := psql.Conn(true); err == nil {
	shifter.NewShifter().CreateStruct(conn, "address", "")
}
```
#### OUTPUT
![Screenshot 2019-12-08 at 10 09 43 PM](https://user-images.githubusercontent.com/20511920/70392617-db073f80-1a07-11ea-856c-cf83247db3dd.png)

### Create table from struct
CreateTable(conn *pg.DB, structModel interface{}) (err error)
```
type TestAddress struct {
	tableName struct{} `sql:"test_address"`
	AddressID int      `sql:"address_id,type:serial NOT NULL PRIMARY KEY"`
	Address   string   `sql:"city,type:text"`
	City      string   `sql:"city,type:varchar(25) NULL"`
}

s := NewShifter()
err := s.CreateTable(conn, &TestAddress{})
```