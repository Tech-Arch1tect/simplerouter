package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tech-arch1tect/simplerouter"
)

func main() {
	router := simplerouter.New()

	// Simple middleware for logging
	logMiddleware := func(next simplerouter.HandlerFunc) simplerouter.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("%s %s\n", r.Method, r.URL.Path)
			next(w, r)
		}
	}

	// Apply middleware to all routes
	router = router.Use(logMiddleware)

	// Basic routes
	router.GET("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	router.GET("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Get all users")
	})

	router.POST("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Create user")
	})

	router.PUT("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Update user")
	})

	router.DELETE("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Delete user")
	})

	// Route groups with prefix
	apiGroup := router.Group("/api/v1")

	apiGroup.GET("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "API is healthy")
	})

	apiGroup.GET("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "API status: running")
	})

	// Nested groups
	usersGroup := apiGroup.Group("/users")

	usersGroup.GET("/profile", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "User profile")
	})

	usersGroup.POST("/profile", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Update user profile")
	})

	fmt.Println("Server starting on :8083")
	log.Fatal(http.ListenAndServe(":8083", router))
}
