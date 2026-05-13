# atm-auth (Go)

Go rewrite of the QR-based ATM cardless authentication middleware. API-compatible
with the original Node/TypeScript backend — drop-in replacement for the React ATM
simulator and the React Native mobile app.

## Why Go

- Single static binary, no `node_modules` supply-chain surface
- Memory safety, strong typing, explicit error handling
- Goroutines model "one pending QR session waiting for approval" naturally
- Stdlib + 4 well-vetted libraries (gorilla/websocket, golang-jwt, google/uuid, rs/cors)

## Run

```bash
cd backend-go
go mod tidy
go run ./cmd/server
```

Expect:

```
auth middleware listening on :4000
ws endpoint  ws://localhost:4000/ws?sid=<sessionId>
```

## Env

| Var                   | Default                  |
| --------------------- | ------------------------ |
| `PORT`                | `4000`                   |
| `JWT_SECRET`          | `dev-secret-change-me`   |
| `SESSION_TTL_SECONDS` | `60`                     |

## API

Same as the Node version. See `docs/API.md` in the repo root.

| Method | Path                          | Purpose                          |
| ------ | ----------------------------- | -------------------------------- |
| POST   | `/api/session`                | Create a pending QR session       |
| GET    | `/api/session/{id}`           | Read session state                |
| POST   | `/api/session/approve`        | Mobile app approves a scanned QR  |
| POST   | `/api/session/{id}/consume`   | ATM consumes session post-dispense|
| GET    | `/api/audit`                  | All audit entries (demo only)     |
| GET    | `/api/health`                 | Liveness + nonce                  |
| WS     | `/ws?sid=<sessionId>`         | Push updates for one session      |

## Build a release binary

```bash
go build -o atm-auth ./cmd/server
./atm-auth
```

On Windows: `go build -o atm-auth.exe ./cmd/server`.

## Layout

```
backend-go/
  cmd/server/main.go        — entrypoint + graceful shutdown
  internal/
    config/                 — env loader
    session/                — types, store, JWT sign/verify
    realtime/               — gorilla/websocket hub
    httpapi/                — handlers + router + CORS
```
