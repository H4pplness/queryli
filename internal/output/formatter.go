package output

// QueryResult is a simplified result for formatting (avoids import cycles).
type QueryResult struct {
	Columns []string
	Rows    [][]string
}

// Formatter formats query results for display.
type Formatter interface {
	Format(result *QueryResult) string
}

// NewFormatter returns the appropriate formatter for the given format.
func NewFormatter(format string) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "csv":
		return &CSVFormatter{}
	default:
		return &TableFormatter{}
	}
}
