package hook

type Hook struct {
	name string
	ch   chan *Data
}

type Data struct {
	files []string
	ch    chan error
}

var all map[string]*Hook
var names []string

func New(name string, worker func(int, *Hook), nWorkers int) *Hook {
	h := &Hook{name: name, ch: make(chan *Data, 100)}
	for i := 0; i < nWorkers; i++ {
		go worker(i, h)
	}
	return h
}

func Names() ([]string, error) {
	return names
}

func Hooks() map[string]*Hook {
	return all
}

func (h *Hook) Name() string {
	return h.name
}

func (h *Hook) Channel() chan *Data {
	return h.ch
}
