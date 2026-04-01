// Package static embeds the compiled frontend assets.
package static

import "embed"

//go:embed dist
var Assets embed.FS
