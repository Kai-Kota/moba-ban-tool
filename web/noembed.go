//go:build !embed

package web

import "io/fs"

// Dist は開発用ビルドでは nil を返す。フロントは Vite の dev server から配信する。
func Dist() fs.FS { return nil }
