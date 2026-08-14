// Package web embeds static assets and content shipped inside the binary.
package web

import "embed"

// StaticFS holds CSS/JS/images served under /static/.
//
//go:embed static
var StaticFS embed.FS

// ContentFS holds editable static content (e.g. the About page).
//
//go:embed content
var ContentFS embed.FS
