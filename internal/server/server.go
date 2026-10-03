// Package server は HTTP のルーティングをまとめる。
package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"

	"moba-ban-tool/internal/config"
	"moba-ban-tool/internal/hub"
)

// New はルーティング済みのハンドラを返す。
//
//	/ws          WebSocket(state の配信と操作の受付)
//	/api/config  キャラ一覧とドラフト順
//	/api/logos   data/logos/ の画像ファイル名一覧
//	/assets/     data/ 以下の画像(images/, logos/)
//	/control     操作パネル  ┐
//	/overlay     OBS 表示用  ├ ui が nil(開発用ビルド)のときは案内ページを返す
//	/static/     JS / CSS    ┘
func New(cfg *config.Config, h *hub.Hub, dataDir string, ui fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /ws", h)
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, cfg)
	})
	// ロゴは配信中に追加されることがあるので、毎回フォルダを読む
	mux.HandleFunc("GET /api/logos", func(w http.ResponseWriter, r *http.Request) {
		logos, err := config.Logos(dataDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, logos)
	})
	assets := http.FileServer(http.Dir(filepath.Clean(dataDir)))
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", noDirList(noCache(assets))))

	if ui == nil {
		mux.HandleFunc("GET /", devNotice)
		return mux
	}
	mux.Handle("GET /control", page(ui, "control.html"))
	mux.Handle("GET /overlay", page(ui, "overlay.html"))
	mux.Handle("GET /static/", noDirList(immutable(http.FileServerFS(ui))))
	mux.Handle("GET /{$}", http.RedirectHandler("/control", http.StatusFound))
	return mux
}

// page は埋め込んだ HTML を返す。HTML はビルドごとに参照する JS が変わるのでキャッシュさせない。
func page(ui fs.FS, name string) http.Handler {
	return noCache(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, ui, name)
	}))
}

func devNotice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`<!doctype html><meta charset="utf-8"><title>draftserver (dev)</title>
<p>この draftserver はフロントエンドを埋め込まずにビルドされています(開発用)。</p>
<p>開発中は <code>npm --prefix web run dev</code> を起動して
<a href="http://localhost:5173/control">http://localhost:5173/control</a> を開いてください。</p>
<p>配布用の exe は <code>scripts\build.ps1</code> で作れます。</p>`))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}

// noCache は画像やHTMLを差し替えたとき、OBS 側の再読み込みですぐ反映されるようにする。
func noCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

// staticTypes は Content-Type を固定する拡張子。
// Go は Windows のレジストリから MIME タイプを読むため、環境によっては .js が text/plain になり、
// ブラウザが module スクリプトの実行を拒否する。それを避けるために明示する。
var staticTypes = map[string]string{
	".js":  "text/javascript; charset=utf-8",
	".css": "text/css; charset=utf-8",
}

// immutable は Vite がファイル名にハッシュを付けた JS / CSS 用。
func immutable(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		if ct, ok := staticTypes[path.Ext(r.URL.Path)]; ok {
			w.Header().Set("Content-Type", ct)
		}
		next.ServeHTTP(w, r)
	})
}

// noDirList はディレクトリ一覧を返さないようにする。
func noDirList(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := r.URL.Path; p == "" || p[len(p)-1] == '/' {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
