package main

import (
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/ren-kt/uranai_api/fortune"
)

type DB interface {
	CreateTable() error
	GetText(result string) (string, error)
	GetFortune(id int) (*fortune.Fortune, error)
	GetFortuneAll() ([]*fortune.Fortune, error)
	Updatefortune(f *fortune.Fortune) error
	Deletefortune(id int) error
	Newfortune(fortune *fortune.Fortune) error
	MultipleNewfortune(entityCh <-chan []string, multipluNum int) <-chan error
}

type Sqlite struct {
	db *sql.DB
}

func NewSqlite() (DB, error) {
	dsn := "host=postgres user=user dbname=app_db password=password sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(10 * time.Second)
	db.SetConnMaxLifetime(10 * time.Second)

	return &Sqlite{db: db}, nil
}

func (sqlite *Sqlite) CreateTable() error {
	const sqlStr = `CREATE TABLE IF NOT EXISTS fortunes(
		id		SERIAL PRIMARY KEY,
		result  TEXT NOT NULL,
		text	TEXT NOT NULL
	);`

	_, err := sqlite.db.Exec(sqlStr)
	if err != nil {
		return err
	}

	return nil
}

func (sqlite *Sqlite) GetText(result string) (string, error) {
	const sqlStr = `SELECT fortunes.text FROM fortunes where result = $1 ORDER BY RANDOM() limit 1`
	row := sqlite.db.QueryRow(sqlStr, result)

	var fortune fortune.Fortune
	err := row.Scan(&fortune.Text)
	if err != nil {
		return "", err
	}

	return fortune.Text, nil
}

func (sqlite *Sqlite) GetFortune(id int) (*fortune.Fortune, error) {
	const sqlStr = `SELECT * FROM fortunes where id = $1`
	row := sqlite.db.QueryRow(sqlStr, id)

	var fortune fortune.Fortune
	err := row.Scan(&fortune.Id, &fortune.Result, &fortune.Text)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &fortune, nil
}

func (sqlite *Sqlite) GetFortuneAll() ([]*fortune.Fortune, error) {
	const sqlStr = `SELECT * FROM fortunes ORDER BY id DESC`
	rows, err := sqlite.db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fortunes []*fortune.Fortune
	for rows.Next() {
		var fortune fortune.Fortune
		err := rows.Scan(&fortune.Id, &fortune.Result, &fortune.Text)
		if err != nil {
			return nil, err
		}
		fortunes = append(fortunes, &fortune)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return fortunes, nil
}

func (sqlite *Sqlite) Updatefortune(f *fortune.Fortune) error {
	const sqlStr = `UPDATE fortunes SET result = $1, text = $2 WHERE id = $3`
	_, err := sqlite.db.Exec(sqlStr, f.Result, f.Text, f.Id)
	if err != nil {
		return err
	}

	return nil
}

func (sqlite *Sqlite) Deletefortune(id int) error {
	const sqlStr = `DELETE FROM fortunes WHERE id = $1`
	_, err := sqlite.db.Exec(sqlStr, id)
	if err != nil {
		return err
	}

	return nil
}

func (sqlite *Sqlite) Newfortune(fortune *fortune.Fortune) error {
	const sqlStr = `INSERT INTO fortunes(result, text) VALUES ($1,$2);`
	_, err := sqlite.db.Exec(sqlStr, fortune.Result, fortune.Text)
	if err != nil {
		return err
	}
	return nil
}

func (sqlite *Sqlite) MultipleNewfortune(lineCh <-chan []string, multipluNum int) <-chan error {
	errCh := make(chan error)

	stmt, err := sqlite.db.Prepare("INSERT INTO fortunes(result, text) VALUES ($1,$2)")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(multipluNum)
	for i := 0; i < multipluNum; i++ {
		go func() {
			defer wg.Done()
			for fortune := range lineCh {
				_, err := stmt.Exec(fortune[0], fortune[1])
				if err != nil {
					errCh <- err
				}
			}
		}()
	}

	go func() {
		defer stmt.Close()
		wg.Wait()
		close(errCh)
	}()

	return errCh
}
