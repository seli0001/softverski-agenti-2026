package persistence

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
)

type Journal interface {
	Append(id string, event any)
	Events(id string) []any
}

type InMemoryJournal struct {
	mu     sync.Mutex
	events map[string][]any
}

func NewInMemoryJournal() *InMemoryJournal {
	return &InMemoryJournal{
		events: make(map[string][]any),
	}
}

func (j *InMemoryJournal) Append(id string, event any) {
	j.mu.Lock()
	j.events[id] = append(j.events[id], event)
	j.mu.Unlock()
}

func (j *InMemoryJournal) Events(id string) []any {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.events[id]
}

type FileJournal struct {
	mu  sync.Mutex
	dir string
}

func NewFileJournal(dir string) *FileJournal {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil
	}
	return &FileJournal{
		dir: dir,
	}
}

func (j *FileJournal) path(id string) string {
	return filepath.Join(j.dir, id)
}

func (j *FileJournal) Append(id string, event any) {
	j.mu.Lock()
	defer j.mu.Unlock()
	events := j.read(id)
	events = append(events, event)
	j.write(id, events)
}

func (j *FileJournal) Events(id string) []any {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.read(id)
}

func (j *FileJournal) read(id string) []any {
	file, err := os.Open(j.path(id))
	if err != nil {
		return nil
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	var events []any
	err1 := gob.NewDecoder(file).Decode(&events)
	if err1 != nil {
		return nil
	}
	return events

}

func (j *FileJournal) write(id string, events []any) {
	file, err := os.Create(j.path(id))
	if err != nil {
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	err1 := gob.NewEncoder(file).Encode(events)
	if err1 != nil {
		return
	}
}
