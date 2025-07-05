package simplerouter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	router := New()
	if router == nil {
		t.Fatal("New() returned nil")
	}
	if router.mux == nil {
		t.Fatal("mux is nil")
	}
	if router.routes == nil {
		t.Fatal("routes map is nil")
	}
}

func TestBasicRouting(t *testing.T) {
	router := New()

	router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET response"))
	})

	router.POST("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("POST response"))
	})

	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/test", "GET response"},
		{"POST", "/test", "POST response"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}

		if rr.Body.String() != tt.expected {
			t.Errorf("Expected body %q, got %q", tt.expected, rr.Body.String())
		}
	}
}

func TestMethodNotAllowed(t *testing.T) {
	router := New()

	router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET response"))
	})

	router.POST("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("POST response"))
	})

	req := httptest.NewRequest("PUT", "/test", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}

	allowHeader := rr.Header().Get("Allow")
	if !strings.Contains(allowHeader, "GET") || !strings.Contains(allowHeader, "POST") {
		t.Errorf("Expected Allow header to contain GET and POST, got %q", allowHeader)
	}
}

func TestRouteGroups(t *testing.T) {
	router := New()

	apiGroup := router.Group("/api")
	apiGroup.GET("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API users"))
	})

	v1Group := apiGroup.Group("/v1")
	v1Group.GET("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API v1 status"))
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/api/users", "API users"},
		{"/api/v1/status", "API v1 status"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d for %s, got %d", http.StatusOK, tt.path, rr.Code)
		}

		if rr.Body.String() != tt.expected {
			t.Errorf("Expected body %q for %s, got %q", tt.expected, tt.path, rr.Body.String())
		}
	}
}

func TestMiddleware(t *testing.T) {
	router := New()

	// Middleware that adds a header
	headerMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			next(w, r)
		}
	}

	// Middleware that appends to response
	responseMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
			w.Write([]byte(" + middleware"))
		}
	}

	routerWithMiddleware := router.Use(headerMiddleware, responseMiddleware)

	routerWithMiddleware.GET("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Header().Get("X-Test") != "middleware" {
		t.Errorf("Expected header X-Test to be 'middleware', got %q", rr.Header().Get("X-Test"))
	}

	expected := "handler + middleware"
	if rr.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, rr.Body.String())
	}
}

func TestPathNormalization(t *testing.T) {
	router := New()

	// Test various path formats
	router.GET("test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("no slash"))
	})

	router.GET("/test2", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("with slash"))
	})

	group := router.Group("api")
	group.GET("users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("group path"))
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/test", "no slash"},
		{"/test2", "with slash"},
		{"/api/users", "group path"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d for %s, got %d", http.StatusOK, tt.path, rr.Code)
		}

		if rr.Body.String() != tt.expected {
			t.Errorf("Expected body %q for %s, got %q", tt.expected, tt.path, rr.Body.String())
		}
	}
}

func TestAllHTTPMethods(t *testing.T) {
	router := New()

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		switch method {
		case "GET":
			router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(method))
			})
		case "POST":
			router.POST("/test", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(method))
			})
		case "PUT":
			router.PUT("/test", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(method))
			})
		case "DELETE":
			router.DELETE("/test", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(method))
			})
		case "PATCH":
			router.PATCH("/test", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(method))
			})
		case "HEAD":
			router.HEAD("/test", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(method))
			})
		case "OPTIONS":
			router.OPTIONS("/test", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(method))
			})
		}
	}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/test", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d for %s, got %d", http.StatusOK, method, rr.Code)
		}

		if method != "HEAD" && rr.Body.String() != method {
			t.Errorf("Expected body %q for %s, got %q", method, method, rr.Body.String())
		}
	}
}

func TestJoinPathsEdgeCases(t *testing.T) {
	router := New()

	tests := []struct {
		name     string
		base     string
		path     string
		expected string
	}{
		{"empty base and path", "", "", "/"},
		{"empty base with path", "", "users", "/users"},
		{"empty base with slash path", "", "/users", "/users"},
		{"base with empty path", "/api", "", "/api"},
		{"base with slash and empty path", "/api/", "", "/api/"},
		{"both with slashes", "/api/", "/users", "/api/users"},
		{"base without slash, path with slash", "/api", "/users", "/api/users"},
		{"base with slash, path without slash", "/api/", "users", "/api/users"},
		{"nested paths", "/api/v1", "users/profile", "/api/v1/users/profile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.joinPaths(tt.base, tt.path)
			if result != tt.expected {
				t.Errorf("joinPaths(%q, %q) = %q, expected %q", tt.base, tt.path, result, tt.expected)
			}
		})
	}
}

