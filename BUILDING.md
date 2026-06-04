# Building 88pay

88pay is a self-hosted, non-custodial crypto donation alert system (Solana-first).
Intended as a lightweight Entropy replacement for streamers. It supports on-screen alerts, and TTS.
Soon to support tiered donations

## Requirements

- **Go 1.25**
- **C compiler + SQLite development headers** (required for `github.com/mattn/go-sqlite3`)
  - **Ubuntu / Debian**: `sudo apt update && sudo apt install build-essential libsqlite3-dev`
- Git (for cloning and dependency management)

### Windows Requirements:
-  [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) (or MSYS2 / MinGW)
-  [Go 1.25](https://go.dev/dl/go1.25.11.windows-amd64.msi)

### Mac/Linux:
- **Ubuntu / Debian**: `sudo apt update && sudo apt install build-essential libsqlite3-dev`


### Linux/Mac Workflow
This project vendors all dependencies so builds remain reproducible and protected against upstream repositories disappearing.

```bash
# 1. Clone the repository
git clone https://github.com/Zennox-97/88pay.git
`cd 88pay`

# 2. (Optional but recommended) Work on a feature branch
`git checkout -b feature/your-feature-name`

# 3. Vendor dependencies (only needed once, or after changing go.mod)
`go mod tidy`
`go mod vendor`

# 4. Run the webserver
`./start.sh` or `go run main.go` 
