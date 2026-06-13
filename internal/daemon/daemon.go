package daemon

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/dongt/queryli/internal/config"
	"github.com/dongt/queryli/internal/db"
	"github.com/dongt/queryli/internal/ipc"
)

// Daemon holds the running daemon state.
type Daemon struct {
	ConfigDir string
	Profile   config.Profile
	Connector db.Connector
	Server    *ipc.Server
	Meta      *Meta
	logFile   *os.File
}

// New creates a new Daemon instance by loading the profile from config.
func New(configDir, profileName string) (*Daemon, error) {
	cfgPath := configDir + "/config.yaml"
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	p, ok := cfg.Profiles[profileName]
	if !ok {
		return nil, fmt.Errorf("profile '%s' not found in config", profileName)
	}

	// Resolve password from env if not in profile
	if p.Password == "" {
		p.Password = os.Getenv("QUERYLI_PASSWORD")
	}

	conn, err := db.NewConnector(p)
	if err != nil {
		return nil, err
	}

	if err := conn.Connect(); err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	meta := &Meta{
		Profile:   profileName,
		DBType:    p.Type,
		Host:      formatHost(&p),
		StartTime: time.Now(),
	}

	handler := NewRequestHandler(conn, meta)
	server := ipc.NewServer(SocketPath(configDir), handler)

	return &Daemon{
		ConfigDir: configDir,
		Profile:   p,
		Connector: conn,
		Server:    server,
		Meta:      meta,
	}, nil
}

// Run starts the daemon (writes PID, meta, and begins listening).
func (d *Daemon) Run() error {
	// Write PID file
	if err := WritePID(PIDFile(d.ConfigDir)); err != nil {
		d.Connector.Close()
		return fmt.Errorf("write PID: %w", err)
	}

	// Write metadata
	if err := SaveMeta(MetaPath(d.ConfigDir), d.Meta); err != nil {
		d.Connector.Close()
		os.Remove(PIDFile(d.ConfigDir))
		return fmt.Errorf("write meta: %w", err)
	}

	// Setup logging — keep file open for daemon lifetime
	logFile, err := os.OpenFile(LogPath(d.ConfigDir), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err == nil {
		d.logFile = logFile
		log.SetOutput(logFile)
		log.Printf("Daemon started for profile '%s' (%s)", d.Meta.Profile, d.Meta.DBType)
	} else {
		log.SetOutput(io.Discard)
	}

	// Listen on socket
	return d.Server.Listen()
}

// Shutdown gracefully stops the daemon.
func (d *Daemon) Shutdown() error {
	log.Printf("Daemon shutting down")

	if err := d.Connector.Close(); err != nil {
		log.Printf("Error closing connector: %v", err)
	}
	if err := d.Server.Close(); err != nil {
		log.Printf("Error closing server: %v", err)
	}
	os.Remove(PIDFile(d.ConfigDir))
	os.Remove(MetaPath(d.ConfigDir))
	if d.logFile != nil {
		d.logFile.Close()
	}
	return nil
}

func formatHost(p *config.Profile) string {
	switch p.Type {
	case "sqlite":
		return p.Path
	case "oracle":
		return fmt.Sprintf("%s:%d/%s", p.Host, p.Port, p.Service)
	default:
		return fmt.Sprintf("%s:%d/%s", p.Host, p.Port, p.DBName)
	}
}
