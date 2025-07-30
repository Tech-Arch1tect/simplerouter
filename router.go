package simplerouter

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
)

type Router struct {
	mux         *http.ServeMux
	prefix      string
	middlewares []Middleware
	routes      map[string]map[string]HandlerFunc
	routeInfo   *[]RouteInfo
}

type RouteInfo struct {
	Method string
	Path   string
	Prefix string
}

type HandlerFunc func(http.ResponseWriter, *http.Request)

type Middleware func(HandlerFunc) HandlerFunc

func New() *Router {
	routeInfo := make([]RouteInfo, 0)
	return &Router{
		mux:         http.NewServeMux(),
		prefix:      "",
		middlewares: make([]Middleware, 0),
		routes:      make(map[string]map[string]HandlerFunc),
		routeInfo:   &routeInfo,
	}
}

func NewWithDefaults() *Router {
	return New().Use(AccessLogging(AccessLogConfig{
		Output: os.Stdout,
		Format: CombinedLogFormat,
	}))
}

func (r *Router) Handler() http.Handler {
	return r.mux
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) joinPaths(base, path string) string {
	if base == "" {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		return path
	}
	if path == "" {
		return base
	}

	// Ensure base ends with / and path doesn't start with /
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	return base + path
}

func (r *Router) Group(prefix string) *Router {
	newPrefix := r.joinPaths(r.prefix, prefix)

	middlewares := make([]Middleware, len(r.middlewares))
	copy(middlewares, r.middlewares)

	return &Router{
		mux:         r.mux,
		prefix:      newPrefix,
		middlewares: middlewares,
		routes:      r.routes,
		routeInfo:   r.routeInfo,
	}
}

func (r *Router) Handle(method, path string, handler HandlerFunc) {
	fullPath := r.joinPaths(r.prefix, path)

	finalHandler := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		finalHandler = r.middlewares[i](finalHandler)
	}

	if r.routes[fullPath] == nil {
		r.routes[fullPath] = make(map[string]HandlerFunc)
		r.mux.HandleFunc(fullPath, r.dispatch(fullPath))
	}

	r.routes[fullPath][method] = finalHandler

	*r.routeInfo = append(*r.routeInfo, RouteInfo{
		Method: method,
		Path:   fullPath,
		Prefix: r.prefix,
	})
}

