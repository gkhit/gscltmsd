package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	// MS SQL
	_ "github.com/denisenkom/go-mssqldb"
)

type (
	// Options
	Options struct {
		Host           string `json:"host"`
		Port           uint16 `json:"port,omitempty"`
		DBName         string `json:"dbname"`
		User           string `json:"user"`
		Password       string `json:"password,omitempty"`
		Timeout        int64  `json:"timeout,omitempty"`
		EntryPointFunc string `json:"entry_point"`
		ToXML          bool   `json:"to_xml,omitempty"`
		XMLRoot        string `json:"xml_root,omitempty"`
		XMLExtArray    bool   `json:"xml_ext_array,omitempty"`
	}
)

// New return new database connection pool
func New(o *Options) *sql.DB {
	var (
		err     error
		connStr string
		poolDB  *sql.DB
	)

	// Build connection string
	if o.Port <= 0 {
		connStr = fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;",
			o.Host, o.User, o.Password, o.DBName)
	} else {
		connStr = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
			o.Host, o.User, o.Password, o.Port, o.DBName)
	}
	// Create connection pool
	log.Printf("[INFO] Try connect to SQL Server: %s\n", connStr)
	poolDB, err = sql.Open("sqlserver", connStr)
	if err != nil {
		log.Fatalf("[ERROR] Can't connect to SQL server. %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Timeout)*time.Second)
	defer cancel()
	err = poolDB.PingContext(ctx)
	if err != nil {
		log.Fatalf("[ERROR] Can't PING to SQL server. %v\n", err)
	}
	return poolDB
}
