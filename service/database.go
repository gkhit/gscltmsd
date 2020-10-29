package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// PoolDB sql.DB wrapper
type PoolDB sql.DB

// NewDatabase return new database connection pool
func NewDatabase(ctx context.Context, c *DatabaseOptions) (*PoolDB, error) {
	var (
		err     error
		connstr string
		pDB     *sql.DB
	)

	// Build connection string
	if c.Port <= 0 {
		connstr = fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;",
			c.Host, c.User, c.Password, c.Database)
	} else {
		connstr = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
			c.Host, c.User, c.Password, c.Port, c.Database)
	}
	// Create connection pool
	log.Printf("Try connect to SQL Server: %s\n", connstr)
	pDB, err = sql.Open("sqlserver", connstr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.Timeout)*time.Second)
	defer cancel()
	err = pDB.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return (*PoolDB)(pDB), nil
}

// CallEntryPoint call entry point of database with payload and topic
func (p *PoolDB) CallEntryPoint(ctx context.Context, entry string, topic string, payload string, timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := (*sql.DB)(p).Conn(ctx)
	if err != nil {
		return
	}
	defer conn.Close()
	_, err = conn.ExecContext(ctx, entry, topic, payload)
	if err != nil {
		log.Println(err)
	}
	return
}
