// Package config はタイトルごとの設定ファイル(data/ 以下)を読み込む。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"moba-ban-tool/internal/draft"
)

// Character は characters.json の1要素。image は data/images/ からの相対パス。
type Character struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// Config は /api/config でフロントに渡す設定一式。
type Config struct {
	Characters []Character  `json:"characters"`
	DraftOrder []draft.Step `json:"draftOrder"`
}

// Load は dataDir から characters.json と draft_order.json を読む。
func Load(dataDir string) (*Config, error) {
	var c Config
	if err := readJSON(filepath.Join(dataDir, "characters.json"), &c.Characters); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dataDir, "draft_order.json"), &c.DraftOrder); err != nil {
		return nil, err
	}
	if len(c.DraftOrder) == 0 {
		return nil, fmt.Errorf("draft_order.json が空です")
	}
	return &c, nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s を読めません: %w", path, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s の形式が不正です: %w", path, err)
	}
	return nil
}
