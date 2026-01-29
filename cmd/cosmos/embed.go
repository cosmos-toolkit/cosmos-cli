package main

import (
	"embed"
	"io/fs"
)

//go:embed templates
var embeddedTemplates embed.FS

func getTemplatesFS() (fs.FS, error) {
	return fs.Sub(embeddedTemplates, "templates")
}
