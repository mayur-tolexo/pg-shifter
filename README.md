[![Godocs](https://img.shields.io/badge/golang-documentation-blue.svg)](https://www.godoc.org/github.com/mayur-tolexo/pg-shifter)
[![Go Report Card](https://goreportcard.com/badge/github.com/mayur-tolexo/pg-shifter)](https://goreportcard.com/report/github.com/mayur-tolexo/pg-shifter)
[![Open Source Helpers](https://www.codetriage.com/mayur-tolexo/sworker/badges/users.svg)](https://www.codetriage.com/mayur-tolexo/pg-shifter)
[![Release](https://img.shields.io/github/release/mayur-tolexo/sworker.svg?style=flat-square)](https://github.com/mayur-tolexo/pg-shifter/releases)

# pg-shifter
Golang struct to postgres table shifter. go1.9+ required.  

The main objective is to keep the table complete schema details in golang table struct only.
If any column is missing is struct which is their in the database table then that column will be dropped.  
If any column is extra in struct which is not their in the database table then that column will be added.  
If any modification in column type/constraint in struct which is not matching with database then that will be modified.  
So, to make sure that your go table struct tells you the actual schema of table and to make alter easy we have created this migrator.  

You are most welcome to contribute :)  

## Install
1. go get github.com/mayur-tolexo/pg-shifter
1. go get github.com/tools/godep
1. cd $GOPATH/src/github.com/mayur-tolexo/pg-shifter/
1. godep restore -v

## Features
1. [Create Table](#create-table)
2. [Create Enum](#create-enum)
3. [Upsert Enum](#upsert-enum)
4. [Create Index](#create-index)
5. [Create Unique Key](#create-unique-key)
6. [Upsert Unique Key](#upsert-unique-key)
7. [Create All Tables](#create-all-tables)
6. [Alter Table](#alter-table)
6. [Alter All Tables](#alter-all-tables)
6. [Drop Table](#drop-table)
6. [Drop All Tables](#drop-all-tables)
8. [Create Table Struct](#create-table-struct)
8. Create history table
8. Add trigger

## Alter table supported operations:
1. [Add New Column](#add-new-column)
2. [Remove existing column](#remove-existing-column)
3. Modify existing column
	1. Modify datatype
	2. Modify data length (e.g. varchar(255) to varchar(100))
	3. Add/Drop default value
	4. Add/Drop Not Null Constraint
	5. Add/Drop constraint (Unique/Foreign Key)
	6. Modify constraint
		1. Set constraint deferrable
			1. Initially deferred
			1. Initially immediate
		2. Set constraint not deferrable
		3. Add/Drop FOREIGN KEY **ON DELETE** DEFAULT/NO ACTION/RESTRICT/CASCADE/SET NULL
		4. Add/Drop FOREIGN KEY **ON UPDATE** DEFAULT/NO ACTION/RESTRICT/CASCADE/SET NULL

## Create Table
__CreateTable(conn *pg.DB, model interface{}) (err error)__  

This will create table if not exists from go struct.
Also, if any enum associated to the table struct then that will be created as well.  
All the unique keys and index associated to the table struct will be created as well.
```
i) Directly passing struct model  
ii) Passing table name after setting model  
```

##### i) Directly passing struct model
```
type TestAddress struct {
	tableName struct{} `sql:"test_address"`
	AddressID int      `sql:"address_id,type:serial NOT NULL PRIMARY KEY"`
	Address   string   `sql:"address,type:text"`
	City      string   `sql:"city,type:varchar(25) NULL"`
}

s := shifter.NewShifter()
err := s.CreateTable(conn, &TestAddress{})
```
##### ii) Passing table name after setting model
```
s := shifter.NewShifter()
s.SetTableModel(&TestAddress{})
err := s.CreateTable(conn, "test_address")
```

## Create Enum
__CreateAllEnum(conn *pg.DB, model interface{}) (err error)__   

This will create all the enum associated to the given table.  
To define enum on table struct you need to create a method with following signature:  
```
func (tableStruct) Enum() map[string][]string
```
Here returned map's key is enum name and value is slice of enum values.  
If enum already exist in database then it will not create enum again.
```
i) Directly passing struct model   
ii) Passing table name after setting model  
```

##### i) Directly passing struct model
```
type TestAddress struct {
	tableName struct{} `sql:"test_address"`
	AddressID int      `sql:"address_id,type:serial NOT NULL PRIMARY KEY"`
	Address   string   `sql:"address,type:text"`
	City      string   `sql:"city,type:varchar(25) NULL"`
	Status    string   `sql:"status,type:address_status"`
}

//Enum of the table.
func (TestAddress) Enum() map[string][]string {
	enm := map[string][]string{
		"address_status": {"enable", "disable"},
	}
	return enm
}

s := shifter.NewShifter()
err := s.CreateAllEnum(conn, &TestAddress{})
```
##### ii) Passing table name after setting model
```
s := shifter.NewShifter()
s.SetTableModel(&TestAddress{})
err = s.CreateAllEnum(conn, "test_address")
```

---------------

__CreateEnum(conn *pg.DB, model interface{}, enumName string) (err error)__  

This will create given enum if associated to given table  
```
i) Directly passing struct model   
ii) Passing table name after setting model  
```

##### i) Directly passing struct model
```
s := shifter.NewShifter()
err := s.CreateEnum(conn, &TestAddress{}, "address_status")
```
##### ii) Passing table name after setting model
```
s := shifter.NewShifter()
s.SetTableModel(&TestAddress{})
err = s.CreateEnum(conn, "test_address", "address_status")
```


## Upsert Enum
__UpsertAllEnum(conn *pg.DB, model interface{}) (err error)__   

This will create/update all the enum associated to the given table.
To define enum on table struct you need to create a method with following signature:  
```
func (tableStruct) Enum() map[string][]string
```
Here returned map's key is enum name and value is slice of enum values.  
If enum already exist in database then it will update the enum value which are missing in the database.

```
i) Directly passing struct model   
ii) Passing table name after setting model  
```

##### i) Directly passing struct model
```
type TestAddress struct {
	tableName struct{} `sql:"test_address"`
	AddressID int      `sql:"address_id,type:serial NOT NULL PRIMARY KEY"`
	Address   string   `sql:"address,type:text"`
	City      string   `sql:"city,type:varchar(25) NULL"`
	Status    string   `sql:"status,type:address_status"`
}

//Enum of the table.
func (TestAddress) Enum() map[string][]string {
	enm := map[string][]string{
		"address_status": {"enable", "disable"},
	}
	return enm
}

s := shifter.NewShifter()
err := s.UpsertAllEnum(conn, &TestAddress{})
```
##### ii) Passing table name after setting model
```
s := shifter.NewShifter()
s.SetTableModel(&TestAddress{})
err = s.UpsertAllEnum(conn, "test_address")
```

---------------

__UpsertEnum(conn *pg.DB, model interface{}, enumName string) (err error)__  

This will create/update given enum if associated to given table  
```
i) Directly passing struct model   
ii) Passing table name after setting model  
```

##### i) Directly passing struct model
```
s := shifter.NewShifter()
err := s.UpsertEnum(conn, &TestAddress{}, "address_status")
```
##### ii) Passing table name after setting model
```
s := shifter.NewShifter()
s.SetTableModel(&TestAddress{})
err = s.UpsertEnum(conn, "test_address", "address_status")
```


## Create Index
__CreateAllIndex(conn *pg.DB, model interface{}, skipPrompt ...bool) (err error)__   

This will create all the index associated to the given table.  
If __skipPrompt__ is enabled then it won't ask for confirmation before creating index. Default is disable.  
To define index on table struct you need to create a method with following signature:  
```
func (tableStruct) Index() map[string]string  
```
Here returned map's key is column which need to index and value is the type of data structure to user for indexing. Default is btree. For composite index you can add column comma seperated.
```
i) Directly passing struct model   
ii) Passing table name after setting model  
```

##### i) Directly passing struct model
```
type TestAddress struct {
	tableName struct{}    `sql:"test_address"`
	AddressID int         `json:"address_id,omitempty" sql:"address_id,type:serial PRIMARY KEY"`
	City      string      `json:"city" sql:"city,type:varchar(25) UNIQUE"`
	Status    string      `json:"status,omitempty" sql:"status,type:address_status"`
	Info      interface{} `sql:"info,type:jsonb"`
}

//Index of the table. For composite index use ,
//Default index type is btree. For gin index use gin
func (TestAddress) Index() map[string]string {
	idx := map[string]string{
		"status":            shifter.BtreeIndex,
		"info":              shifter.GinIndex,
		"address_id,status": shifter.BtreeIndex,
	}
	return idx
}

s := shifter.NewShifter()
err := s.CreateAllIndex(conn, &TestAddress{})
```
##### ii) Passing table name after setting model
```
s := shifter.NewShifter()
s.SetTableModel(&TestAddress{})
err = s.CreateAllIndex(conn, "test_address")
```


## Create Unique Key
__CreateAllUniqueKey(conn *pg.DB, model interface{}, skipPrompt ...bool) (err error)__   

This will create all the composite unique key associated to the given table.  
If __skipPrompt__ is enabled then it won't ask for confirmation before creating unique key. Default is disable.  
To define composite unique key on table struct you need to create a method with following signature:  
```
func (tableStruct) UniqueKey() []string
```
Here returned slice is the columns comma seperated.  
If single column need to create unique key then use UNIQUE sql tag for column.  
```
i) Directly passing struct model   
ii) Passing table name after setting model  
```

##### i) Directly passing struct model
```
type TestAddress struct {
	tableName struct{}  `sql:"test_address"`
	AddressID int       `json:"address_id,omitempty" sql:"address_id,type:serial PRIMARY KEY"`
	City      string    `json:"city" sql:"city,type:varchar(25) UNIQUE"`
	Status    string    `json:"status,omitempty"
}

//UniqueKey of the table. This is for composite unique keys
func (TestAddress) UniqueKey() []string {
	uk := []string{
		"address_id,status,city",
	}
	return uk
}

s := shifter.NewShifter()
err := s.CreateAllUniqueKey(conn, &TestAddress{})
```
##### ii) Passing table name after setting model
```
s := shifter.NewShifter()
s.SetTableModel(&TestAddress{})
err = s.CreateAllUniqueKey(conn, "test_address")
```

## Upsert Unique Key
__UpsertAllUniqueKey(conn *pg.DB, model interface{}, skipPrompt ...bool) (err error)__   

This will create all the composite unique key associated to the given table.  
Modify composite unique key which are not matching with table and struct.  
Drop composite unique key which exists in table but not in struct.  
If __skipPrompt__ is enabled then it won't ask for confirmation before upserting unique key. Default is disable.  
To define composite unique key on table struct you need to create a method with following signature:  
```
func (tableStruct) UniqueKey() []string
```
Here returned slice is the columns comma seperated.  
If single column need to create unique key then use UNIQUE sql tag for column.  
```
i) Directly passing struct model   
ii) Passing table name after setting model  
```

##### i) Directly passing struct model
```
type TestAddress struct {
	tableName struct{}  `sql:"test_address"`
	AddressID int       `json:"address_id,omitempty" sql:"address_id,type:serial PRIMARY KEY"`
	City      string    `json:"city" sql:"city,type:varchar(25) UNIQUE"`
	Status    string    `json:"status,omitempty"
}

//UniqueKey of the table. This is for composite unique keys
func (TestAddress) UniqueKey() []string {
	uk := []string{
		"address_id,status,city",
	}
	return uk
}

s := shifter.NewShifter()
err := s.UpsertAllUniqueKey(conn, &TestAddress{})
```
##### ii) Passing table name after setting model
```
s := shifter.NewShifter()
s.SetTableModel(&TestAddress{})
err = s.UpsertAllUniqueKey(conn, "test_address")
```


## Create All Tables
__CreateAllTable(conn *pg.DB) (err error)__  

This will create table if not exists from go struct which are set in shifter.
Also, if any enum associated to the table struct then that will be created as well.  
All the unique keys and index associated to the table struct will be created as well.
```
db := []interface{}{&TestAddress{}, &TestUser{}, &TestAdminUser{}}

s := shifter.NewShifter()
s.SetTableModels(db)
err := s.CreateAllTable(conn)
```


## Alter Table
__AlterTable(conn *pg.DB, model interface{}, skipPrompt ...bool) (err error)__   

This will alter table.  
If __skipPrompt__ is enabled then it won't ask for confirmation before upserting unique key. Default is disable.  
```
i) Directly passing struct model   
ii) Passing table name after setting model  
```

##### i) Directly passing struct model
```
type TestAddress struct {
	tableName struct{}  `sql:"test_address"`
	AddressID int       `json:"address_id,omitempty" sql:"address_id,type:serial PRIMARY KEY"`
	City      string    `json:"city" sql:"city,type:varchar(25) UNIQUE"`
	Status    string    `json:"status,omitempty"
}

s := shifter.NewShifter()
err := s.AlterTable(conn, &TestAddress{})
```
##### ii) Passing table name after setting model
```
s := shifter.NewShifter()
s.SetTableModel(&TestAddress{})
err = s.AlterTable(conn, "test_address")
```

## Alter All Tables
__AlterAllTable(conn *pg.DB, skipPrompt ...bool) (err error)__  
This will alter all tables added in shifter using SetTableModels().    
If __skipPrompt__ is enabled then it won't ask for confirmation before upserting unique key. Default is disable.  


```
db := []interface{}{&TestAddress{}, &TestUser{}, &TestAdminUser{}}

s := shifter.NewShifter()
s.SetTableModels(db)
err := s.AlterAllTable(conn)
```

## Drop Table
__DropTable(conn *pg.DB, model interface{}, cascade bool) (err error)__  

This will drop the table from database if exists. Also, if history table associated to this table exists then that will be dropped as well.
If cascade is true then it will drop table with cascade.

##### i) Directly passing struct model
```
s := shifter.NewShifter()
err := s.DropTable(conn, &TestAddress{}, true)
```

##### ii) Passing table name
```
s := shifter.NewShifter()
err := s.DropTable(conn, "test_address", true)
```

## Drop All Tables
__DropAllTable(conn *pg.DB, cascade bool) (err error)__  

This will drop all the table from database if exists which are set in shifter. So, before calling it you need to SetTableModels() on shifter.
Also, if history table associated to this table exists then that will be dropped as well.
If cascade is true then it will drop tables with cascade.

```
db := []interface{}{&TestAddress{}, &TestUser{}, &TestAdminUser{}}

s := shifter.NewShifter()
s.SetTableModels(db)
err := s.DropAllTable(conn, true)
```


## Create Table Struct
CreateStruct(conn *pg.DB, tableName string, filePath string) (err error)
```
if conn, err := psql.Conn(true); err == nil {
	shifter.NewShifter().CreateStruct(conn, "address", "")
}
```
#### OUTPUT
![Screenshot 2019-12-08 at 10 09 43 PM](https://user-images.githubusercontent.com/20511920/70392617-db073f80-1a07-11ea-856c-cf83247db3dd.png)



## Add New Column
Just add new field in the table struct and run AlterTable().  

## Remove Existing Column
Remove field from the table struct which you want to remove and run AlterTable().  

