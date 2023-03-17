package server

import (
	"net/http"
	"sync"

	"github.com/idproxy/httpserver/pkg/hctx"
	"github.com/idproxy/httpserver/pkg/params"
	"github.com/idproxy/httpserver/pkg/router"
	"github.com/idproxy/httpserver/pkg/routetree"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	default404Body = []byte("404 page not found")
	//default405Body = []byte("405 method not allowed")
)

type Server interface {
	Router() router.Router
	Run(address string) error
	PrintRoutes()
}

type Config struct {
}

func New() Server {
	routes := routetree.New()
	router := router.New(routes)
	s := &server{
		routes: routes,
		router: router,
	}
	s.pool.New = func() any {
		return s.allocateContext()
	}
	return s
}

type server struct {
	routes routetree.Routes
	router router.Router // used to add routes to the route tree

	// UseRawPath if enabled, the url.RawPath will be used to find parameters.
	UseRawPath bool

	// UnescapePathValues if true, the path value will be unescaped.
	// If UseRawPath is false (by default), the UnescapePathValues effectively is true,
	// as url.Path gonna be used, which is already unescaped.
	UnescapePathValues bool

	// UseH2C enable h2c support.
	useH2C bool

	pool sync.Pool
}

func (r *server) Use(middleware ...hctx.HandlerFunc) router.Router {
	r.Router().Use(middleware...)
	//r.rebuild404Handlers()
	//r.rebuild405Handlers()
	return r.Router()
}

func (r *server) Router() router.Router {
	return r.router
}

func (r *server) PrintRoutes() {
	r.routes.Print()
}

func (r *server) Handler() http.Handler {
	if !r.useH2C {
		return r
	}
	h2s := &http2.Server{}
	return h2c.NewHandler(r, h2s)
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (r *server) Run(address string) error {
	return http.ListenAndServe(address, r.Handler())
}

// ServeHTTP implements the http.Handler interface.
func (r *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// get a context
	ctx := r.pool.Get().(hctx.Context)
	// configure the context
	ctx.Init(&hctx.Config{
		// TODO handle max params
		Params:             params.New(16),
		Request:            req,
		Writer:             w,
		UseRawPath:         r.UseRawPath,
		UnescapePathValues: r.UnescapePathValues,
	})

	r.handleHTTPRequest(ctx)

	r.pool.Put(ctx)
}

func (r *server) allocateContext() hctx.Context {
	return hctx.NewContext()
}

func (r *server) handleHTTPRequest(hctx hctx.Context) {
	// find the route matching the path through the context
	r.routes.GetRouteContext(hctx)

	if hctx.GetHandlers() != nil && hctx.GetHandlers().Size() > 0 {
		hctx.Next()
		return
	}

	/*
		if httpMethod != http.MethodConnect && rPath != "/" {
			if valueCtx.tsr && r.RedirectTrailingSlash {
				redirectTrailingSlash(gctx)
				return
			}
			if r.RedirectFixedPath && redirectFixedPath(gctx, root, r.RedirectFixedPath) {
				return
			}
		}
	*/

	//hctx.SetHandlers(r.allNoRoute)
	serveError(hctx, http.StatusNotFound, default404Body)
}

func serveError(hctx hctx.Context, code int, defaultMessage []byte) {
	hctx.String(code, string(defaultMessage))
}
