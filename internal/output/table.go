package output

import (
	"bytes"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// TableFormatter formats results as an ASCII table.
type TableFormatter struct{}

func (f *TableFormatter) Format(result *QueryResult) string {
	if len(result.Columns) == 0 {
		return "No results.\n"
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.Options(
		tablewriter.WithHeader(result.Columns),
		tablewriter.WithBorders(tw.BorderNone),
		tablewriter.WithHeaderAutoFormat(tw.Off),
		tablewriter.WithRowAlignment(tw.AlignLeft),
	)

	for _, row := range result.Rows {
		table.Append(row)
	}

	table.Render()

	// Add row count
	if len(result.Rows) > 0 {
		buf.WriteString(fmt.Sprintf("(%d row(s))\n", len(result.Rows)))
	}

	return buf.String()
}
