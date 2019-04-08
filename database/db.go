package database

import (
	"io/ioutil"

	"github.com/jackc/pgx"
)

var connPool *pgx.ConnPool = nil

const maxConn = 2000
const dbSchema = "./database/config/db_schema.sql"

func Connect(psqURI string) error {
	if connPool != nil {
		return nil
	}
	config, err := pgx.ParseURI(psqURI)
	if err != nil {
		return err
	}

	connPool, err = pgx.NewConnPool(
		pgx.ConnPoolConfig{
			ConnConfig:     config,
			MaxConnections: maxConn,
		})
	if err != nil {
		return err
	}

	err = LoadSchemaSQL()
	if err != nil {
		return err
	}

	return nil
}

func LoadSchemaSQL() error {
	if connPool == nil {
		return pgx.ErrDeadConn
	}

	content, err := ioutil.ReadFile(dbSchema)
	if err != nil {
		return err
	}

	tx, err := connPool.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(string(content)); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func Disconn() {
	if connPool != nil {
		connPool.Close()
		connPool = nil
	}
}

func Query(sql string, args ...interface{}) (*pgx.Rows, error) {
	if connPool == nil {
		return nil, pgx.ErrDeadConn
	}
	return connPool.Query(sql, args...)
}

func QueryRow(sql string, args ...interface{}) *pgx.Row {
	if connPool == nil {
		return nil
	}
	return connPool.QueryRow(sql, args...)
}

func Exec(sql string, args ...interface{}) (pgx.CommandTag, error) {
	if connPool == nil {
		return "", pgx.ErrDeadConn
	}

	tx, err := connPool.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	tag, err := tx.Exec(sql, args...)
	if err != nil {
		return "", err
	}
	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return tag, nil
}
