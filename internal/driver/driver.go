package driver

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

// DB holds the database connection pool
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifeTime = 5 * time.Minute

// ConnectSQL creates database pool for MySql
func ConnectSQL(dsn string) (*DB, error) {
	database, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}
	database.SetMaxOpenConns(maxOpenDbConn)
	database.SetMaxIdleConns(maxIdleDbConn)
	database.SetConnMaxLifetime(maxDbLifeTime)

	dbConn.SQL = database

	err = testDB(database)
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

// testDB tries to ping the database
func testDB(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	return nil
}

// NewDatabase creates a new database using mySql for the application
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
