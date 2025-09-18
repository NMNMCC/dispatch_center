package database

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"sync"

	"github.com/google/uuid"
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
	temp := path.Join(os.TempDir(), uuid.NewString()+".json")
	os.WriteFile(temp, bs, 0600)

	return os.Rename(temp, db.path)
}

func (db *DB[T]) Read(fn func(data *T) error) error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return fn(db.data)
}

func (db *DB[T]) Write(fn func(data *T) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	defer db.Save()
	return fn(db.data)
}
