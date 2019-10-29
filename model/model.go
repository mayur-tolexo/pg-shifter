package model

//DBCSchema : Database Column Schema Model
type DBCSchema struct {
	ColumnName        string `sql:"column_name"`
	ColumnDefault     string `sql:"column_default"`
	DataType          string `sql:"data_type"`
	UdtName           string `sql:"udt_name"`
	IsNullable        string `sql:"is_nullable"`
	CharMaxLen        string `sql:"character_maximum_length"`
	ConstraintType    string `sql:"constraint_type"`
	ConstraintName    string `sql:"constraint_name"`
	IsDeferrable      string `sql:"is_deferrable"`
	InitiallyDeferred string `sql:"initially_deferred"`
	ForeignTableName  string `sql:"foreign_table_name"`
	ForeignColumnName string `sql:"foreign_column_name"`
	UpdateType        string `sql:"confupdtype"`
	DeleteType        string `sql:"confdeltype"`
}

//UKSchema : Unique Schema Model
type UKSchema struct {
	ConstraintName string `sql:"conname"`
	Columns        string `sql:"col"`
}

//SCSchema : Struct Column Schema
type SCSchema struct {
	TableName         string
	ColumnName        string
	ColumnDefault     string
	DataType          string
	IsNullable        string
	CharMaxLen        string
	ConstraintType    string
	IsDeferrable      string
	InitiallyDeferred string
	ForeignTableName  string
	ForeignColumnName string
	UpdateType        string
	DeleteType        string
}
