package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
)

// optStr returns a string argument value, or "" if absent or nil.
func optStr(req mcplib.CallToolRequest, key string) string {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok || v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// reqStr returns a string argument value and true, or ("", false) if absent.
func reqStr(req mcplib.CallToolRequest, key string) (string, bool) {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok || v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// optInt64 returns a numeric argument as int64, or (0, false) if absent.
// JSON numbers unmarshal as float64; this converts safely.
func optInt64(req mcplib.CallToolRequest, key string) (int64, bool) {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok || v == nil {
		return 0, false
	}
	f, ok := v.(float64)
	return int64(f), ok
}

// optBool returns a boolean argument value, or false if absent.
func optBool(req mcplib.CallToolRequest, key string) bool {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok || v == nil {
		return false
	}
	b, _ := v.(bool)
	return b
}

// optFloat64 returns a numeric argument as float64, or (0, false) if absent.
func optFloat64(req mcplib.CallToolRequest, key string) (float64, bool) {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok || v == nil {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}
