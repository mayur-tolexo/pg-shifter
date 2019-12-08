package model

//ColSchema : Table Column Schema Model
type ColSchema struct {
	TableName         string `sql:"-"`
	StructColumnName  string `sql:"-"`
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
	SeqName           string `sql:"seq_name"`
	SeqDataType       string `sql:"seq_data_type"`
	Position          int    `sql:"position"`
	IsFkUnique        bool   `sql:"-"`
	FkUniqueName      string `sql:"-"`
}

//UKSchema : Unique Schema Model
type UKSchema struct {
	ConstraintName string `sql:"conname"`
	Columns        string `sql:"col"`
}

//Index model
type Index struct {
	IdxName string `sql:"index_name"`
	IType   string `sql:"itype"`
	Columns string `sql:"col"`
}
