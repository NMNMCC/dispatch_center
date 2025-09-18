package database

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type DB[T any] struct {
	mu   sync.RWMutex
	path string
	data *T
}

func Open[T any](path string) (*DB[T], error) {
	bs, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &DB[T]{
			data: new(T),
			path: path,
		}, nil
	} else if err != nil {
		return nil, err
	}

	var val T
	if err := json.Unmarshal(bs, &val); err != nil {
		return nil, err
	}

	return &DB[T]{
		data: &val,
		path: path,
	}, nil
}

func (db *DB[T]) Save() error {
	bs, err := json.Marshal(db.data)
	if err != nil {
		return err
	}

	// Ensure target directory exists
	dir := filepath.Dir(db.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	// Write to a temp file in the same directory for atomic rename
	f, err := os.CreateTemp(dir, "tmp-*.json")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()

	if _, err := f.Write(bs); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return os.Rename(f.Name(), db.path)
}

func (db *DB[T]) Read(fn func(data *T) error) error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return fn(db.data)
}

func (db *DB[T]) Write(fn func(data *T) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := fn(db.data); err != nil {
		return err
	}

	return db.Save()
}
