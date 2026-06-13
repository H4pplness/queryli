package db

import (
	"database/sql"
	"fmt"
)

// baseQuery runs a SQL query with dynamic column scanning.
func baseQuery(d *sql.DB, rawSQL string) (*QueryResult, error) {
	rows, err := d.Query(rawSQL)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	var resultRows [][]string

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		row := make([]string, len(cols))
		for i, val := range values {
			row[i] = fmt.Sprintf("%v", val)
		}
		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return &QueryResult{
		Columns: cols,
		Rows:    resultRows,
	}, nil
}

// baseExec runs a SQL statement that doesn't return rows.
func baseExec(d *sql.DB, rawSQL string) (int64, error) {
	result, err := d.Exec(rawSQL)
	if err != nil {
		return 0, fmt.Errorf("exec failed: %w", err)
	}
	return result.RowsAffected()
}
