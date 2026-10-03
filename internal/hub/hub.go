// Package hub は WebSocket クライアントを管理し、DraftState の更新を全員へ配信する。
package hub

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"moba-ban-tool/internal/draft"
)

const writeTimeout = 5 * time.Second

// サーバーから送るメッセージ。
type stateMsg struct {
	Type  string      `json:"type"` // "state"
	State draft.State `json:"state"`
}

type errorMsg struct {
	Type    string `json:"type"` // "error"
	Message string `json:"message"`
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub は唯一の DraftState を持つ。状態の更新と配信は mu で直列化する。
type Hub struct {
	mu      sync.Mutex
	engine  *draft.Engine
	state   draft.State
	clients map[*client]struct{}
	save    func(draft.State) error
}

func New(engine *draft.Engine, initial draft.State, save func(draft.State) error) *Hub {
	return &Hub{
		engine:  engine,
		state:   initial,
		clients: map[*client]struct{}{},
		save:    save,
	}
}

// ServeHTTP は /ws のハンドラ。
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Vite の dev server 経由(localhost:5173 → :8080)を許可する
		OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
	})
	if err != nil {
		log.Printf("ws accept: %v", err)
		return
	}
	c := &client{conn: conn, send: make(chan []byte, 16)}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	first, _ := json.Marshal(stateMsg{Type: "state", State: h.state})
	c.send <- first
	h.mu.Unlock()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	go h.writeLoop(ctx, c)
	h.readLoop(ctx, c)

	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	conn.CloseNow()
}

func (h *Hub) readLoop(ctx context.Context, c *client) {
	for {
		var a draft.Action
		if err := wsjson.Read(ctx, c.conn, &a); err != nil {
			return
		}
		if err := h.apply(a); err != nil {
			b, _ := json.Marshal(errorMsg{Type: "error", Message: err.Error()})
			h.trySend(c, b)
		}
	}
}

func (h *Hub) writeLoop(ctx context.Context, c *client) {
	for {
		select {
		case <-ctx.Done():
			return
		case b := <-c.send:
			wctx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Write(wctx, websocket.MessageText, b)
			cancel()
			if err != nil {
				c.conn.CloseNow()
				return
			}
		}
	}
}

// apply は操作を適用し、成功したら保存して全員へ配信する。
func (h *Hub) apply(a draft.Action) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.engine.Apply(&h.state, a); err != nil {
		return err
	}
	if err := h.save(h.state); err != nil {
		log.Printf("state の保存に失敗: %v", err)
	}
	b, err := json.Marshal(stateMsg{Type: "state", State: h.state})
	if err != nil {
		return err
	}
	for c := range h.clients {
		h.trySend(c, b)
	}
	return nil
}

// trySend は送信キューが詰まっているクライアントを切断する(配信全体を止めないため)。
func (h *Hub) trySend(c *client, b []byte) {
	select {
	case c.send <- b:
	default:
		c.conn.CloseNow()
	}
}
