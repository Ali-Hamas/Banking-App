package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/demobank/atm-auth/internal/config"
	"github.com/demobank/atm-auth/internal/realtime"
	"github.com/demobank/atm-auth/internal/session"
	"github.com/google/uuid"
)

type Handlers struct {
	Cfg   config.Config
	Store *session.Store
	Hub   *realtime.Hub
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code string, extra map[string]interface{}) {
	out := map[string]interface{}{"error": code}
	for k, v := range extra {
		out[k] = v
	}
	writeJSON(w, status, out)
}

func nowMillis() int64 { return time.Now().UnixMilli() }

// POST /api/session
func (h *Handlers) CreateSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AtmID string `json:"atmId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	atmID := body.AtmID
	if atmID == "" {
		atmID = "ATM-DEMO-001"
	}

	now := nowMillis()
	s := &session.Session{
		ID:        uuid.NewString(),
		AtmID:     atmID,
		Status:    session.StatusPending,
		CreatedAt: now,
		ExpiresAt: now + int64(h.Cfg.SessionTTLSeconds)*1000,
	}
	h.Store.Put(s)

	payload := session.QrPayload{
		Sid: s.ID,
		Atm: s.AtmID,
		Iat: now / 1000,
		Exp: s.ExpiresAt / 1000,
	}
	qrToken, err := session.SignQR(h.Cfg.JWTSecret, payload)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sign_failed", nil)
		return
	}

	h.Store.Log(session.AuditEntry{
		Ts: now, SessionID: s.ID, Event: "session.created",
		Detail: map[string]interface{}{"atmId": atmID},
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sessionId":  s.ID,
		"atmId":      s.AtmID,
		"expiresAt":  s.ExpiresAt,
		"ttlSeconds": h.Cfg.SessionTTLSeconds,
		"qrToken":    qrToken,
	})
}

// GET /api/session/{id}
func (h *Handlers) GetSession(w http.ResponseWriter, r *http.Request) {
	id := pathTail(r.URL.Path, "/api/session/")
	if id == "" || strings.Contains(id, "/") {
		writeErr(w, http.StatusNotFound, "not_found", nil)
		return
	}
	s, ok := h.Store.Get(id)
	if !ok {
		writeErr(w, http.StatusNotFound, "not_found", nil)
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// POST /api/session/approve
func (h *Handlers) ApproveSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		QRToken     string `json:"qrToken"`
		DeviceID    string `json:"deviceId"`
		CustomerRef string `json:"customerRef"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "missing_fields", nil)
		return
	}
	if body.QRToken == "" || body.DeviceID == "" {
		writeErr(w, http.StatusBadRequest, "missing_fields", nil)
		return
	}

	payload, err := session.VerifyQR(h.Cfg.JWTSecret, body.QRToken)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_token", nil)
		return
	}

	s, ok := h.Store.Get(payload.Sid)
	if !ok {
		writeErr(w, http.StatusNotFound, "session_not_found", nil)
		return
	}
	if s.Status != session.StatusPending {
		writeErr(w, http.StatusConflict, "session_not_pending",
			map[string]interface{}{"status": string(s.Status)})
		return
	}
	if nowMillis() > s.ExpiresAt {
		h.Store.Update(s.ID, func(x *session.Session) { x.Status = session.StatusExpired })
		writeErr(w, http.StatusGone, "session_expired", nil)
		return
	}

	customerRef := body.CustomerRef
	if customerRef == "" {
		customerRef = "anon-customer"
	}

	updated, _ := h.Store.Update(s.ID, func(x *session.Session) {
		x.Status = session.StatusApproved
		x.ApprovedAt = nowMillis()
		x.ApprovedByDeviceID = body.DeviceID
		x.ApprovedByCustomerRef = customerRef
	})

	h.Store.Log(session.AuditEntry{
		Ts: nowMillis(), SessionID: s.ID, Event: "session.approved",
		Detail: map[string]interface{}{
			"deviceId":    body.DeviceID,
			"customerRef": customerRef,
		},
	})

	h.Hub.Notify(updated, "approved")
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "session": updated})
}

// POST /api/session/{id}/consume
func (h *Handlers) ConsumeSession(w http.ResponseWriter, r *http.Request) {
	rest := pathTail(r.URL.Path, "/api/session/")
	id := strings.TrimSuffix(rest, "/consume")
	if id == "" || strings.Contains(id, "/") {
		writeErr(w, http.StatusNotFound, "not_found", nil)
		return
	}
	s, ok := h.Store.Get(id)
	if !ok {
		writeErr(w, http.StatusNotFound, "not_found", nil)
		return
	}
	if s.Status != session.StatusApproved {
		writeErr(w, http.StatusConflict, "not_approved", nil)
		return
	}
	updated, _ := h.Store.Update(s.ID, func(x *session.Session) {
		x.Status = session.StatusConsumed
		x.ConsumedAt = nowMillis()
	})
	h.Store.Log(session.AuditEntry{Ts: nowMillis(), SessionID: s.ID, Event: "session.consumed"})
	h.Hub.Notify(updated, "consumed")
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

// GET /api/audit
func (h *Handlers) Audit(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.Store.Audit())
}

// GET /api/health
func (h *Handlers) Health(w http.ResponseWriter, _ *http.Request) {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"ts":    nowMillis(),
		"nonce": hex.EncodeToString(b),
	})
}

func pathTail(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return path[len(prefix):]
}
