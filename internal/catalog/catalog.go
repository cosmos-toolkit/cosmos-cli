package catalog

import (
	"io/fs"
)

// templatesFS is set by embed.go init function
var templatesFS fs.FS

type Catalog struct {
	templates map[string]fs.FS
}

func New() *Catalog {
	c := &Catalog{
		templates: make(map[string]fs.FS),
	}
	c.loadEmbeddedTemplates()
	return c
}

func (c *Catalog) loadEmbeddedTemplates() {
	if templatesFS == nil {
		return
	}

	// templatesFS root is the "templates" dir (api, worker, cli)
	entries, err := templatesFS.ReadDir(".")
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			templateType := entry.Name()
			subFS, err := fs.Sub(templatesFS, templateType)
			if err == nil {
				c.templates[templateType] = subFS
			}
		}
	}
}

func (c *Catalog) GetEmbeddedTemplate(templateType string) (fs.FS, bool) {
	fs, ok := c.templates[templateType]
	return fs, ok
}

func (c *Catalog) ListEmbeddedTypes() []string {
	types := make([]string, 0, len(c.templates))
	for t := range c.templates {
		types = append(types, t)
	}
	return types
}
