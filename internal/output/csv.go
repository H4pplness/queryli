package output

import (
	"bytes"
	"encoding/csv"
)

// CSVFormatter formats results as CSV.
type CSVFormatter struct{}

func (f *CSVFormatter) Format(result *QueryResult) string {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	if len(result.Columns) > 0 {
		writer.Write(result.Columns)
	}

	// Write rows
	for _, row := range result.Rows {
		writer.Write(row)
	}

	writer.Flush()
	return buf.String()
}