func (r *Router) dispatch(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		methodHandlers := r.routes[path]
		if handler, exists := methodHandlers[req.Method]; exists {
			handler(w, req)
		} else {
			allowedMethods := make([]string, 0, len(methodHandlers))
			for method := range methodHandlers {
				allowedMethods = append(allowedMethods, method)
			}
			w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (r *Router) HandleFunc(method, path string, handler http.HandlerFunc) {
	r.Handle(method, path, HandlerFunc(handler))
}

func (r *Router) Use(middlewares ...Middleware) *Router {
	newMiddlewares := make([]Middleware, len(r.middlewares)+len(middlewares))
	copy(newMiddlewares, r.middlewares)
	copy(newMiddlewares[len(r.middlewares):], middlewares)

	return &Router{
		mux:         r.mux,
		prefix:      r.prefix,
		middlewares: newMiddlewares,
		routes:      r.routes,
		routeInfo:   r.routeInfo,
	}
}

func (r *Router) With(middlewares ...Middleware) *Router {
	return r.Use(middlewares...)
}

func (r *Router) ListenAndServe(addr string) error {
	r.PrintRoutes()
	return http.ListenAndServe(addr, r)
}

func (r *Router) ListenAndServeTLS(addr, certFile, keyFile string) error {
	r.PrintRoutes()
	return http.ListenAndServeTLS(addr, certFile, keyFile, r)
}

func (r *Router) PrintRoutes() {
	if len(*r.routeInfo) == 0 {
		fmt.Println("No routes registered")
		return
	}

	sortedRoutes := make([]RouteInfo, len(*r.routeInfo))
	copy(sortedRoutes, *r.routeInfo)
	sort.Slice(sortedRoutes, func(i, j int) bool {
		if sortedRoutes[i].Path == sortedRoutes[j].Path {
			return sortedRoutes[i].Method < sortedRoutes[j].Method
		}
		return sortedRoutes[i].Path < sortedRoutes[j].Path
	})

	fmt.Println("\n📋 Registered Routes:")
	fmt.Println("┌─────────┬─────────────────────────────────────────────┐")
	fmt.Println("│ Method  │ Path                                    │")
	fmt.Println("├─────────┼─────────────────────────────────────────────┤")

	for _, route := range sortedRoutes {
		fmt.Printf("│ %-7s │ %-43s │\n", route.Method, route.Path)
	}

	fmt.Println("└─────────┴─────────────────────────────────────────────┘")
	fmt.Printf("Total routes: %d\n\n", len(sortedRoutes))
}

func (r *Router) GET(path string, handler HandlerFunc, middlewares ...Middleware) {
	if len(middlewares) > 0 {
		r.With(middlewares...).Handle("GET", path, handler)
	} else {
		r.Handle("GET", path, handler)
	}
}

func (r *Router) POST(path string, handler HandlerFunc, middlewares ...Middleware) {
	if len(middlewares) > 0 {
		r.With(middlewares...).Handle("POST", path, handler)
	} else {
		r.Handle("POST", path, handler)
	}
}

func (r *Router) PUT(path string, handler HandlerFunc, middlewares ...Middleware) {
	if len(middlewares) > 0 {
		r.With(middlewares...).Handle("PUT", path, handler)
	} else {
		r.Handle("PUT", path, handler)
	}
}

func (r *Router) DELETE(path string, handler HandlerFunc, middlewares ...Middleware) {
	if len(middlewares) > 0 {
		r.With(middlewares...).Handle("DELETE", path, handler)
	} else {
		r.Handle("DELETE", path, handler)
	}
}

func (r *Router) PATCH(path string, handler HandlerFunc, middlewares ...Middleware) {
	if len(middlewares) > 0 {
		r.With(middlewares...).Handle("PATCH", path, handler)
	} else {
		r.Handle("PATCH", path, handler)
	}
}

func (r *Router) HEAD(path string, handler HandlerFunc, middlewares ...Middleware) {
	if len(middlewares) > 0 {
		r.With(middlewares...).Handle("HEAD", path, handler)
	} else {
		r.Handle("HEAD", path, handler)
	}
}

func (r *Router) OPTIONS(path string, handler HandlerFunc, middlewares ...Middleware) {
	if len(middlewares) > 0 {
		r.With(middlewares...).Handle("OPTIONS", path, handler)
	} else {
		r.Handle("OPTIONS", path, handler)
	}
}

type RouteBuilder struct {
	router      *Router
	path        string
	middlewares []Middleware
}

func (r *Router) Route(path string) *RouteBuilder {
	return &RouteBuilder{
		router:      r,
		path:        path,
		middlewares: make([]Middleware, 0),
	}
}

func (rb *RouteBuilder) Use(middlewares ...Middleware) *RouteBuilder {
	rb.middlewares = append(rb.middlewares, middlewares...)
	return rb
}

func (rb *RouteBuilder) GET(handler HandlerFunc) {
	rb.router.GET(rb.path, handler, rb.middlewares...)
}

func (rb *RouteBuilder) POST(handler HandlerFunc) {
	rb.router.POST(rb.path, handler, rb.middlewares...)
}

func (rb *RouteBuilder) PUT(handler HandlerFunc) {
	rb.router.PUT(rb.path, handler, rb.middlewares...)
}

func (rb *RouteBuilder) DELETE(handler HandlerFunc) {
	rb.router.DELETE(rb.path, handler, rb.middlewares...)
}

func (rb *RouteBuilder) PATCH(handler HandlerFunc) {
	rb.router.PATCH(rb.path, handler, rb.middlewares...)
}

func (rb *RouteBuilder) HEAD(handler HandlerFunc) {
	rb.router.HEAD(rb.path, handler, rb.middlewares...)
}

func (rb *RouteBuilder) OPTIONS(handler HandlerFunc) {
	rb.router.OPTIONS(rb.path, handler, rb.middlewares...)
}
