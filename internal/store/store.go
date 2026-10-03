// Package store は DraftState を state.json に保存・復元する。
package store

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"moba-ban-tool/internal/draft"
)

// Store は1つの JSON ファイルへの保存先。
type Store struct {
	path string
}

func New(path string) *Store { return &Store{path: path} }

// Load は保存済みの状態を読む。ファイルがなければ ok=false を返す。
func (s *Store) Load() (st draft.State, ok bool, err error) {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return st, false, nil
	}
	if err != nil {
		return st, false, err
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return st, false, err
	}
	return st, true, nil
}

// Save は一時ファイルに書いてから置き換える。書き込み途中で落ちても壊れないようにするため。
func (s *Store) Save(st draft.State) error {
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".state-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path)
}
