package openapi

import (
	"database/sql"
	_ "github.com/lib/pq"
	"time"
)

type DB struct {
	PostgresUri string `yaml:"url"`
	MaxOpen     int    `yaml:"maxOpen"`
	MaxIdle     int    `yaml:"maxIdle"`
	MaxLifetime int    `yaml:"maxLifetime"`
	db          *sql.DB
}

func (d *DB) Start() error {
	if d.db != nil {
		return nil
	}

	db, err := sql.Open("postgres", d.PostgresUri)
	if err != nil {
		return err
	}
	d.db = db

	if d.MaxOpen < 0 {
		d.MaxOpen = 1
	}

	if d.MaxIdle < 0 {
		d.MaxIdle = 1
	} else if d.MaxIdle > d.MaxOpen {
		d.MaxIdle = d.MaxOpen
	}

	d.db.SetMaxOpenConns(d.MaxOpen)
	d.db.SetMaxIdleConns(d.MaxIdle)

	if d.MaxLifetime > 0 {
		d.db.SetConnMaxLifetime(time.Second * time.Duration(d.MaxLifetime))
	}

	return nil
}

func (d *DB) Stop() {
	if d.db != nil {
		_ = d.db.Close()
		d.db = nil
	}
}

func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	r, e := d.db.Exec(query, args...)
	return r, e
}

func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	r, e := d.db.Query(query, args...)
	return r, e
}

func (d *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

func (d *DB) Prepare(sql string) (*sql.Stmt, error) {
	return d.db.Prepare(sql)
}

func (d *DB) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}
