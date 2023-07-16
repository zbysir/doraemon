package httpsrv

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

// TryFile If the file does not exist, return root index.html
// 需要结合 http.FileServer 使用
func TryFile(req *http.Request, d fs.FS) {
	// try file
	p := strings.TrimPrefix(req.URL.Path, "/")
	f, err := d.Open(p)
	if err != nil {
		f, err = d.Open(filepath.Join(p, "index.html"))
		if err != nil {
			// fallback to index.html
			req.URL.Path = ""
		} else {
			f.Close()
		}
	} else {
		f.Close()
	}
}