func TestWithMethod(t *testing.T) {
	router := New()

	// Middleware that adds a header
	headerMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-With-Test", "applied")
			next(w, r)
		}
	}

	// Middleware that appends to response
	responseMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
			w.Write([]byte(" + with"))
		}
	}

	// Test With() method
	router.With(headerMiddleware, responseMiddleware).GET("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Header().Get("X-With-Test") != "applied" {
		t.Errorf("Expected header X-With-Test to be 'applied', got %q", rr.Header().Get("X-With-Test"))
	}

	expected := "handler + with"
	if rr.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, rr.Body.String())
	}
}

func TestDirectMiddlewareParameters(t *testing.T) {
	router := New()

	// Test middleware
	authMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Auth", "passed")
			next(w, r)
		}
	}

	logMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Log", "logged")
			next(w, r)
		}
	}

	// Test all HTTP methods with middleware parameters
	router.GET("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET"))
	}, authMiddleware, logMiddleware)

	router.POST("/post", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("POST"))
	}, authMiddleware)

	router.PUT("/put", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("PUT"))
	}, logMiddleware)

	router.DELETE("/delete", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("DELETE"))
	}, authMiddleware, logMiddleware)

	tests := []struct {
		method      string
		path        string
		expectAuth  bool
		expectLog   bool
		expectBody  string
	}{
		{"GET", "/get", true, true, "GET"},
		{"POST", "/post", true, false, "POST"},
		{"PUT", "/put", false, true, "PUT"},
		{"DELETE", "/delete", true, true, "DELETE"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d for %s %s, got %d", http.StatusOK, tt.method, tt.path, rr.Code)
		}

		if rr.Body.String() != tt.expectBody {
			t.Errorf("Expected body %q for %s %s, got %q", tt.expectBody, tt.method, tt.path, rr.Body.String())
		}

		if tt.expectAuth && rr.Header().Get("X-Auth") != "passed" {
			t.Errorf("Expected X-Auth header for %s %s", tt.method, tt.path)
		}

		if tt.expectLog && rr.Header().Get("X-Log") != "logged" {
			t.Errorf("Expected X-Log header for %s %s", tt.method, tt.path)
		}
	}
}

func TestRouteBuilderPattern(t *testing.T) {
	router := New()

	// Test middleware
	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-1", "applied")
			next(w, r)
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-2", "applied")
			next(w, r)
		}
	}

	// Test route builder with single middleware
	router.Route("/single").Use(middleware1).GET(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("single"))
	})

	// Test route builder with multiple middleware
	router.Route("/multiple").Use(middleware1, middleware2).POST(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("multiple"))
	})

	// Test route builder with chained middleware
	router.Route("/chained").Use(middleware1).Use(middleware2).PUT(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("chained"))
	})

	tests := []struct {
		method         string
		path           string
		expectMW1      bool
		expectMW2      bool
		expectBody     string
	}{
		{"GET", "/single", true, false, "single"},
		{"POST", "/multiple", true, true, "multiple"},
		{"PUT", "/chained", true, true, "chained"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d for %s %s, got %d", http.StatusOK, tt.method, tt.path, rr.Code)
		}

		if rr.Body.String() != tt.expectBody {
			t.Errorf("Expected body %q for %s %s, got %q", tt.expectBody, tt.method, tt.path, rr.Body.String())
		}

		if tt.expectMW1 && rr.Header().Get("X-Middleware-1") != "applied" {
			t.Errorf("Expected X-Middleware-1 header for %s %s", tt.method, tt.path)
		}

		if tt.expectMW2 && rr.Header().Get("X-Middleware-2") != "applied" {
			t.Errorf("Expected X-Middleware-2 header for %s %s", tt.method, tt.path)
		}
	}
}

func TestMiddlewareExecutionOrder(t *testing.T) {
	router := New()

	var executionOrder []string

	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "middleware1")
			next(w, r)
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "middleware2")
			next(w, r)
		}
	}

	router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
		executionOrder = append(executionOrder, "handler")
		w.Write([]byte("test"))
	}, middleware1, middleware2)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	executionOrder = []string{}
	router.ServeHTTP(rr, req)

	expectedOrder := []string{"middleware1", "middleware2", "handler"}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected execution order length %d, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("Expected execution order %v, got %v", expectedOrder, executionOrder)
			break
		}
	}
}
