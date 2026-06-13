package ipc

// Request is a message sent from client to daemon.
type Request struct {
	Type string `json:"type"` // query | exec | ping | status | shutdown
	SQL  string `json:"sql"`
}

// Response is a message sent from daemon to client.
type Response struct {
	OK           bool       `json:"ok"`
	Columns      []string   `json:"columns,omitempty"`
	Rows         [][]string `json:"rows,omitempty"`
	RowsAffected int64      `json:"rows_affected,omitempty"`
	Message      string     `json:"message,omitempty"`
	Error        string     `json:"error,omitempty"`
	Status       *Status    `json:"status,omitempty"`
}

// Status contains information about the daemon and active connection.
type Status struct {
	Profile string `json:"profile"`
	DBType  string `json:"db_type"`
	Host    string `json:"host"`
	Uptime  string `json:"uptime"`
}
