package simplerouter

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func Compression() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next(w, r)
				return
			}

			gw := gzip.NewWriter(w)
			defer gw.Close()

			crw := &compressedWriter{
				ResponseWriter: w,
				writer:         gw,
			}

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Del("Content-Length")

			next(crw, r)
		}
	}
}

type compressedWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w *compressedWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}
