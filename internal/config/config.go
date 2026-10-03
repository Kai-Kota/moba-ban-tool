// Package config はタイトルごとの設定ファイル(data/ 以下)を読み込む。
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

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

// CharacterIDs はキャラIDの一覧を返す。
func (c *Config) CharacterIDs() []string {
	ids := make([]string, len(c.Characters))
	for i, ch := range c.Characters {
		ids[i] = ch.ID
	}
	return ids
}

// Load は dataDir から characters.json と draft_order.json を読んで検証する。
// 画像ファイルが見つからない場合は致命的ではないので警告だけ出す。
func Load(dataDir string) (*Config, error) {
	var c Config
	if err := readJSON(filepath.Join(dataDir, "characters.json"), &c.Characters); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dataDir, "draft_order.json"), &c.DraftOrder); err != nil {
		return nil, err
	}
	if err := validateCharacters(c.Characters); err != nil {
		return nil, fmt.Errorf("characters.json: %w", err)
	}
	if err := validateOrder(c.DraftOrder); err != nil {
		return nil, fmt.Errorf("draft_order.json: %w", err)
	}
	for _, ch := range c.Characters {
		if ch.Image == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(dataDir, "images", filepath.FromSlash(ch.Image))); err != nil {
			log.Printf("警告: %s の画像 images/%s が見つかりません", ch.ID, ch.Image)
		}
	}
	return &c, nil
}

func validateCharacters(chars []Character) error {
	if len(chars) == 0 {
		return errors.New("キャラが1体もありません")
	}
	seen := map[string]bool{}
	for i, ch := range chars {
		if ch.ID == "" {
			return fmt.Errorf("%d番目のキャラに id がありません", i+1)
		}
		if seen[ch.ID] {
			return fmt.Errorf("id %q が重複しています", ch.ID)
		}
		seen[ch.ID] = true
		if strings.Contains(ch.Image, "..") {
			return fmt.Errorf("%s の image に .. は使えません", ch.ID)
		}
	}
	return nil
}

func validateOrder(order []draft.Step) error {
	if len(order) == 0 {
		return errors.New("ステップが1つもありません")
	}
	picks := map[draft.Side]int{}
	for i, st := range order {
		if st.Team != draft.Blue && st.Team != draft.Red {
			return fmt.Errorf("%d番目: team は blue か red にしてください(%q)", i+1, st.Team)
		}
		if st.Type != draft.Ban && st.Type != draft.Pick {
			return fmt.Errorf("%d番目: type は ban か pick にしてください(%q)", i+1, st.Type)
		}
		if st.Type == draft.Pick {
			picks[st.Team]++
		}
	}
	for side, n := range picks {
		if n > draft.PlayerCount {
			return fmt.Errorf("%s のピックが %d 回あります(最大 %d)", side, n, draft.PlayerCount)
		}
	}
	return nil
}

var imageExts = []string{".png", ".jpg", ".jpeg", ".webp", ".svg", ".gif"}

// Logos は dataDir/logos にある画像ファイル名の一覧を返す。
func Logos(dataDir string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(dataDir, "logos"))
	if errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	logos := []string{}
	for _, e := range entries {
		if !e.IsDir() && slices.Contains(imageExts, strings.ToLower(filepath.Ext(e.Name()))) {
			logos = append(logos, e.Name())
		}
	}
	return logos, nil
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
