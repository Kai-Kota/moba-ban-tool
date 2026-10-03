// Package server は HTTP のルーティングをまとめる。
package server

import (
	"encoding/json"
	"net/http"
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
func New(cfg *config.Config, h *hub.Hub, dataDir string) http.Handler {
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
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", noDirList(assets)))
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
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
