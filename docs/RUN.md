# Running the demo locally

You'll run three things: the backend, the ATM simulator (browser), and the mobile app (Expo).

## Prereqs

- Node.js 20+ and npm
- For the mobile app on a real device: the Expo Go app from the Play Store / App Store
- Both your phone and your laptop on the same Wi-Fi

## 1. Backend (Go)

Install Go 1.22+ from https://go.dev/dl/, then:

```bash
cd backend-go
go mod tidy
go run ./cmd/server
```

You should see:
```
auth middleware listening on :4000
ws endpoint  ws://localhost:4000/ws?sid=<sessionId>
```

## 2. ATM simulator

In a second terminal:

```bash
cd atm-simulator
npm install
npm run dev
```

Open the URL Vite prints (usually `http://localhost:5173`). You'll see the ATM screen.

## 3. Mobile app (Expo)

In a third terminal:

```bash
cd mobile-app
npm install
npx expo start
```

### Pointing the app at your backend

The mobile app defaults to `http://10.0.2.2:4000`, which is the Android emulator's loopback to your host machine. If you're running on:

- **A real phone with Expo Go** — open `mobile-app/App.tsx` and change `API_BASE` to `http://<your-laptop-LAN-IP>:4000` (e.g. `http://192.168.1.42:4000`)
- **iOS simulator** — change `API_BASE` to `http://localhost:4000`

Scan the Expo QR with your phone (Android: Expo Go's scanner; iOS: the Camera app).

## 4. The demo

1. In the ATM browser tab: click **Continue** → **Cardless Cash Withdrawal**. A QR appears with a 60s timer.
2. In the bank-app on your phone: tap **Cardless ATM Access** → grant camera permission → point at the QR on the laptop screen.
3. Tap **Approve with biometrics (mocked)** in the app.
4. The ATM screen flips to **Authenticated** within a second (WebSocket push).
5. Pick an amount → ATM shows "Please take your cash" and resets.

## Troubleshooting

- **App stuck on "Authorising…"** — the phone can't reach the backend. Check `API_BASE` is your laptop's LAN IP and that your laptop firewall allows port 4000.
- **QR scan does nothing** — make sure the ATM tab is fully visible on screen (not minimised). Some autofocus needs a moment.
- **"session_expired"** — the QR expires in 60 seconds by default. Set `SESSION_TTL_SECONDS=120` in the backend's environment if you need longer for live demos.
