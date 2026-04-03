package web

import "embed"

// Static holds the built React app. Run `make build` before `go build`
// to ensure web/dist exists.
//
//go:embed dist
var Static embed.FS
