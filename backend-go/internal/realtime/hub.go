package realtime

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/demobank/atm-auth/internal/session"
	"github.com/gorilla/websocket"
)

// Hub fan-outs session state changes to all WebSocket subscribers of a given session id.
type Hub struct {
	mu       sync.RWMutex
	subs     map[string]map[*conn]struct{}
	upgrader websocket.Upgrader
}

type conn struct {
	ws      *websocket.Conn
	writeMu sync.Mutex // gorilla/websocket forbids concurrent writes
}

func NewHub() *Hub {
	return &Hub{
		subs: make(map[string]map[*conn]struct{}),
		upgrader: websocket.Upgrader{
			// CORS is enforced at the HTTP layer; allow WS from any origin in dev.
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// ServeHTTP handles /ws?sid=<sessionId>. Closes with code 1008 if sid missing.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query().Get("sid")
	ws, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ws] upgrade failed: %v", err)
		return
	}
	if sid == "" {
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(1008, "missing sid"), time.Now().Add(time.Second))
		ws.Close()
		return
	}

	c := &conn{ws: ws}
	h.subscribe(sid, c)
	defer h.unsubscribe(sid, c)

	// Greeting matches Node: { type: "subscribed", sid }
	c.writeJSON(map[string]interface{}{"type": "subscribed", "sid": sid})

	// Read pump — we don't expect messages from client, just detect close.
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *Hub) subscribe(sid string, c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.subs[sid]; !ok {
		h.subs[sid] = make(map[*conn]struct{})
	}
	h.subs[sid][c] = struct{}{}
}

func (h *Hub) unsubscribe(sid string, c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.subs[sid]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.subs, sid)
		}
	}
	c.ws.Close()
}

// Notify broadcasts { type: event, session } to all subscribers of session.ID.
func (h *Hub) Notify(s *session.Session, event string) {
	h.mu.RLock()
	subs := h.subs[s.ID]
	conns := make([]*conn, 0, len(subs))
	for c := range subs {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	msg := map[string]interface{}{"type": event, "session": s}
	for _, c := range conns {
		c.writeJSON(msg)
	}
}

func (c *conn) writeJSON(v interface{}) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	if err := c.ws.WriteMessage(websocket.TextMessage, b); err != nil {
		// connection closed or broken; reader loop will clean up
		return
	}
}
