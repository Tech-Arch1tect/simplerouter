package simplerouter

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompression(t *testing.T) {
	router := New().Use(Compression())

	content := strings.Repeat("This content will be compressed. ", 50)

	router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(content))
	})

	tests := []struct {
		name           string
		acceptEncoding string
		expectGzip     bool
	}{
		{"With gzip support", "gzip", true},
		{"With gzip and deflate", "gzip, deflate", true},
		{"No compression support", "", false},
		{"Only deflate support", "deflate", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
			}

			if tt.expectGzip {
				if rr.Header().Get("Content-Encoding") != "gzip" {
					t.Errorf("Expected gzip compression")
				}
				if rr.Header().Get("Vary") != "Accept-Encoding" {
					t.Errorf("Expected Vary header")
				}

				// Decompress and verify
				reader := bytes.NewReader(rr.Body.Bytes())
				gzipReader, err := gzip.NewReader(reader)
				if err != nil {
					t.Fatalf("Failed to create gzip reader: %v", err)
				}
				defer gzipReader.Close()

				decompressed, err := io.ReadAll(gzipReader)
				if err != nil {
					t.Fatalf("Failed to decompress: %v", err)
				}

				if string(decompressed) != content {
					t.Errorf("Decompressed content doesn't match")
				}
			} else {
				if rr.Header().Get("Content-Encoding") != "" {
					t.Errorf("Expected no compression")
				}
				if rr.Body.String() != content {
					t.Errorf("Uncompressed content doesn't match")
				}
			}
		})
	}
}
