// +build !darwin
// +build !windows

package cmd

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Database interface
type Database interface {
	Get(shortcode string) (string, error)
	Save(shortcode, url string) (int64, string, error)
	List() ([]Response, error)
}

type sqlite struct {
	Path string
}

func (s sqlite) Save(shortcode, url string) (int64, string, error) {
	db, err := sql.Open("sqlite3", s.Path)
	tx, err := db.Begin()
	if err != nil {
		return 0, shortcode, err
	}
	stmt, err := tx.Prepare("insert into urls(shortcode, url) values(?,?)")
	if err != nil {
		return 0, shortcode, err
	}
	defer stmt.Close()
	result, err := stmt.Exec(shortcode, url)
	if err != nil {
		return 0, shortcode, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, shortcode, err
	}
	tx.Commit()
	//result
	return id, shortcode, nil
}

func (s sqlite) List() ([]Response, error) {
	var response []Response
	db, err := sql.Open("sqlite3", s.Path)
	stmt, err := db.Prepare("select shortcode, url from urls")
	if err != nil {
		return response, err
	}
	defer stmt.Close()
	var url string
	var shortcode string
	rows, err := stmt.Query()
	if err != nil {
		return response, err
	}

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&shortcode, &url)
		if err != nil {
			return response, err
		}
		r := Response{
			ShortCode: shortcode,
			URL:       url,
		}
		response = append(response, r)
	}
	err = rows.Err()
	return response, err
}
func (s sqlite) Get(shortcode string) (string, error) {
	db, err := sql.Open("sqlite3", s.Path)
	stmt, err := db.Prepare("select url from urls where shortcode = ?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	var url string
	err = stmt.QueryRow(shortcode).Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s sqlite) Init() {
	c, err := sql.Open("sqlite3", s.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	sqlStmt := `create table if not exists urls (shortcode text not null primary key, url text);`
	_, err = c.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}
