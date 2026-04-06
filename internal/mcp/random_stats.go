package mcp

import "github.com/digitalghost404/inkandbone/internal/ruleset"

// rollStats delegates to the shared ruleset package.
func rollStats(system, archetype string) map[string]any {
	return ruleset.RollStats(system, archetype)
}
