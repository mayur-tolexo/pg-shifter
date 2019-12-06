[![Godocs](https://img.shields.io/badge/golang-documentation-blue.svg)](https://www.godoc.org/github.com/mayur-tolexo/pg-shifter)
[![Go Report Card](https://goreportcard.com/badge/github.com/mayur-tolexo/pg-shifter)](https://goreportcard.com/report/github.com/mayur-tolexo/pg-shifter)
[![Open Source Helpers](https://www.codetriage.com/mayur-tolexo/sworker/badges/users.svg)](https://www.codetriage.com/mayur-tolexo/pg-shifter)
[![Release](https://img.shields.io/github/release/mayur-tolexo/sworker.svg?style=flat-square)](https://github.com/mayur-tolexo/pg-shifter/releases)

# pg-shifter
Golang struct to postgres table shifter.

### Features
- [Create table from struct](#recovery)
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
			
