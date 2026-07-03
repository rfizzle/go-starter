//go:build tools

package main

// Import to keep it in go.mod.
//
// Note: sqlc is intentionally NOT listed here. Its dependency tree conflicts
// with this module's (ambiguous genproto imports) and cannot be vendored
// cleanly, so sqlc is installed as a standalone binary and invoked by the
// Makefile's `sqlc` target rather than via `go run`. See sqlc.yaml.
import _ "github.com/ogen-go/ogen/cmd/ogen"
