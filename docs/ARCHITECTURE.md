# Architecture — QR ATM Authentication Layer

## One-line description

A pre-authentication middleware that issues short-lived, signed QR sessions which the customer approves from inside their existing bank app. The middleware then tells the ATM "session authenticated" — and the ATM continues with its normal cash-withdrawal flow.

## What the middleware does

- Issues a unique, short-lived (60s default), signed session token to the ATM
- Validates the customer's approval coming from a trusted bank-app device
- Notifies the ATM in real time when the session is approved
- Writes an immutable audit log of every event

## What the middleware does NOT do

- Move money
- Read or modify account balances
- Talk to the core banking ledger
- Talk to the ATM cash dispenser
- Hold customer credentials, PINs, or biometric templates

That separation is the entire commercial story. The bank's existing systems keep doing what they already do.

## Flow

```
+----------+           +------------------+           +-------------+
|   ATM    |           |  Auth Middleware |           |  Bank App   |
+----------+           +------------------+           +-------------+
     |   POST /session       |                              |
     |---------------------->|                              |
     |  qrToken, sessionId   |                              |
     |<----------------------|                              |
     |  show QR + WS subscribe                              |
     |--- WS /ws?sid=...---->|                              |
     |                       |   user scans QR with app     |
     |                       |<-----------------------------|
     |                       |  POST /session/approve       |
     |                       |  { qrToken, deviceId }       |
     |                       |                              |
     |    WS push: approved  |                              |
     |<----------------------|                              |
     |  unlocks withdraw UI  |                              |
     |  POST /consume        |                              |
     |---------------------->|                              |
     |                       |                              |
```

## Components in this repo

| Component       | Stack                | Purpose                                                  |
|-----------------|----------------------|----------------------------------------------------------|
| `backend-go/`   | Go (gorilla/websocket, golang-jwt) | Session engine, JWT signing, audit log, WebSocket push   |
| `atm-simulator/`| React + Vite         | Browser-rendered ATM screen, shows QR, gets WS unlock    |
| `mobile-app/`   | React Native (Expo)  | Bank-app mockup with camera scanner + approve button     |

## Security properties of v1

| Property              | How it's enforced                                                          |
|-----------------------|-----------------------------------------------------------------------------|
| Short-lived sessions  | `expiresAt = createdAt + 60s`; checked on every read; sweeper marks expired |
| Single-use            | `approve` requires `status == pending`; status flips to `approved` then `consumed` |
| Tamper-evident QR     | JWT HS256 signed; backend re-verifies before approving                      |
| Device binding        | `approve` carries `deviceId`; recorded for audit                            |
| Real-time unlock      | WebSocket push to ATM the instant the approval lands                       |
| Replay protection     | Status state machine + single-use guarantee                                 |
| Audit                 | Every state change appended to audit log                                    |

## What's deliberately out of scope for v1

- OTP / SMS fallback
- Face-match / liveness
- NFC tap
- Production-grade key management (HSM)
- Persistence (in-memory store; Redis swap is one file)
- Bank-core API integration (we stop at "session authenticated")

## What changes between MVP and a real bank pilot

| Area              | MVP today                       | Pilot                                      |
|-------------------|---------------------------------|--------------------------------------------|
| Session store     | In-memory `Map`                 | Redis with TTL                             |
| Signing key       | Env var                         | HSM (cloud KMS or on-prem)                 |
| Transport         | HTTP + WS                       | mTLS, certificate-pinned                   |
| Audit             | In-memory append + console      | Append-only DB + SIEM sink                 |
| ATM connector     | Browser simulator               | Vendor-specific connector (NCR, Diebold)   |
| App side          | Standalone Expo mockup          | SDK embedded into the bank's existing app  |
