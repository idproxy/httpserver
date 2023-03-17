package routetree

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/idproxy/httpserver/internal/pathsegment"
	"github.com/idproxy/httpserver/pkg/hctx"
	"github.com/idproxy/httpserver/pkg/params"
)

type node struct {
	pathsegment.PathSegment
	m        sync.RWMutex
	children map[string]*node
	handlers hctx.HandlerChain
}

func (r *node) Print(i int) {
	r.m.RLock()
	defer r.m.RUnlock()
	if r.handlers != nil {
		fmt.Printf("%*s pathSegment: %s, kind: %s, handlers: %d\n", i, "", r.PathSegment.Value, r.PathSegment.Kind.String(), r.handlers.Size())
	} else {
		fmt.Printf("%*s pathSegment: %s, kind: %s, handlers: %d\n", i, "", r.PathSegment.Value, r.PathSegment.Kind.String(), 0)
	}
	for _, n := range r.children {
		n.Print(i + 1)
	}
}

func (r *node) addroute(idx int, pathSegments pathsegment.PathSegments, handlers hctx.HandlerChain) error {
	ps := pathSegments.Get(idx)
	psValue := ps.Value
	if ps.Kind == pathsegment.Param || ps.Kind == pathsegment.CatchAll {
		psValue = wildcard
	}
	n, ok := r.children[psValue]
	if !ok {
		n = &node{
			PathSegment: ps,
			children:    map[string]*node{},
		}
		r.children[psValue] = n
	}
	if ok && psValue == wildcard {
		if !ok {
			return fmt.Errorf("got wildcard: %s, but another wildcard was already in place: %s", ps.Value, n.PathSegment.Value)
		}
	}
	fmt.Printf("routes addRoute: pathSegments: %v, idx: %d ps length: %d\n", pathSegments, idx, pathSegments.Size()-1)
	if idx == pathSegments.Size()-1 {
		fmt.Printf("routes addRoute: addHandlers %d\n", handlers.Size())
		// for the last path segment add the handlers and return
		n.handlers = handlers
		return nil
	}
	return n.addroute(idx+1, pathSegments, handlers)
}

// dynamic processing per http request

func (r *node) GetRouteContext(hctx hctx.Context) {
	r.m.RLock()
	defer r.m.RUnlock()
	node, ok := r.children[hctx.GetPathSegments().Get(hctx.GetPathSegmentIndex()).Value]
	if !ok {
		fmt.Printf("getValue: children %v\n", r.children)
		fmt.Printf("getValue: not found -> %s\n", hctx.GetPathSegments().Get(hctx.GetPathSegmentIndex()).Value)
		// if the node was not found we need to validate if there is a wildcard
		node, ok = r.children[wildcard]
		if !ok {
			// no wildcard found in this pathSegment
			hctx.SetStatus(http.StatusNotFound)
			hctx.SetMessage(string(default404Body))
			return
		}
		// wildcard exists for the pathSegment
		// add the param KeyValue to the parameter list
		//fmt.Printf("params: %v\n", hctx.GetParams())
		hctx.GetParams().Add(params.Param{
			Key:   node.PathSegment.Value[1:],
			Value: hctx.GetPathSegments().Get(hctx.GetPathSegmentIndex()).Value,
		})
		//fmt.Printf("params: %v\n", hctx.GetParams().List())
	}

	fmt.Printf("getValue: pathSegments %v pathSegmentIdx: %d total ps length: %d\n",
		hctx.GetPathSegments(),
		hctx.GetPathSegmentIndex(),
		hctx.GetPathSegments().Size()-1)
	if hctx.GetPathSegmentIndex() == hctx.GetPathSegments().Size()-1 {
		/*
			if r.handlers != nil {
				fmt.Printf("getValue: handlers: %v\n", r.handlers.Size())
			}
		*/

		hctx.SetHandlers(node.handlers)
		if node.handlers == nil || node.handlers.Size() == 0 {
			hctx.SetStatus(http.StatusNotFound)
			hctx.SetMessage(string(default404Body))
		}
		return
	}
	hctx.IncrementPathSegmentIndex()
	node.GetRouteContext(hctx)
}
