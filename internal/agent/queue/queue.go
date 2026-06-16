package queue

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/rksdevs/sleepguard/internal/domain"
)

// File stores events as JSONL when the cloud is unreachable.
type File struct {
	mu   sync.Mutex
	path string
}

// NewFile opens or creates the queue file path.
func NewFile(path string) (*File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	_ = f.Close()
	return &File{path: path}, nil
}

// Append adds an event to the tail of the queue.
func (q *File) Append(event domain.IngestEvent) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	f, err := os.OpenFile(q.path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	line, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return err
}

// Flush uploads queued events and keeps failures on disk.
func (q *File) Flush(upload func(domain.IngestEvent) error) (sent int, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	data, err := os.ReadFile(q.path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return 0, nil
	}

	var pending []domain.IngestEvent
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var event domain.IngestEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}
		pending = append(pending, event)
	}
	if err := scanner.Err(); err != nil {
		return sent, err
	}

	var remaining []domain.IngestEvent
	for _, event := range pending {
		if err := upload(event); err != nil {
			remaining = append(remaining, event)
			continue
		}
		sent++
	}

	return sent, rewrite(q.path, remaining)
}

func rewrite(path string, events []domain.IngestEvent) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, event := range events {
		line, err := json.Marshal(event)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			return err
		}
	}
	return nil
}
