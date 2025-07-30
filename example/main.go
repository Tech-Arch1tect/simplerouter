package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tech-arch1tect/simplerouter"
)

func main() {
	// Example 1: Quick setup with default access logging (combined format to stdout)
	routerWithDefaults := simplerouter.NewWithDefaults()

	// Example 2: Custom access logging - JSON format to stdout
	jsonRouter := simplerouter.New().Use(simplerouter.AccessLogging(simplerouter.AccessLogConfig{
		Output: os.Stdout,
		Format: simplerouter.JSONLogFormat,
	}))

	// Example 3: Router with compression middleware
	compressedRouter := simplerouter.New().Use(
		simplerouter.Compression(), // Gzip compression
		simplerouter.AccessLogging(simplerouter.AccessLogConfig{
			Output: os.Stdout,
			Format: simplerouter.CombinedLogFormat,
		}),
	)

	// Example 4: Router without middleware (for comparison)
	router := simplerouter.New()

	logMiddleware := func(next simplerouter.HandlerFunc) simplerouter.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("LOG: %s %s\n", r.Method, r.URL.Path)
			next(w, r)
		}
	}

	authMiddleware := func(next simplerouter.HandlerFunc) simplerouter.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token != "Bearer valid-token" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			fmt.Printf("AUTH: User authenticated for %s\n", r.URL.Path)
			next(w, r)
		}
	}

	adminMiddleware := func(next simplerouter.HandlerFunc) simplerouter.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			role := r.Header.Get("X-User-Role")
			if role != "admin" {
				http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
				return
			}
			fmt.Printf("ADMIN: Admin access granted for %s\n", r.URL.Path)
			next(w, r)
		}
	}

	// Routes with different logging setups

	// Default router (no access logging)
	router.GET("/no-logging", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "No access logging")
	})

	// Router with default access logging (combined format)
	routerWithDefaults.GET("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello with default access logging!")
	})

	// Router with JSON access logging
	jsonRouter.GET("/json-logged", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "JSON access logging")
	})

	// Compression example - large content to demonstrate compression
	largeContent := strings.Repeat("This is test content that will be compressed when the client supports gzip. ", 50)

	// Compressed content endpoint
	compressedRouter.GET("/compressed", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, largeContent)
	})

	// Combining access logging with custom middleware
	routerWithDefaults.GET("/public", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Public endpoint with access logging + custom middleware")
	}, logMiddleware)

	// Access logging + authentication
	routerWithDefaults.With(authMiddleware).GET("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Get all users (with access logging)")
	})

	// Route builder with access logging
	routerWithDefaults.Route("/api/admin").
		Use(authMiddleware, adminMiddleware).
		GET(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Admin API endpoint (with access logging)")
		})

	jsonRouter.Route("/api/profile").
		Use(authMiddleware).
		PUT(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Update profile (JSON access logging)")
		})

	// Router-wide middleware with access logging
	protectedRouter := routerWithDefaults.Use(authMiddleware)
	protectedRouter.GET("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "User dashboard (with access logging)")
	})

	// Route groups with access logging
	apiGroup := routerWithDefaults.Group("/api/v1")

	apiGroup.GET("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "API is healthy (with access logging)")
	})

	// Groups with additional middleware on top of access logging
	apiGroup.With(logMiddleware).GET("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "API status: running (access + custom logging)")
	})

	// Nested groups with access logging + auth
	usersGroup := apiGroup.With(authMiddleware).Group("/users")
	usersGroup.GET("/profile", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "User profile (with access logging)")
	})

	// Start servers to demonstrate different features
	fmt.Println("Starting servers:")
	fmt.Println("  :8083 - Default access logging (combined format)")
	fmt.Println("  :8084 - JSON access logging")
	fmt.Println("  :8085 - No middleware")
	fmt.Println("  :8086 - Compression + access logging")
	fmt.Println("")
	fmt.Println("Try these endpoints:")
	fmt.Println("  GET http://localhost:8083/ - Default access logging")
	fmt.Println("  GET http://localhost:8084/json-logged - JSON access logging")
	fmt.Println("  GET http://localhost:8085/no-logging - No middleware")
	fmt.Println("  GET http://localhost:8086/compressed - Gzip compressed content")
	fmt.Println("  GET http://localhost:8083/users -H 'Authorization: Bearer valid-token' - Auth + access logging")
	fmt.Println("")
	fmt.Println("To test compression:")
	fmt.Println("  curl -H 'Accept-Encoding: gzip' http://localhost:8086/compressed")
	fmt.Println("")

	// Start JSON logging server
	go func() {
		log.Fatal(jsonRouter.ListenAndServe(":8084"))
	}()

	// Start no-middleware server
	go func() {
		log.Fatal(router.ListenAndServe(":8085"))
	}()

	// Start compression server
	go func() {
		log.Fatal(compressedRouter.ListenAndServe(":8086"))
	}()

	// Start main server with default access logging
	log.Fatal(routerWithDefaults.ListenAndServe(":8083"))
}
