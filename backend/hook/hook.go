package hook

import (
	"fmt"
)

type Hook struct {
	name string
	ch   chan *Data
}

type Data struct {
	Files []string
	Extra interface{}
	Ch    chan error
}

var all map[string]*Hook
var names []string
var order map[string]int

func init() {
	all = make(map[string]*Hook)
	order = make(map[string]int)
}

func New(name string, worker func(int, *Hook), nWorkers int) *Hook {
	h := &Hook{name: name, ch: make(chan *Data, 100)}
	for i := 0; i < nWorkers; i++ {
		go worker(i, h)
	}
	return h
}

func Names() []string { return names }

func All() map[string]*Hook { return all }

func Order(h string) (int, error) {
	if i, ok := order[h]; ok {
		return i, nil
	} else {
		return i, fmt.Errorf("%v is not a valid hook", h)
	}
}

func (h *Hook) Name() string { return h.name }

func (h *Hook) Channel() chan *Data { return h.ch }

type ByOrder []string

func (o ByOrder) Len() int           { return len(o) }
func (o ByOrder) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o ByOrder) Less(i, j int) bool { return order[o[i]] < order[o[j]] }
