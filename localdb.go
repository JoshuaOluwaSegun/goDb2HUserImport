package main

import (
	"fmt"
	"strconv"

	//SQL Package
	"github.com/hornbill/sqlx"
	//SQL Drivers
	_ "github.com/alexbrainman/odbc"
	_ "github.com/hornbill/go-mssqldb"
	_ "github.com/hornbill/mysql"
	_ "github.com/hornbill/mysql320"
)

//buildConnectionString -- Build the connection string for the SQL driver
func buildConnectionString() string {
	//	if SQLImportConf.SQLConf.Server == "" || SQLImportConf.SQLConf.Database == "" || SQLImportConf.SQLConf.UserName == "" || SQLImportConf.SQLConf.Password == "" {
	if SQLImportConf.SQLConf.Server == "" || SQLImportConf.SQLConf.Database == "" || SQLImportConf.SQLConf.UserName == "" {
		//Conf not set - log error and return empty string
		logger(4, "Database configuration not set.", true)
		return ""
	}
	logger(1, "Connecting to Database Server: "+SQLImportConf.SQLConf.Server, true)
	connectString := ""
	switch SQLImportConf.SQLConf.Driver {

	case "mssql":
		connectString = "server=" + SQLImportConf.SQLConf.Server
		connectString = connectString + ";database=" + SQLImportConf.SQLConf.Database
		connectString = connectString + ";user id=" + SQLImportConf.SQLConf.UserName
		connectString = connectString + ";password=" + SQLImportConf.SQLConf.Password
		if !SQLImportConf.SQLConf.Encrypt {
			connectString = connectString + ";encrypt=disable"
		}
		if SQLImportConf.SQLConf.Port != 0 {
			dbPortSetting := strconv.Itoa(SQLImportConf.SQLConf.Port)
			connectString = connectString + ";port=" + dbPortSetting
		}

	case "mysql":
		connectString = SQLImportConf.SQLConf.UserName + ":" + SQLImportConf.SQLConf.Password
		connectString = connectString + "@tcp(" + SQLImportConf.SQLConf.Server + ":"
		if SQLImportConf.SQLConf.Port != 0 {
			dbPortSetting := strconv.Itoa(SQLImportConf.SQLConf.Port)
			connectString = connectString + dbPortSetting
		} else {
			connectString = connectString + "3306"
		}
		connectString = connectString + ")/" + SQLImportConf.SQLConf.Database

	case "mysql320":
		var dbPortSetting string
		if SQLImportConf.SQLConf.Port != 0 {
			dbPortSetting = strconv.Itoa(SQLImportConf.SQLConf.Port)
		} else {
			dbPortSetting = "3306"
		}
		connectString = "tcp:" + SQLImportConf.SQLConf.Server + ":" + dbPortSetting
		connectString = connectString + "*" + SQLImportConf.SQLConf.Database + "/" + SQLImportConf.SQLConf.UserName + "/" + SQLImportConf.SQLConf.Password
	case "csv":
		connectString = "DSN=" + SQLImportConf.SQLConf.Database + ";Extended Properties='text;HDR=Yes;FMT=Delimited'"
		SQLImportConf.SQLConf.Driver = "odbc"
	case "excel":
		connectString = "DSN=" + SQLImportConf.SQLConf.Database + ";"
		SQLImportConf.SQLConf.Driver = "odbc"

	}

	return connectString
}

//queryDatabase -- Query Asset Database for assets of current type
//-- Builds map of assets, returns true if successful
func queryDB() bool {

	//Set SWSQLDriver to mysql320
	if SQLImportConf.SQLConf.Driver == "swsql" {
		SQLImportConf.SQLConf.Driver = "mysql320"
	}

	//Clear existing Asset Map down
	ArrUserMaps := make([]map[string]interface{}, 0)
	connString := buildConnectionString()
	if connString == "" {
		return false
	}
	//Connect to the JSON specified DB
	db, err := sqlx.Open(SQLImportConf.SQLConf.Driver, connString)
	if err != nil {
		logger(4, " [DATABASE] Database Connection Error: "+fmt.Sprintf("%v", err), true)
		return false
	}
	defer db.Close()
	//Check connection is open
	err = db.Ping()
	if err != nil {
		logger(4, " [DATABASE] [PING] Database Connection Error: "+fmt.Sprintf("%v", err), true)
		return false
	}
	logger(3, "[DATABASE] Connection Successful", true)
	logger(3, "[DATABASE] Running database query for Customers. Please wait...", true)
	//build query
	sqlQuery := SQLImportConf.SQLConf.Query //BaseSQLQuery
	logger(3, "[DATABASE] Query:"+sqlQuery, false)
	//Run Query
	rows, err := db.Queryx(sqlQuery)
	if err != nil {
		logger(4, " [DATABASE] Database Query Error: "+fmt.Sprintf("%v", err), true)
		return false
	}

	//Build map full of assets
	intUserCount := 0
	for rows.Next() {
		intUserCount++
		results := make(map[string]interface{})
		err = rows.MapScan(results)
		if err != nil {
			//We are going to skip this record as it did not scan properly
			logger(4, " [DATABASE] Database Scan Error: "+fmt.Sprintf("%v", err), true)
			continue
		}
		//Stick marshalled data map in to parent slice
		ArrUserMaps = append(ArrUserMaps, results)
	}
	defer rows.Close()
	logger(3, fmt.Sprintf("[DATABASE] Found %d results", intUserCount), false)
	localDBUsers = ArrUserMaps
	return true
}
