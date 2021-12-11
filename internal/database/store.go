package database

import (
	"github.com/denismitr/lemon"
	"github.com/pkg/errors"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var ErrInvalidDatabaseName = errors.New("invalid database name")

const (
	baseDir = "data/" // move to config
	ext     = ".ldb"
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

	fullDBPath, err := createFullDBPath(name)
	if err != nil {
		return nil, err
	}

	db, closer, err := lemon.Open(fullDBPath)
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

var validDBNameRegEx = regexp.MustCompile(`^[0-9a-zA-Z_-]{1,120}$`)

func createFullDBPath(name string) (string, error) {
	if !validDBNameRegEx.MatchString(name) {
		return "", ErrInvalidDatabaseName
	}

	name = filepath.Base(name)
	name = strings.TrimSuffix(name, ext)
	return "./" + filepath.Join(baseDir, name+ext), nil
}
