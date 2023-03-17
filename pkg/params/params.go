package params

import "sync"

type Params interface {
	Add(Param)
	Get(name string) (string, bool)
	Size() int
	List() map[string]string
}

type Param struct {
	Key   string
	Value string
}

func New(maxParams uint16) Params {
	return &params{
		ps: make(map[string]string, maxParams),
	}
}

type params struct {
	m  sync.RWMutex
	ps map[string]string
}

func (r *params) Add(p Param) {
	r.m.Lock()
	defer r.m.Unlock()
	r.ps[p.Key] = p.Value
}

func (r *params) Get(name string) (string, bool) {
	r.m.RLock()
	defer r.m.RUnlock()
	v, ok := r.ps[name]
	return v, ok
}

func (r *params) Size() int {
	return len(r.ps)
}

func (r *params) List() map[string]string {
	r.m.RLock()
	defer r.m.RUnlock()
	ps := make(map[string]string, len(r.ps))
	for k, v := range r.ps {
		ps[k] = v
	}
	return ps
}
