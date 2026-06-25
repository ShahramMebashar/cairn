// Package connect turns "point this agent at cairn" into one call: it knows where each
// AI coding agent keeps its MCP config and in what shape, and merges a cairn server entry
// in without disturbing the rest of the file. The cairn process itself does the writing —
// it already runs locally with the user's permissions — so the same code serves the
// desktop app and `cairn web` in a browser.
package connect

import (
	"bytes"
	"encoding/json"
	"fmt"

	toml "github.com/pelletier/go-toml/v2"
)

// serverConfig is the resolved cairn MCP entry written into an agent's config file.
type serverConfig struct {
	Name string   // server key, always serverName
	Bin  string   // absolute path to the cairn binary (sidecar-safe)
	Args []string // launch args after the binary
}

// argv is bin + args as one slice, for formats whose command is an array (OpenCode).
func (c serverConfig) argv() []string {
	return append([]string{c.Bin}, c.Args...)
}

// format reads and writes one agent-config shape. upsert merges the cairn entry into
// existing bytes (possibly empty) preserving everything else; connected reports whether a
// cairn entry is already present; lang names the syntax for the manual-guide snippet.
type format interface {
	upsert(existing []byte, c serverConfig) ([]byte, error)
	remove(existing []byte) ([]byte, error) // drop only the cairn entry, keep the rest
	connected(existing []byte) bool
	lang() string
}

// mcpServersJSON is the common `{ "mcpServers": { "<name>": { command, args } } }` shape
// used by Claude Code (.mcp.json), Cursor (.cursor/mcp.json) and Windsurf (mcp_config.json).
type mcpServersJSON struct{}

func (mcpServersJSON) lang() string { return "json" }

func (mcpServersJSON) upsert(existing []byte, c serverConfig) ([]byte, error) {
	root, err := decodeJSON(existing)
	if err != nil {
		return nil, err
	}
	servers, _ := root["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers[c.Name] = map[string]any{"command": c.Bin, "args": c.Args}
	root["mcpServers"] = servers
	return encodeJSON(root)
}

func (mcpServersJSON) remove(existing []byte) ([]byte, error) {
	return jsonNestedDelete(existing, "mcpServers")
}

func (mcpServersJSON) connected(existing []byte) bool {
	return jsonNestedHas(existing, "mcpServers", serverName)
}

// openCodeJSON writes OpenCode's `{ "mcp": { "<name>": { type:"local", command:[...] } } }`,
// where command is the full argv. It seeds the $schema for fresh files.
type openCodeJSON struct{}

func (openCodeJSON) lang() string { return "json" }

func (openCodeJSON) upsert(existing []byte, c serverConfig) ([]byte, error) {
	root, err := decodeJSON(existing)
	if err != nil {
		return nil, err
	}
	if _, ok := root["$schema"]; !ok {
		root["$schema"] = "https://opencode.ai/config.json"
	}
	mcp, _ := root["mcp"].(map[string]any)
	if mcp == nil {
		mcp = map[string]any{}
	}
	mcp[c.Name] = map[string]any{"type": "local", "command": c.argv(), "enabled": true}
	root["mcp"] = mcp
	return encodeJSON(root)
}

func (openCodeJSON) remove(existing []byte) ([]byte, error) {
	return jsonNestedDelete(existing, "mcp")
}

func (openCodeJSON) connected(existing []byte) bool {
	return jsonNestedHas(existing, "mcp", serverName)
}

// codexTOML writes `[mcp_servers.<name>]` into Codex's config.toml, preserving other tables.
type codexTOML struct{}

func (codexTOML) lang() string { return "toml" }

func (codexTOML) upsert(existing []byte, c serverConfig) ([]byte, error) {
	root := map[string]any{}
	if len(bytes.TrimSpace(existing)) > 0 {
		if err := toml.Unmarshal(existing, &root); err != nil {
			return nil, fmt.Errorf("parse existing config: %w", err)
		}
	}
	servers, _ := root["mcp_servers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers[c.Name] = map[string]any{"command": c.Bin, "args": c.Args}
	root["mcp_servers"] = servers

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)
	if err := enc.Encode(root); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (codexTOML) remove(existing []byte) ([]byte, error) {
	root := map[string]any{}
	if len(bytes.TrimSpace(existing)) > 0 {
		if err := toml.Unmarshal(existing, &root); err != nil {
			return nil, fmt.Errorf("parse existing config: %w", err)
		}
	}
	if servers, ok := root["mcp_servers"].(map[string]any); ok {
		delete(servers, serverName)
		if len(servers) == 0 {
			delete(root, "mcp_servers") // don't leave an empty table header behind
		} else {
			root["mcp_servers"] = servers
		}
	}
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)
	if err := enc.Encode(root); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (codexTOML) connected(existing []byte) bool {
	root := map[string]any{}
	if toml.Unmarshal(existing, &root) != nil {
		return false
	}
	m, _ := root["mcp_servers"].(map[string]any)
	_, ok := m[serverName]
	return ok
}

// --- JSON helpers (2-space indent, trailing newline; tolerant of empty input) ---

func decodeJSON(existing []byte) (map[string]any, error) {
	root := map[string]any{}
	if len(bytes.TrimSpace(existing)) > 0 {
		if err := json.Unmarshal(existing, &root); err != nil {
			return nil, fmt.Errorf("parse existing config: %w", err)
		}
	}
	return root, nil
}

func encodeJSON(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// jsonNestedDelete removes root[outer][serverName], dropping the outer map if it becomes
// empty, and re-encodes — preserving every other key in the file.
func jsonNestedDelete(existing []byte, outer string) ([]byte, error) {
	root, err := decodeJSON(existing)
	if err != nil {
		return nil, err
	}
	if m, ok := root[outer].(map[string]any); ok {
		delete(m, serverName)
		if len(m) == 0 {
			delete(root, outer)
		} else {
			root[outer] = m
		}
	}
	return encodeJSON(root)
}

func jsonNestedHas(existing []byte, outer, name string) bool {
	root := map[string]any{}
	if json.Unmarshal(existing, &root) != nil {
		return false
	}
	m, _ := root[outer].(map[string]any)
	_, ok := m[name]
	return ok
}
