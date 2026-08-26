package storage

import (
	"database/sql"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// структура задачи
type DBtodo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	Created   time.Time `json:"created_at"`
}

// структура патча задачи
type UpdateTodoRequest struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

// заготовки запросов к бд
const (
	schema = `
		CREATE TABLE IF NOT EXISTS todos(
  		id INTEGER PRIMARY KEY AUTOINCREMENT, 
  		title TEXT,
  		completed BOOLEAN,
  		created DATETIME,	
  		user_id INTEGER NOT NULL DEFAULT 1
		);
	`

	insert = `
		INSERT INTO todos (
		title, completed, created, user_id
		) VALUES (?, ?, ?, ?)
	`

	insertUsers = `
		INSERT INTO users (
		username, password_hash
		) VALUES (?,?)
	`

	UserShema = `
		CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at	DATETIME DEFAULT CURRENT_TIMESTAMP
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
	if ping != nil {
		return nil, ping
	}

	if _, err := newDB.Exec(schema); err != nil {
		return nil, err
	}
	if _, err := newDB.Exec(UserShema); err != nil {
		return nil, err
	}

	return newDB, nil
}

func AddTodo(t DBtodo, db *sql.DB, user_id int64) (int64, error) {

	var (
		result sql.Result
		err    error
	)
	if result, err = db.Exec(insert, t.Title, t.Completed, t.Created, user_id); err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func GetTodos(filter string, sortby string, order string, db *sql.DB, userID int64) ([]DBtodo, error) {

	var (
		todos       []DBtodo
		t           DBtodo
		BaseRequest string = "select * from todos"
		request     string
		WhereClause string = ""
	)
	switch filter {
	case "all":
		request = BaseRequest

	case "completed":
		request = BaseRequest
		WhereClause = " AND completed = 1"

	case "active":
		request = BaseRequest
		WhereClause = " AND completed = 0"

	default:
		request = BaseRequest

	}

	switch sortby {
	case "id":
		if order == "desc" {
			request = request + " order by id DESC"
		} else {
			request = request + " order by id ASC"
		}

	case "title":
		if order == "desc" {
			request = request + " order by title DESC"
		} else {
			request = request + " order by title ASC"
		}

	case "completed":
		if order == "desc" {
			request = request + " order by completed DESC"
		} else {
			request = request + " order by completed ASC"
		}

	case "created":
		if order == "desc" {
			request = request + " order by created DESC"
		} else {
			request = request + " order by created ASC"
		}

	default:
		if order == "desc" {
			request = request + " order by title DESC"
		} else {
			request = request + " order by title ASC"
		}
	}

	rows, err := db.Query(request+" where user_id = ?"+WhereClause, userID)
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
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func UpdateTodo(id int, upd UpdateTodoRequest, db *sql.DB, userID int64) (DBtodo, error) {

	var (
		t   DBtodo
		err error
	)
	if upd.Completed != nil && upd.Title == nil {
		if _, err = db.Exec("update todos set completed = ? where id =? AND user_id = ?", *upd.Completed, id, userID); err != nil {
			return t, err
		}
	}
	if upd.Title != nil && upd.Completed == nil {
		if _, err = db.Exec("update todos set title = ? where id =? AND user_id = ?", *upd.Title, id, userID); err != nil {
			return t, err
		}
	}
	if upd.Title != nil && upd.Completed != nil {
		if _, err = db.Exec("update todos set title = ?, completed = ? where id =? AND user_id = ?", *upd.Title, *upd.Completed, id, userID); err != nil {
			return t, err
		}
	}
	rows, err := db.Query("select * from todos where id = ? AND user_id = ?", id, userID)
	if err != nil {
		return t, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.Created)
		if err != nil {
			log.Println(err)
			return t, err
		}
	}
	if err = rows.Err(); err != nil {
		return DBtodo{}, err
	}
	if t.ID == 0 {
		return DBtodo{}, sql.ErrNoRows
	}

	return t, nil
}

func DeleteTodo(id int, db *sql.DB, userID int64) (DBtodo, error) {

	var (
		err error
		t   DBtodo
	)
	//получаем задачу
	rows, err := db.Query("select * from todos where id = ? AND user_id = ?", id, userID)
	if err != nil {
		return DBtodo{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.Created)
		if err != nil {
			log.Println(err)
			return DBtodo{}, err
		}
	}
	if t.ID == 0 {
		return DBtodo{}, sql.ErrNoRows
	}
	if err = rows.Err(); err != nil {
		return DBtodo{}, err
	}

	if _, err = db.Exec("delete from todos where id = ? AND user_id = ?", id, userID); err != nil {
		return DBtodo{}, err
	}
	return t, err
}

func GetStats(db *sql.DB, userID int64) (int, int, int, error) {

	var (
		total     int
		completed int
		pending   int
	)
	rows, err := db.Query("select count(*) from todos where user_id = ?", userID)
	if err != nil {
		return 0, 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, 0, 0, err
		}
	}
	if err = rows.Err(); err != nil {
		return 0, 0, 0, err
	}
	rows, err = db.Query("select count(*) from todos where completed = 1 AND user_id = ?", userID)
	if err != nil {
		return 0, 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&completed); err != nil {
			return 0, 0, 0, err
		}
	}
	if err = rows.Err(); err != nil {
		return 0, 0, 0, err
	}
	pending = total - completed

	return total, completed, pending, nil
}

func CreateUser(username, password string, db *sql.DB) (int64, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	result, err := db.Exec(insertUsers, username, string(hash))
	if err != nil {
		return 0, err
	} else if err == sql.ErrTxDone {
		return 0, err
	}

	return result.LastInsertId()

}

func GetUserByUsername(username string, db *sql.DB) (int64, string, error) {

	var (
		id   int64
		hash string
	)
	row := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username)
	err := row.Scan(&id, &hash)
	if err != nil {
		return 0, "", err
	}

	return id, hash, nil

}
