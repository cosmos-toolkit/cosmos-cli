package catalog

import (
	"io/fs"
)

// templatesFS will be initialized by calling SetTemplatesFS
var templatesFS fs.FS

func SetTemplatesFS(fs fs.FS) {
	templatesFS = fs
}
