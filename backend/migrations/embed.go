// Package migrations embeds the .sql files that live next to it so the binary can
// apply them on boot. The same files are pasteable into the Neon SQL console by
// hand; see MANUAL_SETUP.md.
package migrations

import "embed"

//go:embed *.sql
var files embed.FS

// FS exposes the embedded migrations as a read-only filesystem.
func FS() embed.FS { return files }
