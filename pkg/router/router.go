package router

// This package contains logic to add routes to the route tree before the
// http server starts up.
// It gets initialized with the route Tree which will be used during runtime
// operation to lookup the respective handlers

import (
	"net/http"

	"github.com/idproxy/httpserver/internal/utils"
	"github.com/idproxy/httpserver/pkg/hctx"
	"github.com/idproxy/httpserver/pkg/routetree"
)

const (
	maxHandlers = 127
)

type Router interface {
	Route
	// Group creates a new router using a new relative path from the base router
	Group(string, ...hctx.HandlerFunc) Router
	GetHandlers() hctx.HandlerChain
}

type Route interface {
	Use(middleware ...hctx.HandlerFunc) Router

	Any(string, ...hctx.HandlerFunc) Router
	GET(string, ...hctx.HandlerFunc) Router
	POST(string, ...hctx.HandlerFunc) Router
	DELETE(string, ...hctx.HandlerFunc) Router
	PATCH(string, ...hctx.HandlerFunc) Router
	PUT(string, ...hctx.HandlerFunc) Router
	OPTIONS(string, ...hctx.HandlerFunc) Router
	HEAD(string, ...hctx.HandlerFunc) Router

	// internal
	getAbsolutePath(relativePath string) string
	addRoute(httpMethod, absolutePath string, handlers hctx.HandlerChain)
	getSupportedmethods() []string
}

func New(routes routetree.Routes) Router {
	return &router{
		basePath: "/",
		handlers: hctx.New(),
		parent:   nil,
		children: map[string]Router{},
		routes:   routes,
	}
}

type router struct {
	handlers hctx.HandlerChain
	basePath string
	parent   Router
	children map[string]Router
	routes   routetree.Routes
}

func (r *router) GetHandlers()  hctx.HandlerChain {
	return r.handlers
}

func (r *router) Group(relativePath string, handlers ...hctx.HandlerFunc) Router {
	return &router{
		handlers: hctx.New(handlers...),
		parent:   r,
		children: map[string]Router{},
		basePath: relativePath,
	}
}

func (r *router) Use(middleware ...hctx.HandlerFunc) Router {
	r.handlers = r.combineHandlers(hctx.New(middleware...))
	return r
}

// GET is a shortcut for router.Handle("GET", path, handlers).
func (r *router) GET(relativePath string, handlers ...hctx.HandlerFunc) Router {
	return r.add(http.MethodGet, relativePath, hctx.New(handlers...))
}

// POST is a shortcut for router.Handle("POST", path, handlers).
func (r *router) POST(relativePath string, handlers ...hctx.HandlerFunc) Router {
	return r.add(http.MethodPost, relativePath, hctx.New(handlers...))
}

// DELETE is a shortcut for router.Handle("DELETE", path, handlers).
func (r *router) DELETE(relativePath string, handlers ...hctx.HandlerFunc) Router {
	return r.add(http.MethodDelete, relativePath, hctx.New(handlers...))
}

// PATCH is a shortcut for router.Handle("PATCH", path, handlers).
func (r *router) PATCH(relativePath string, handlers ...hctx.HandlerFunc) Router {
	return r.add(http.MethodPatch, relativePath, hctx.New(handlers...))
}

// PUT is a shortcut for router.Handle("PUT", path, handlers).
func (r *router) PUT(relativePath string, handlers ...hctx.HandlerFunc) Router {
	return r.add(http.MethodPut, relativePath, hctx.New(handlers...))
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handlers).
func (r *router) OPTIONS(relativePath string, handlers ...hctx.HandlerFunc) Router {
	return r.add(http.MethodOptions, relativePath, hctx.New(handlers...))
}

// HEAD is a shortcut for router.Handle("HEAD", path, handlers).
func (r *router) HEAD(relativePath string, handlers ...hctx.HandlerFunc) Router {
	return r.add(http.MethodHead, relativePath, hctx.New(handlers...))
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (r *router) Any(relativePath string, handlers ...hctx.HandlerFunc) Router {
	for _, method := range r.getSupportedmethods() {
		r.add(method, relativePath, hctx.New(handlers...))
	}
	return r
}

func (r *router) add(httpMethod, relativePath string, handlers hctx.HandlerChain) Router {
	absolutePath := r.getAbsolutePath(relativePath)
	handlers = r.combineHandlers(handlers)

	r.addRoute(httpMethod, absolutePath, handlers)
	return r
}

// addRoute find the root of the routers and add the route in the route tree of the root routers
func (r *router) addRoute(httpMethod, absolutePath string, handlers hctx.HandlerChain) {
	if r.parent != nil {
		// continue walk the tree until we are at the root
		r.parent.addRoute(httpMethod, absolutePath, handlers)
		return
	}
	// insert the route in the tree
	// when an error occurs we panic since this is a wrong configuration
	if err := r.routes.AddRoute(httpMethod, absolutePath, handlers); err != nil {
		panic(err)
	}
}

func (r *router) combineHandlers(handlers hctx.HandlerChain) hctx.HandlerChain {
	if (r.handlers.Size() + handlers.Size()) > int(maxHandlers) {
		panic("too many handlers")
	}
	return r.handlers.Combine(handlers)
}

func (r *router) getAbsolutePath(relativePath string) string {
	newRelativePath := utils.JoinPaths(r.basePath, relativePath)
	if r.parent != nil {
		return r.parent.getAbsolutePath(newRelativePath)
	}
	return newRelativePath
}

func (r *router) getSupportedmethods() []string {
	if r.parent != nil {
		return r.parent.getSupportedmethods()
	}
	return r.routes.GetSupportedmethods()
}
