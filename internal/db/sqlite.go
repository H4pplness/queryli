package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/dongt/queryli/internal/config"
)

type SQLiteConnector struct {
	profile config.Profile
	db      *sql.DB
}

func newSQLiteConnector(profile config.Profile) *SQLiteConnector {
	return &SQLiteConnector{profile: profile}
}

func (c *SQLiteConnector) Connect() error {
	db, err := sql.Open("sqlite", c.profile.Path)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping sqlite: %w", err)
	}

	c.db = db
	return nil
}

func (c *SQLiteConnector) Query(sqlStr string) (*QueryResult, error) {
	return baseQuery(c.db, sqlStr)
}

func (c *SQLiteConnector) Exec(sqlStr string) (int64, error) {
	return baseExec(c.db, sqlStr)
}

func (c *SQLiteConnector) Ping() error {
	return c.db.Ping()
}

func (c *SQLiteConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
