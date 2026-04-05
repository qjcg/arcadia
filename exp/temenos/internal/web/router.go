package web

import (
	"net/http"
)

// Router is a pattern-based HTTP router
type Router struct {
	// Maps HTTP methods to pattern handlers
	handlers map[string][]patternHandler
}

// patternHandler pairs a pattern with its handler
type patternHandler struct {
	pattern string
	handler http.HandlerFunc
	exact   bool
}

// NewRouter creates a new pattern-based router
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string][]patternHandler),
	}
}

// Get registers a handler for GET requests
func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.register(http.MethodGet, pattern, handler, false)
}

// Post registers a handler for POST requests
func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.register(http.MethodPost, pattern, handler, false)
}

// Put registers a handler for PUT requests
func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.register(http.MethodPut, pattern, handler, false)
}

// Delete registers a handler for DELETE requests
func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.register(http.MethodDelete, pattern, handler, false)
}

// Exact registers a handler that matches the exact path
func (r *Router) Exact(pattern string, handler http.HandlerFunc) {
	r.register(http.MethodGet, pattern, handler, true)
}

func (r *Router) register(method, pattern string, handler http.HandlerFunc, exact bool) {
	r.handlers[method] = append(r.handlers[method], patternHandler{
		pattern: pattern,
		handler: handler,
		exact:   exact,
	})
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	handlers, ok := r.handlers[method]
	if !ok {
		http.NotFound(w, req)
		return
	}

	for _, h := range handlers {
		if h.exact {
			if path == h.pattern {
				h.handler(w, req)
				return
			}
		} else {
			if matchPattern(h.pattern, path) {
				h.handler(w, req)
				return
			}
		}
	}

	http.NotFound(w, req)
}

// matchPattern checks if path matches the pattern
// Supports patterns like /modules/{id} or /study/{id}
func matchPattern(pattern, path string) bool {
	// Simple prefix matching for now
	if len(path) < len(pattern) {
		return false
	}

	// Check if path starts with pattern prefix
	if path[:len(pattern)] != pattern {
		return false
	}

	// Exact match
	if len(path) == len(pattern) {
		return true
	}

	// Match if pattern ends with / or path has / after pattern
	if len(pattern) > 0 && pattern[len(pattern)-1] == '/' {
		return true
	}

	// For patterns with path parameters like /modules/{id}
	// Accept if next char is /
	return path[len(pattern)] == '/'
}
