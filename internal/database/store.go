package database

import (
	"github.com/denismitr/lemon"
	"path/filepath"
	"sync"
	"time"
)

type connection struct {
	db     *lemon.DB
	t      *time.Timer
	closer lemon.Closer
}

type Store struct {
	databases map[string]*connection
	mu        sync.Mutex
}

func NewStore() *Store {
	return &Store{
		databases: make(map[string]*connection),
	}
}

func (s *Store) Get(name string) (*lemon.DB, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.databases[name]; ok {
		c.t.Reset(10 * time.Minute)
		return c.db, nil
	}

	db, closer, err := lemon.Open(filepath.Join("./data/", name))
	if err != nil {
		return nil, err
	}

	c := connection{
		closer: closer,
		db:     db,
		t:      time.NewTimer(10 * time.Minute),
	}

	s.databases[name] = &c

	return db, nil
}
