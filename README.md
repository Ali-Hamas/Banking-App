# Cardless ATM Authentication Layer — MVP

QR-based pre-authentication layer that lets bank customers unlock an ATM cash-withdrawal session via their existing bank app. No card, no new app installation, no transaction handling.

## What this is

A middleware that sits between the existing ATM and the existing mobile banking app and answers exactly one question: **is this person authorised to use this ATM session right now?**

It does not touch balances, ledgers, settlement, or ATM hardware. The bank's existing systems continue to do all of that.

## Repo layout

```
backend-go/      Go auth middleware (REST + WebSocket)
atm-simulator/   React web app simulating the ATM screen
mobile-app/      React Native (Expo) mockup of a bank app with QR scanner
docs/            Architecture, API spec, run instructions
```

## Quick start

See [docs/RUN.md](docs/RUN.md).

## Locked v1 scope

QR-only authentication. No OTP, no biometrics, no face match, no NFC — those are v2+.
