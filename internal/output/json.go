package output

import (
	"encoding/json"
)

// JSONFormatter formats results as JSON.
type JSONFormatter struct{}

func (f *JSONFormatter) Format(result *QueryResult) string {
	type jsonResult struct {
		Columns []string   `json:"columns"`
		Rows    [][]string `json:"rows"`
		Count   int        `json:"count"`
	}

	jr := jsonResult{
		Columns: result.Columns,
		Rows:    result.Rows,
		Count:   len(result.Rows),
	}

	// Handle nil slices
	if jr.Columns == nil {
		jr.Columns = []string{}
	}
	if jr.Rows == nil {
		jr.Rows = [][]string{}
	}

	data, _ := json.MarshalIndent(jr, "", "  ")
	return string(data) + "\n"
}
