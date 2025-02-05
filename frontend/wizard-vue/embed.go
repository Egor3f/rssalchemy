package wizard_vue

import "embed"

//go:embed dist
var EmbedFS embed.FS

const FSPrefix = "dist"
