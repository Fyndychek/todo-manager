package storage

import (
	"database/sql"
	"log"
	"time"
)

type DBtodo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	Created   time.Time `json:"created_at"`
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
	insert = `
INSERT INTO todos (
	title, completed, created
) VALUES (
	?, ?, ?
)
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

func AddTodo(t DBtodo, db *sql.DB) (int64, error) {
	var result sql.Result
	var err error
	if result, err = db.Exec(insert, t.Title, t.Completed, t.Created); err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, err
}
