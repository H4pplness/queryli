package db

import (
	"fmt"

	"github.com/dongt/queryli/internal/config"
)

// QueryResult holds the result of a SQL query.
type QueryResult struct {
	Columns []string
	Rows    [][]string
}

// Connector is the interface for database operations.
type Connector interface {
	Connect() error
	Query(sql string) (*QueryResult, error)
	Exec(sql string) (int64, error)
	Ping() error
	Close() error
}

// NewConnector creates a connector for the given profile type.
func NewConnector(profile config.Profile) (Connector, error) {
	switch profile.Type {
	case "postgres":
		return newPostgresConnector(profile), nil
	case "mysql":
		return newMySQLConnector(profile), nil
	case "sqlite":
		return newSQLiteConnector(profile), nil
	case "oracle":
		return newOracleConnector(profile), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", profile.Type)
	}
}
