package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tech-arch1tect/simplerouter"
)

func main() {
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

	// No middleware
	router.GET("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	// Middleware Method 1: Direct middleware parameters
	router.GET("/public", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Public endpoint with logging")
	}, logMiddleware)

	// Middleware Method 2: With() method for temporary middleware
	router.With(logMiddleware, authMiddleware).GET("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Get all users")
	})

	// Middleware Method 3: Route builder
	router.Route("/api/admin").
		Use(logMiddleware, authMiddleware, adminMiddleware).
		GET(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Admin API endpoint")
		})

	router.Route("/api/profile").
		Use(logMiddleware).
		Use(authMiddleware).
		PUT(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Update profile")
		})

	// Middleware Method 4: Use() for router-wide middleware
	protectedRouter := router.Use(logMiddleware, authMiddleware)
	protectedRouter.GET("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "User dashboard")
	})

	apiGroup := router.Group("/api/v1")

	apiGroup.GET("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "API is healthy")
	})

	// Groups with middleware
	apiGroup.With(logMiddleware).GET("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "API status: running")
	})

	// Nested groups with middleware
	usersGroup := apiGroup.With(logMiddleware, authMiddleware).Group("/users")
	usersGroup.GET("/profile", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "User profile")
	})

	log.Fatal(http.ListenAndServe(":8083", router))
}
