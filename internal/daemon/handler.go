package daemon

import (
	"time"

	"github.com/dongt/queryli/internal/db"
	"github.com/dongt/queryli/internal/ipc"
)

// RequestHandler implements ipc.Handler to process client requests.
type RequestHandler struct {
	connector db.Connector
	meta      *Meta
	startTime time.Time
}

// NewRequestHandler creates a handler for the daemon.
func NewRequestHandler(connector db.Connector, meta *Meta) *RequestHandler {
	return &RequestHandler{
		connector: connector,
		meta:      meta,
		startTime: meta.StartTime,
	}
}

// HandleRequest processes an incoming request and returns a response.
func (h *RequestHandler) HandleRequest(req *ipc.Request) *ipc.Response {
	switch req.Type {
	case "ping":
		return h.handlePing()
	case "query":
		return h.handleQuery(req.SQL)
	case "exec":
		return h.handleExec(req.SQL)
	case "status":
		return h.handleStatus()
	case "shutdown":
		return h.handleShutdown()
	default:
		return &ipc.Response{
			OK:    false,
			Error: "unknown request type: " + req.Type,
		}
	}
}

func (h *RequestHandler) handlePing() *ipc.Response {
	if err := h.connector.Ping(); err != nil {
		return &ipc.Response{
			OK:    false,
			Error: "database ping failed: " + err.Error(),
		}
	}
	return &ipc.Response{
		OK:      true,
		Message: "pong",
		Status:  h.buildStatus(),
	}
}

func (h *RequestHandler) handleQuery(sqlStr string) *ipc.Response {
	result, err := h.connector.Query(sqlStr)
	if err != nil {
		return &ipc.Response{
			OK:    false,
			Error: err.Error(),
		}
	}
	return &ipc.Response{
		OK:      true,
		Columns: result.Columns,
		Rows:    result.Rows,
	}
}

func (h *RequestHandler) handleExec(sqlStr string) *ipc.Response {
	rowsAffected, err := h.connector.Exec(sqlStr)
	if err != nil {
		return &ipc.Response{
			OK:    false,
			Error: err.Error(),
		}
	}
	msg := "executed successfully"
	if rowsAffected > 0 {
		msg = "%d row(s) affected"
	}
	return &ipc.Response{
		OK:           true,
		RowsAffected: rowsAffected,
		Message:      msg,
	}
}

func (h *RequestHandler) handleStatus() *ipc.Response {
	return &ipc.Response{
		OK:     true,
		Status: h.buildStatus(),
	}
}

func (h *RequestHandler) handleShutdown() *ipc.Response {
	return &ipc.Response{
		OK:      true,
		Message: "bye",
	}
}

func (h *RequestHandler) buildStatus() *ipc.Status {
	uptime := time.Since(h.startTime).Truncate(time.Second).String()
	return &ipc.Status{
		Profile: h.meta.Profile,
		DBType:  h.meta.DBType,
		Host:    h.meta.Host,
		Uptime:  uptime,
	}
}
