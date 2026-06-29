package persistence

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
)

type SnapshotStore interface {
	Save(id string, seq int, state any)
	Load(id string) (state any, seq int, ok bool)
}

type snap struct {
	Seq   int
	State any
}

type FileSnapshotStore struct {
	mu  sync.Mutex
	dir string
}

func NewFileSnapshotStore(dir string) *FileSnapshotStore {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil
	}
	return &FileSnapshotStore{dir: dir}
}

func (s *FileSnapshotStore) path(id string) string {
	return filepath.Join(s.dir, id+".snap")
}

func (s *FileSnapshotStore) Save(id string, seq int, state any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Create(s.path(id))
	if err != nil {
		return
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(f)
	err1 := gob.NewEncoder(f).Encode(snap{Seq: seq, State: state})
	if err1 != nil {
		return
	}
}

func (s *FileSnapshotStore) Load(id string) (state any, seq int, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(s.path(id))
	if err != nil {
		return
	}
	defer f.Close()
	var sn snap
	if err := gob.NewDecoder(f).Decode(&sn); err != nil {
		return nil, 0, false
	}
	return sn.State, sn.Seq, true
}
