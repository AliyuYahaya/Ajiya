package gitx

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const startedFile = "ajiya/started.json"

// readStarted reads the start times (Unix seconds by ticket ID) kept in the git
// directory. A missing or corrupt file counts as no records.
func readStarted(dir string) (map[string]int64, string, error) {
	path, err := GitPath(dir, startedFile)
	if err != nil {
		return nil, "", ErrNotRepo
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	m := map[string]int64{}
	if data, err := os.ReadFile(path); err == nil && json.Unmarshal(data, &m) != nil {
		m = map[string]int64{}
	}
	return m, path, nil
}

func writeStarted(path string, m map[string]int64) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// RecordStart notes when a ticket was started, in a local file in the git
// directory that is never committed. The earliest start is kept. Outside a git
// repository it does nothing.
func RecordStart(dir, id string, at int64) error {
	m, path, err := readStarted(dir)
	if err != nil {
		return nil
	}
	if _, ok := m[id]; ok {
		return nil
	}
	m[id] = at
	return writeStarted(path, m)
}

// StartedAt returns when the ticket was started in this clone, if it was.
func StartedAt(dir, id string) (at int64, ok bool) {
	m, _, err := readStarted(dir)
	if err != nil {
		return 0, false
	}
	at, ok = m[id]
	return at, ok
}

// ClearStart forgets the start time of a ticket.
func ClearStart(dir, id string) error {
	m, path, err := readStarted(dir)
	if err != nil {
		return nil
	}
	if _, ok := m[id]; !ok {
		return nil
	}
	delete(m, id)
	return writeStarted(path, m)
}
