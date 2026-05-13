package session

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusConsumed Status = "consumed"
	StatusExpired  Status = "expired"
	StatusRejected Status = "rejected"
)

// Session is the on-the-wire shape consumed by the ATM simulator and mobile app.
// JSON tags MUST match the Node implementation exactly (camelCase, omitempty on optionals).
type Session struct {
	ID                    string `json:"id"`
	AtmID                 string `json:"atmId"`
	Status                Status `json:"status"`
	CreatedAt             int64  `json:"createdAt"`
	ExpiresAt             int64  `json:"expiresAt"`
	ApprovedAt            int64  `json:"approvedAt,omitempty"`
	ConsumedAt            int64  `json:"consumedAt,omitempty"`
	ApprovedByDeviceID    string `json:"approvedByDeviceId,omitempty"`
	ApprovedByCustomerRef string `json:"approvedByCustomerRef,omitempty"`
}

type QrPayload struct {
	Sid string `json:"sid"`
	Atm string `json:"atm"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

type AuditEntry struct {
	Ts        int64                  `json:"ts"`
	SessionID string                 `json:"sessionId"`
	Event     string                 `json:"event"`
	Detail    map[string]interface{} `json:"detail,omitempty"`
}
