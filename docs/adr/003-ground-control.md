# ADR-003: Ground Control — Physical Dashboard

## Status: Active

## Context

The Visionary needs ambient awareness of system state without constantly checking tabs. Ground Control is a Stream Deck plugin that makes the invisible visible — physical buttons that glow, pulse, and flash.

## Repository

`github.com/drdanmaggs/ground-control` (Node.js, Elgato SDK)

Rocket Fuel exposes state via `rf streamdeck serve` (Go, HTTP/WebSocket). Ground Control is pure UI — all logic stays in Go.

## Architecture

```
Stream Deck App ←WebSocket→ Ground Control plugin (JS) ←HTTP/WS→ rf streamdeck serve (Go)
```

## Features

- **Worker grid**: each button = one worker slot. Pulsing blue (working), green (PR ready), red (CI failing), amber (stuck)
- **Meeting button**: call a meeting with the Integrator. Pulses amber when the Integrator needs the Visionary
- **Quick actions**: launch, land, dispatch now
- **Board status**: pipeline counts (Ready: 3, In Progress: 2, Review: 1)

## Without Stream Deck

When the Visionary doesn't have the Stream Deck (laptop, travel), fall back to:
- **iTerm2 notifications** via `osascript` (macOS native notification center)
- **tmux bell** — triggers iTerm2's attention indicator on the tab
- **Dashboard pane** — the split-pane status display in the integrator tab

The notification system should be layered: Stream Deck → macOS notification → tmux bell. Whatever's available gets used.
