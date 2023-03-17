package hctx

import "fmt"

type HandlerFunc func(hctx Context)

type HandlerChain interface {
	Add(fn HandlerFunc)
	Get(index int) HandlerFunc
	Size() int
	Combine(h HandlerChain) HandlerChain
	List() []HandlerFunc
}

func New(handlers ...HandlerFunc) HandlerChain {
	h := make(handlerChain, 0, len(handlers))
	h = append(h, handlers...)
	return &h
}

type handlerChain []HandlerFunc

func (r *handlerChain) Add(fn HandlerFunc) {
	*r = append(*r, fn)
}

func (r *handlerChain) Get(index int) HandlerFunc {
	return (*r)[index]
}

func (r *handlerChain) Size() int {
	return len(*r)
}

func (r *handlerChain) List() []HandlerFunc {
	return *r
}

func (r *handlerChain) Combine(h HandlerChain) HandlerChain {
	finalSize := r.Size() + h.Size()
	mergedHandlerChain := make(handlerChain, 0, finalSize)
	mergedHandlerChain = append(mergedHandlerChain, *r...)
	mergedHandlerChain = append(mergedHandlerChain, h.List()...)
	for _, h := range *r {
		fmt.Printf("merge handlers: %#v\n", h)
	}
	return New(mergedHandlerChain...)
}
