# ink & bone

A solo TTRPG companion app. Claude Code acts as your GM — this app tracks characters, campaigns, sessions, world notes, combat, maps, and dice rolls in real time.

## How it works

One binary starts everything: an MCP stdio server (Claude's GM interface) + an HTTP/WebSocket server + an embedded React UI.

```
┌──────────────┐   MCP stdio   ┌─────────────────┐   WS/HTTP   ┌───────────┐
│  Claude Code │ ◄──────────► │  ./ttrpg binary  │ ──────────► │ localhost │
│  (your GM)   │               │  SQLite + API    │             │   :7432   │
└──────────────┘               └─────────────────┘             └───────────┘
```

## Supported rulesets

- D&D 5e
- Ironsworn
- Vampire: the Masquerade
- Call of Cthulhu
- Cyberpunk Red

## Requirements

- Go 1.22+
- Node 18+
- [air](https://github.com/air-verse/air) (`go install github.com/air-verse/air@latest`)

## Usage

```bash
# Development
make dev

# Production build
make build

# Install to ~/bin
make install
```

Register as an MCP server in `~/.claude/settings.json`:

```json
"mcpServers": {
  "ttrpg": { "command": "/path/to/ttrpg" }
}
```

Open Claude Code and `localhost:7432` side by side. Start playing.
