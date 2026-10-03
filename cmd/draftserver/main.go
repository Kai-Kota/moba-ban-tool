// draftserver は BAN/PICK オーバーレイのサーバー。
package main

import (
	"flag"
	"log"
	"net/http"

	"moba-ban-tool/internal/config"
	"moba-ban-tool/internal/draft"
	"moba-ban-tool/internal/hub"
	"moba-ban-tool/internal/server"
	"moba-ban-tool/internal/store"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "待ち受けアドレス(別PCから操作する場合は 0.0.0.0:8080)")
	dataDir := flag.String("data", "data", "characters.json などを置くディレクトリ")
	statePath := flag.String("state", "state.json", "状態の保存先")
	flag.Parse()

	cfg, err := config.Load(*dataDir)
	if err != nil {
		log.Fatal(err)
	}
	engine := draft.NewEngine(cfg.DraftOrder)

	st := store.New(*statePath)
	state, ok, err := st.Load()
	switch {
	case err != nil:
		log.Printf("%s を読めないため初期状態で起動します: %v", *statePath, err)
		state = engine.Initial()
	case ok:
		engine.Normalize(&state)
		log.Printf("%s から状態を復元しました", *statePath)
	default:
		state = engine.Initial()
	}

	h := hub.New(engine, state, st.Save)
	log.Printf("listening on http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, server.New(cfg, h, *dataDir)))
}
