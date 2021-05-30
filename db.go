package main

import (
	"database/sql"

	"github.com/ren-kt/uranai_api/fortune"

	_ "modernc.org/sqlite"
)

type MyDb struct {
	db *sql.DB
}

func NewMyDb() (*MyDb, error) {
	db, err := sql.Open("sqlite", "fortune.db")
	if err != nil {
		return nil, err
	}
	return &MyDb{db: db}, nil
}

func (md *MyDb) CreateTable() error {
	const sqlStr = `CREATE TABLE IF NOT EXISTS fortunes(
		id		INTEGER PRIMARY KEY,
		result  TEXT NOT NULL,
		text	TEXT NOT NULL
	);`

	_, err := md.db.Exec(sqlStr)
	if err != nil {
		return err
	}

	return nil
}

func (md *MyDb) GetText(result string) (string, error) {
	const sqlStr = `SELECT fortunes.text FROM fortunes ORDER BY RANDOM() limit 1`
	row := md.db.QueryRow(sqlStr, result)

	var fortune fortune.Fortune
	err := row.Scan(&fortune.Text)
	if err != nil {
		return "", err
	}

	return fortune.Text, nil
}

func (md *MyDb) GetFortune(id int) (*fortune.Fortune, error) {
	const sqlStr = `SELECT * FROM fortunes where id = ?`
	row := md.db.QueryRow(sqlStr, id)

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

func (md *MyDb) GetFortuneAll() ([]*fortune.Fortune, error) {
	const sqlStr = `SELECT * FROM fortunes ORDER BY id DESC`
	rows, err := md.db.Query(sqlStr)
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

func (md *MyDb) Updatefortune(f *fortune.Fortune) error {
	const sqlStr = `UPDATE fortunes SET result = ?, text = ? WHERE id = ?`
	_, err := md.db.Exec(sqlStr, f.Result, f.Text, f.Id)
	if err != nil {
		return err
	}

	return nil
}

func (md *MyDb) Deletefortune(id int) error {
	const sqlStr = `DELETE FROM fortunes WHERE id = ?`
	_, err := md.db.Exec(sqlStr, id)
	if err != nil {
		return err
	}

	return nil
}

func (md *MyDb) Newfortune(fortune *fortune.Fortune) error {
	const sqlStr = `INSERT INTO fortunes(result, text) VALUES (?,?);`
	_, err := md.db.Exec(sqlStr, fortune.Result, fortune.Text)
	if err != nil {
		return err
	}
	return nil
}
