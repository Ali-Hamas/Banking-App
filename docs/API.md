# API — Auth Middleware

Base URL (dev): `http://localhost:4000`
All bodies are JSON.

## POST `/api/session`

Called by the ATM when the user picks **Cardless Cash Withdrawal**.

Request:
```json
{ "atmId": "ATM-DEMO-001" }
```

Response:
```json
{
  "sessionId": "uuid",
  "atmId": "ATM-DEMO-001",
  "expiresAt": 1731340000000,
  "ttlSeconds": 60,
  "qrToken": "eyJhbGciOi..."
}
```

The ATM renders `qrToken` as a QR code and opens a WebSocket to `/ws?sid={sessionId}`.

## GET `/api/session/:id`

Polling fallback if the ATM cannot hold a WebSocket. Returns the current session state.

```json
{
  "id": "uuid",
  "atmId": "ATM-DEMO-001",
  "status": "pending | approved | consumed | expired | rejected",
  "createdAt": 1731339940000,
  "expiresAt": 1731340000000
}
```

## POST `/api/session/approve`

Called by the bank app after the customer scans the QR and confirms with the in-app biometric prompt.

Request:
```json
{
  "qrToken": "eyJhbGciOi...",
  "deviceId": "trusted-banking-device-id",
  "customerRef": "opaque-customer-reference"
}
```

Responses:
- `200` — `{ "ok": true, "session": { ... } }`
- `400` — `invalid_token` / `missing_fields`
- `404` — `session_not_found`
- `409` — `session_not_pending` (already used or rejected)
- `410` — `session_expired`

On success, the middleware pushes `{ "type": "approved", "session": { ... } }` over the WebSocket subscribed to that session.

## POST `/api/session/:id/consume`

Called by the ATM when it has handed control to the existing withdrawal flow. Marks the session terminal so it cannot be reused.

## GET `/api/audit`

Returns the in-memory audit log (dev-only; in production this is a SIEM sink).

## WebSocket `/ws?sid={sessionId}`

The ATM connects after creating a session. Messages:
- `{"type":"subscribed","sid":"..."}` — connection acknowledged
- `{"type":"approved","session":{...}}` — bank app approved
- `{"type":"consumed","session":{...}}` — ATM has consumed the session
