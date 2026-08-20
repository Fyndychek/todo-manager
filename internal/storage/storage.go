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
	if ping != nil {
		return nil, ping
	}

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

	return id, nil
}

func GetTodos(filter string, sortby string, order string, db *sql.DB) ([]DBtodo, error) {
	var todos []DBtodo
	var t DBtodo
	var request string
	switch filter {
	case "all":
		request = "select * from todos"

	case "completed":
		request = "select * from todos where completed = 1"

	case "active":
		request = "select * from todos where completed = 0"

	default:
		request = "select * from todos"

	}

	switch sortby {
	case "id":
		if order == "desc" {
			request = request + " order by id DESK"
		} else {
			request = request + " order by id ASK"
		}

	case "title":
		if order == "desc" {
			request = request + " order by title DESK"
		} else {
			request = request + " order by title ASK"
		}

	case "completed":
		if order == "desc" {
			request = request + " order by completed DESK"
		} else {
			request = request + " order by completed ASK"
		}

	case "created":
		if order == "desc" {
			request = request + " order by created DESK"
		} else {
			request = request + " order by created ASK"
		}

	default:
		if order == "desc" {
			request = request + " order by title DESK"
		} else {
			request = request + " order by title ASK"
		}

	}

	rows, err := db.Query(request)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.Created)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		todos = append(todos, t)
	}

	return todos, nil

}
