package routetree

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/idproxy/httpserver/internal/pathsegment"
	"github.com/idproxy/httpserver/pkg/hctx"
)

const (
	wildcard = "wildcard"
)

// Routes define the
type Routes interface {
	// AddRoute adds a route to the routeTree per httpmethod by splitting the path in pathSegments
	// and adding the segment per segment in the routeTree
	AddRoute(httpMethod, absolutePath string, handlers hctx.HandlerChain) error
	// GetRouteContext updated the http context by recursively walking
	// the routeTree per PathSegment
	GetRouteContext(hctx hctx.Context)

	// helper functions
	Print()
	GetSupportedmethods() []string
}

func New() Routes {
	r := &routes{
		// the supported Metods are statically defined to avoid a global var
		supportedMethods: []string{
			http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
			http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
			http.MethodTrace,
		},
		// contain the routes in the http server router per http method
		routes: map[string]*node{},
	}
	// initialize the routes per httpMethod with a root pathSegment
	// since the handlerChain is empty this means the route is not actually active
	for _, method := range r.supportedMethods {
		r.routes[method] = &node{
			PathSegment: pathsegment.PathSegment{Value: "/", Kind: pathsegment.Root},
			children:    map[string]*node{},
		}
	}
	return r
}

// routes contains a list of routes the httpserver operates on
// structured by httpMethod with a tree per pathSegment.
// pathSegment can be of type wildcard if they are determined at runtime
type routes struct {
	m                sync.RWMutex
	routes           map[string]*node
	supportedMethods []string
}

// Print shows the routeTree recursively
func (r *routes) Print() {
	r.m.RLock()
	defer r.m.RUnlock()
	for method, n := range r.routes {
		if method == "GET" {
			fmt.Println(method)
			//fmt.Printf("pathSegment tree: %s, handlers: %d\n", n.pathSegment.value, len(n.handlers))
			n.Print(0)
		}
	}
}

func (r *routes) GetSupportedmethods() []string {
	return r.supportedMethods
}

// AddRoute adds a route to the routeTree per httpmethod by splitting the path in pathSegments
// and adding the segment per segment in the routeTree
func (r *routes) AddRoute(httpMethod, absolutePath string, handlers hctx.HandlerChain) error {
	r.m.Lock()
	defer r.m.Unlock()

	// VALIDATION LOGIC

	// a url path must start with /
	if absolutePath[0] != '/' {
		return errors.New("path must begin with '/'")
	}
	// httpmethod cannot be empty
	if httpMethod == "" {
		return errors.New("HTTP method can not be empty")
	}
	// at least one handler should be present
	if handlers.Size() == 0 {
		return errors.New("there must be at least one handler")
	}
	rn, ok := r.routes[httpMethod]
	if !ok {
		return fmt.Errorf("method: %s not matching supported methods: %v", httpMethod, r.supportedMethods)
	}

	// split the urlPath in path Segments and check for validity (validate the wildcard, etc)

	pathSegments, valid := pathsegment.New(absolutePath)
	if !valid {
		return fmt.Errorf("multiple wildcards found in pathSegment: %s", absolutePath)
	}
	fmt.Printf("routes pathSegments: %v\n", pathSegments)
	// this is a path with only a "/"
	if pathSegments.Size() == 1 {
		rn.handlers = handlers
		return nil
	}
	// continue add the handlers to the routeTree per pathSegment
	return rn.addroute(1, pathSegments, handlers)
}

// Below are the runtime methods

// GetRouteContext provides the route context of the http request based on searching the routes
// in the http server router
func (r *routes) GetRouteContext(hctx hctx.Context) {
	r.m.RLock()
	defer r.m.RUnlock()

	fmt.Printf("getValue -> path: %s, params: %v\n", hctx.GetURLPath(), hctx.GetParams())

	pathSegments, valid := pathsegment.New(hctx.GetURLPath())
	if !valid {
		// error while processing path
		hctx.SetStatus(http.StatusBadRequest)
		hctx.SetMessage(string(default400Body))
		return
	}
	fmt.Printf("getValue -> pathSegments: %v \n", pathSegments)
	n, ok := r.routes[hctx.GetMethod()]
	if !ok {
		// httpMethod not found
		hctx.SetStatus(http.StatusNotFound)
		hctx.SetMessage(string(default404Body))
		return
	}
	if pathSegments.Size() == 1 {
		// return the handlers attached to the route
		hctx.SetHandlers(n.handlers)
		// if no handlers attached the route is not found
		if n.handlers == nil || n.handlers.Size() == 0 {
			hctx.SetStatus(http.StatusNotFound)
			hctx.SetMessage(string(default404Body))
		}
		return
	}

	hctx.SetPathSegments(pathSegments)
	hctx.SetPathSegmentIndex(1)
	n.GetRouteContext(hctx)
}
