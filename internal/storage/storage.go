package storage

import (
	"database/sql"
	"log"
	"time"
)

type todo struct {
	Title     string
	Completed bool
	Created   time.Time
}

const (
	schema = `
	CREATE TABLE IF NOT EXISTS todos(
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  title TEXT,
  completed BOOLEAN,
  created DATETIME
	);
`
)

func NewDatabase(dbFile string) (*sql.DB, error) {

	newDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	ping := newDB.Ping()
	log.Println(ping)

	if _, err := newDB.Exec(schema); err != nil {
		return nil, err
	}

	return newDB, nil
}
