package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/dongt/queryli/internal/config"
)

type MySQLConnector struct {
	profile config.Profile
	db      *sql.DB
}

func newMySQLConnector(profile config.Profile) *MySQLConnector {
	return &MySQLConnector{profile: profile}
}

func (c *MySQLConnector) Connect() error {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.profile.User, c.profile.Password, c.profile.Host, c.profile.Port, c.profile.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping mysql: %w", err)
	}

	c.db = db
	return nil
}

func (c *MySQLConnector) Query(sqlStr string) (*QueryResult, error) {
	return baseQuery(c.db, sqlStr)
}

func (c *MySQLConnector) Exec(sqlStr string) (int64, error) {
	return baseExec(c.db, sqlStr)
}

func (c *MySQLConnector) Ping() error {
	return c.db.Ping()
}

func (c *MySQLConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
