package catalog

import (
	"io/fs"
)

func SetTemplatesFS(f fs.FS) {
	templatesFS = f
}
