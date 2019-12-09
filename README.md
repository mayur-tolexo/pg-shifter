[![Godocs](https://img.shields.io/badge/golang-documentation-blue.svg)](https://www.godoc.org/github.com/mayur-tolexo/pg-shifter)
[![Go Report Card](https://goreportcard.com/badge/github.com/mayur-tolexo/pg-shifter)](https://goreportcard.com/report/github.com/mayur-tolexo/pg-shifter)
[![Open Source Helpers](https://www.codetriage.com/mayur-tolexo/sworker/badges/users.svg)](https://www.codetriage.com/mayur-tolexo/pg-shifter)
[![Release](https://img.shields.io/github/release/mayur-tolexo/sworker.svg?style=flat-square)](https://github.com/mayur-tolexo/pg-shifter/releases)

# pg-shifter
Golang struct to postgres table shifter.

### Features
1. [Create table](#create-table)
2. [Create go struct from postgresql table name](#create-go-struct-from-postgresql-table-name)
3. [Create enum](#recovery)
4. [Create history table with after update/delete triggers](#recovery)
5. [Alter table](#recovery)
	1. [Add New Column](#add-new-column)
	2. [Remove existing column](#remove-existing-column)
	3. [Modify existing column](#modify-existing-column)
		1. [Modify datatype](#modify-datatype)
		2. Modify data length (e.g. varchar(255) to varchar(100))
		3. Add/Drop default value
		4. Add/Drop Not Null Constraint
		5. Add/Drop constraint (Unique/Foreign Key)
		6. [Modify constraint](#modify-constraint)
			1. Set constraint deferrable
				1. Initially deferred
				1. Initially immediate
			2. Set constraint not deferrable
			3. Add/Drop **ON DELETE** DEFAULT/NO ACTION/RESTRICT/CASCADE/SET NULL
			4. Add/Drop **ON UPDATE** DEFAULT/NO ACTION/RESTRICT/CASCADE/SET NULL


### 1. Create table
CreateTable(conn *pg.DB, model interface{}) (err error)  
i) Directly passing struct model  
ii) Passing table name after setting model  

##### i) Directly passing struct model
```
type TestAddress struct {
	tableName struct{} `sql:"test_address"`
	AddressID int      `sql:"address_id,type:serial NOT NULL PRIMARY KEY"`
	Address   string   `sql:"city,type:text"`
	City      string   `sql:"city,type:varchar(25) NULL"`
}

s := shifter.NewShifter()
err := s.CreateTable(conn, &TestAddress{})
```
##### ii) Passing table name after setting model
```
type TestAddress struct {
	tableName struct{} `sql:"test_address"`
	AddressID int      `sql:"address_id,type:serial NOT NULL PRIMARY KEY"`
	Address   string   `sql:"city,type:text"`
	City      string   `sql:"city,type:varchar(25) NULL"`
}

s := shifter.NewShifter()
s.SetTableModel(&TestAddress{})
err := s.CreateTable(conn, "test_address")
```


### 2. Create go struct from postgresql table name
CreateStruct(conn *pg.DB, tableName string, filePath string) (err error)
```
if conn, err := psql.Conn(true); err == nil {
	shifter.NewShifter().CreateStruct(conn, "address", "")
}
```
#### OUTPUT
![Screenshot 2019-12-08 at 10 09 43 PM](https://user-images.githubusercontent.com/20511920/70392617-db073f80-1a07-11ea-856c-cf83247db3dd.png)
