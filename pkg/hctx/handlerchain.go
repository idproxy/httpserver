package hctx

type HandlerFunc func(hctx Context)

type HandlerChain interface {
	Add(fn HandlerFunc)
	Get(index int) HandlerFunc
	Size() int
	Merge(h HandlerChain) HandlerChain
	list() []HandlerFunc
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

func (r *handlerChain) list() []HandlerFunc {
	return *r
}

func (r *handlerChain) Merge(h HandlerChain) HandlerChain {
	finalSize := r.Size() + h.Size()
	mergedHandlerChain := make(handlerChain, finalSize)
	copy(mergedHandlerChain, *r)
	copy(mergedHandlerChain[r.Size():], h.list())
	return &mergedHandlerChain
}
