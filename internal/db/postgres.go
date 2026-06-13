package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/dongt/queryli/internal/config"
)

type PostgresConnector struct {
	profile config.Profile
	db      *sql.DB
}

func newPostgresConnector(profile config.Profile) *PostgresConnector {
	return &PostgresConnector{profile: profile}
}

func (c *PostgresConnector) Connect() error {
	sslmode := c.profile.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.profile.Host, c.profile.Port, c.profile.User, c.profile.Password, c.profile.DBName, sslmode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping postgres: %w", err)
	}

	c.db = db
	return nil
}

func (c *PostgresConnector) Query(sqlStr string) (*QueryResult, error) {
	return baseQuery(c.db, sqlStr)
}

func (c *PostgresConnector) Exec(sqlStr string) (int64, error) {
	return baseExec(c.db, sqlStr)
}

func (c *PostgresConnector) Ping() error {
	return c.db.Ping()
}

func (c *PostgresConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
