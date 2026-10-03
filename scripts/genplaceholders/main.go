// genplaceholders は開発用のダミーキャラ(characters.json と SVG 画像)とダミーロゴを生成する。
//
//	go run ./scripts/genplaceholders            # data/ に書き出す
//	go run ./scripts/genplaceholders -data dir  # 別のディレクトリに書き出す
//
// 既存の characters.json は上書きされるので、本番データのあるディレクトリでは実行しないこと。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var names = []string{
	"Aldric", "Brynn", "Cassia", "Dorran", "Elowen",
	"Fenrik", "Galen", "Hesper", "Isolde", "Jarek",
	"Kaida", "Lucan", "Mireille", "Nox", "Orrin",
	"Petra", "Quill", "Rowan", "Seraph", "Talon",
}

type character struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

func main() {
	dataDir := flag.String("data", "data", "出力先ディレクトリ")
	flag.Parse()

	imgDir := filepath.Join(*dataDir, "images")
	logoDir := filepath.Join(*dataDir, "logos")
	for _, d := range []string{imgDir, logoDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			log.Fatal(err)
		}
	}

	chars := make([]character, len(names))
	for i, name := range names {
		id := strings.ToLower(name)
		file := id + ".svg"
		hue := i * 360 / len(names)
		write(filepath.Join(imgDir, file), characterSVG(name, hue))
		chars[i] = character{ID: id, Name: name, Image: file}
	}
	b, err := json.MarshalIndent(chars, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	write(filepath.Join(*dataDir, "characters.json"), string(b)+"\n")

	write(filepath.Join(logoDir, "team-alpha.svg"), logoSVG("A", "#2f7de1"))
	write(filepath.Join(logoDir, "team-omega.svg"), logoSVG("Ω", "#e13a3a"))

	log.Printf("%d 体のキャラとロゴ2つを %s に生成しました", len(chars), *dataDir)
}

func write(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		log.Fatal(err)
	}
}

// characterSVG は縦長の立ち絵風プレースホルダー。名前はオーバーレイ側で表示するので画像には入れない。
// オーバーレイのピック枠は縦長、
// 操作パネルのグリッドは正方形なので、どちらで切り抜いても頭文字が中央に残る比率にする。
func characterSVG(name string, hue int) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="360" height="640" viewBox="0 0 360 640">
  <defs>
    <linearGradient id="g" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0" stop-color="hsl(%[2]d,70%%,55%%)"/>
      <stop offset="1" stop-color="hsl(%[2]d,60%%,18%%)"/>
    </linearGradient>
  </defs>
  <rect width="360" height="640" fill="url(#g)"/>
  <circle cx="180" cy="300" r="120" fill="hsl(%[2]d,70%%,75%%)" fill-opacity="0.25"/>
  <text x="180" y="350" font-family="Segoe UI, sans-serif" font-size="160" font-weight="700"
        fill="#fff" text-anchor="middle">%[3]s</text>
</svg>
<!-- %[1]s -->
`, name, hue, name[:1])
}

func logoSVG(letter, color string) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256">
  <path d="M128 12 L232 52 L220 168 L128 244 L36 168 L24 52 Z" fill="%s" stroke="#fff" stroke-width="10"/>
  <text x="128" y="164" font-family="Segoe UI, sans-serif" font-size="120" font-weight="800"
        fill="#fff" text-anchor="middle">%s</text>
</svg>
`, color, letter)
}
