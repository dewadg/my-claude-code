// Package migrations embeds the goose SQL migrations so they ship inside the
// binary; no goose CLI or loose files are needed at runtime.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
