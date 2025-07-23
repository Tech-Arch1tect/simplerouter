package simplerouter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type AccessLogFormat int

const (
	JSONLogFormat AccessLogFormat = iota
	CombinedLogFormat
)

type AccessLogConfig struct {
	Output io.Writer
	Format AccessLogFormat
}

type AccessLogEntry struct {
	RemoteAddr string    `json:"remote_addr"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Status     int       `json:"status"`
	Size       int       `json:"size"`
	UserAgent  string    `json:"user_agent"`
	Referer    string    `json:"referer"`
	Duration   int64     `json:"duration_ms"`
	Timestamp  time.Time `json:"timestamp"`
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("responseWriter does not implement http.Hijacker")
}

func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rw *responseWriter) CloseNotify() <-chan bool {
	if notifier, ok := rw.ResponseWriter.(http.CloseNotifier); ok {
		return notifier.CloseNotify()
	}
	return make(<-chan bool)
}

func (rw *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return fmt.Errorf("responseWriter does not implement http.Pusher")
}

func AccessLogging(config AccessLogConfig) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &responseWriter{
				ResponseWriter: w,
				status:         0,
				size:           0,
			}

			next(wrapped, r)

			duration := time.Since(start)
			entry := AccessLogEntry{
				RemoteAddr: r.RemoteAddr,
				Method:     r.Method,
				Path:       r.URL.Path,
				Status:     wrapped.status,
				Size:       wrapped.size,
				UserAgent:  r.UserAgent(),
				Referer:    r.Referer(),
				Duration:   duration.Milliseconds(),
				Timestamp:  start,
			}

			switch config.Format {
			case JSONLogFormat:
				logJSON(config.Output, entry)
			case CombinedLogFormat:
				logCombined(config.Output, entry)
			}
		}
	}
}

func logJSON(output io.Writer, entry AccessLogEntry) {
	data, _ := json.Marshal(entry)
	fmt.Fprintf(output, "%s\n", data)
}

func logCombined(output io.Writer, entry AccessLogEntry) {
	timestamp := entry.Timestamp.Format("02/Jan/2006:15:04:05 -0700")
	fmt.Fprintf(output, "%s - - [%s] \"%s %s HTTP/1.1\" %d %d \"%s\" \"%s\" %dms\n",
		entry.RemoteAddr,
		timestamp,
		entry.Method,
		entry.Path,
		entry.Status,
		entry.Size,
		entry.Referer,
		entry.UserAgent,
		entry.Duration,
	)
}
