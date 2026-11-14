# GOG Achievements Manager

A small, open-source Windows tool to browse your GOG library and unlock or lock achievements — with a desktop UI inspired by Steam Achievement Manager, plus a command-line mode for scripting.

This project is intentional and transparent: it only talks to official GOG / GOGDB endpoints, reads the GOG Galaxy refresh token from your local Windows registry, and never downloads extra binaries or phones home.

---

## Features

- **Desktop GUI** — search your owned games, open a game, select achievements, bulk unlock or lock
- **CLI** — list games, list achievements, unlock or clear by ID
- **Local auth** — uses the refresh token already stored by GOG Galaxy (you must be signed in)
- **Caching** — auth and product client credentials are cached next to the executable (`auths.json`, `products.json`)

---

## Requirements

| Requirement | Details |
|---|---|
| OS | **Windows** (reads `HKCU\Software\GOG.com\Galaxy`) |
| GOG Galaxy | Installed and **signed in** at least once |
| Network | Access to `auth.gog.com`, `gameplay.gog.com`, `embed.gog.com`, `www.gogdb.org`, and GOG image CDN (GUI covers) |
| Go (from source only) | Go **1.24+** |

---

## Security & trust

This tool is meant to be auditable:

- **No telemetry** — nothing is sent except requests to GOG / GOGDB APIs needed for achievements
- **No obfuscation** — plain Go source; build it yourself from this repository
- **No bundled malware surface** — dependencies are the Go standard library plus Gio (GUI), Kong (CLI), and `golang.org/x/sys` (registry)
- **Secrets stay local** — the Galaxy refresh token is read from the Windows registry; verbose logs do **not** print token or client-secret values
- **Cache files** — `auths.json` / `products.json` may contain access tokens; keep them private and do not commit them

You are responsible for complying with GOG’s terms of service. Use at your own risk.

---

## Download (Windows)

The app is the file at the root of this repository:

**[`GOG-Achievements-Manager.exe`](./GOG-Achievements-Manager.exe)** — download it, double-click, use the interface.

Also available on [GitHub Releases](https://github.com/KillingMoonxx/GOG-Achievement-Manager-open-source-/releases).

## Quick start (GUI)

1. Sign in to **GOG Galaxy** on the same Windows account.
2. Download [`GOG-Achievements-Manager.exe`](./GOG-Achievements-Manager.exe).
3. Run the executable.
4. Use the search box to find a game in your library.
5. Click a game card.
6. Select achievements (or use **Select locked** / **Select unlocked**).
7. Click **Unlock selected** or **Lock selected**.

### GUI layout

| Area | What it does |
|---|---|
| Top bar | App title + signed-in Galaxy user ID |
| Search | Filters your library by title or product ID |
| Game grid | Owned games with cover art (loaded asynchronously) |
| Achievement list | Checkbox + locked/unlocked status + name, description, ID |
| Bulk bar | Unlock / Lock selected, select helpers, clear selection |
| Footer | Trust line: open source, local token, no telemetry |

If authentication fails, the window shows a clear error (Galaxy not signed in, registry missing, or network failure).

---

## CLI usage (optional)

The shipped product is the GUI exe above. A CLI is still available if you build from source:

```bat
go build -o gog-achievements.exe ./cmd/gog-achievements
```

### List all owned games

```bat
gog-achievements.exe
```

### List achievements for a game

```bat
gog-achievements.exe <game-id>
```

Example:

```bat
gog-achievements.exe 1558393671
```

### Unlock one or more achievements

```bat
gog-achievements.exe <game-id> <achievement-id> [achievement-id ...]
```

### Lock (clear) one or more achievements

```bat
gog-achievements.exe <game-id> -c <achievement-id> [achievement-id ...]
```

### Verbose logging

Add `-v` to any command:

```bat
gog-achievements.exe -v
gog-achievements.exe 1558393671 -v
```

---

## Build from source

```bat
git clone <this-repo-url>
cd gog_achievements
go mod download
```

### CLI

```bat
go build -o gog-achievements.exe ./cmd/gog-achievements
```

Or the legacy root entrypoint:

```bat
go build -o gog-achievements.exe .
```

### GUI

```bat
go build -o gog-achievements-gui.exe ./cmd/gog-achievements-gui
```

### Cross-compile from another OS

```bash
GOOS=windows GOARCH=amd64 go build -o gog-achievements-gui.exe ./cmd/gog-achievements-gui
GOOS=windows GOARCH=amd64 go build -o gog-achievements.exe ./cmd/gog-achievements
```

---

## How it works

1. Reads `refreshToken` from `HKCU\Software\GOG.com\Galaxy`
2. Exchanges it at `https://auth.gog.com/token` for an access token
3. Lists owned product IDs from `https://embed.gog.com/user/data/games`
4. Resolves game titles / images via [GOGDB](https://www.gogdb.org)
5. Loads per-game `clientId` / `clientSecret` from the latest listed GOGDB build (cached in `products.json`)
6. Lists or updates achievements via `https://gameplay.gog.com/clients/{clientId}/users/{userId}/achievements`

Unlock sends a `POST` with `date_unlocked` set; lock/clear sends `date_unlocked: null`.

---

## Project layout

```
.
├── cmd/gog-achievements/          # CLI
├── cmd/gog-achievements-gui/      # Desktop UI (Gio)
├── internal/                      # Shared auth, API, cache, types
├── tokens.css                     # Design tokens (Hallmark)
├── main.go / auth.go / …          # Legacy root CLI (still builds)
└── README.md
```

---

## Design

The GUI follows a custom **late-night neon** palette (GOG-inspired purple on near-black). Visual tokens live in `tokens.css` for portability; the Gio theme maps those colours at runtime.

---

## Troubleshooting

| Problem | What to try |
|---|---|
| “No refresh token” | Open GOG Galaxy, sign in, restart the app |
| Authentication failed | Confirm Galaxy is online; delete `auths.json` and retry |
| Empty library | Check network; confirm the Galaxy account owns games |
| Game opens but no achievements | Not all GOG titles expose achievements through this API |
| Unlock fails for one ID | Re-list achievements; verify the achievement ID; try `-v` on the CLI |
| Covers missing | Image CDN blocked or GOGDB missing `image_logo` — titles still work |

---

## Disclaimer

This is an unofficial community tool, not affiliated with GOG.com / CD PROJEKT. Achievement manipulation may violate platform rules. Use responsibly on accounts you own.
