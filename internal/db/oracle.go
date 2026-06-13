package db

import (
	"database/sql"
	"fmt"

	_ "github.com/sijms/go-ora/v2"

	"github.com/dongt/queryli/internal/config"
)

type OracleConnector struct {
	profile config.Profile
	db      *sql.DB
}

func newOracleConnector(profile config.Profile) *OracleConnector {
	return &OracleConnector{profile: profile}
}

func (c *OracleConnector) Connect() error {
	// go-ora DSN format: oracle://user:pass@host:port/service
	dsn := fmt.Sprintf(
		"oracle://%s:%s@%s:%d/%s",
		c.profile.User, c.profile.Password, c.profile.Host, c.profile.Port, c.profile.Service,
	)

	db, err := sql.Open("oracle", dsn)
	if err != nil {
		return fmt.Errorf("open oracle: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping oracle: %w", err)
	}

	c.db = db
	return nil
}

func (c *OracleConnector) Query(sqlStr string) (*QueryResult, error) {
	return baseQuery(c.db, sqlStr)
}

func (c *OracleConnector) Exec(sqlStr string) (int64, error) {
	return baseExec(c.db, sqlStr)
}

func (c *OracleConnector) Ping() error {
	return c.db.Ping()
}

func (c *OracleConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
