//go:build embed

// Package web は Vite のビルド成果物(web/dist)を Go バイナリに埋め込む。
// 埋め込みは配布用ビルド(-tags embed)のときだけ行う。
// 開発中は dist がなくても go run できるよう、通常は noembed.go が使われる。
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Dist は埋め込んだフロントエンドを返す。ルートが dist/ の中身になる。
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
