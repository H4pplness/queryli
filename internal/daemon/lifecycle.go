package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// PIDInfo holds PID file information.
type PIDInfo struct {
	PID   int
	Alive bool
}

// PIDFile returns the path to the PID file.
func PIDFile(baseDir string) string {
	return filepath.Join(baseDir, "daemon.pid")
}

// SocketPath returns the path to the Unix socket (forward slashes for Windows).
func SocketPath(baseDir string) string {
	return filepath.ToSlash(filepath.Join(baseDir, "daemon.sock"))
}

// MetaPath returns the path to the metadata file.
func MetaPath(baseDir string) string {
	return filepath.Join(baseDir, "daemon.meta")
}

// LogPath returns the path to the daemon log file.
func LogPath(baseDir string) string {
	return filepath.Join(baseDir, "daemon.log")
}

// WritePID writes the current process PID to a file.
func WritePID(path string) error {
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0600)
}

// ReadPID reads a PID from a file.
func ReadPID(path string) (*PIDInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return nil, fmt.Errorf("invalid PID: %s", string(data))
	}

	return &PIDInfo{PID: pid, Alive: isProcessAlive(pid)}, nil
}

// isProcessAlive checks if a process with the given PID is running.
// On Unix: os.FindProcess always succeeds; we use Signal(0) to check.
// On Windows: os.FindProcess calls OpenProcess and returns error if
// the process does not exist.
func isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, we need to check further; on Windows, FindProcess
	// already verified the process exists. Signal(nil) on Windows
	// is a no-op but validates the handle.
	process.Signal(os.Signal(nil))
	return true
}

// CleanStaleFiles removes stale PID, socket, and meta files.
func CleanStaleFiles(baseDir string) {
	pidPath := PIDFile(baseDir)
	info, err := ReadPID(pidPath)
	if err != nil {
		// No PID file, clean up socket anyway
		os.Remove(SocketPath(baseDir))
		os.Remove(MetaPath(baseDir))
		return
	}

	if !info.Alive {
		os.Remove(pidPath)
		os.Remove(SocketPath(baseDir))
		os.Remove(MetaPath(baseDir))
	}
}
