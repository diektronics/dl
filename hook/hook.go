package hook

type Hook interface {
	Name() string
	Channel() chan *Data
}

type Data struct {
	files []string
	ch    chan error
}

var all map[string]*Hook
var names []string

func init() {
	u := &unrar{make(chan *Data)}
	all[u.Name()] = u
	names = append(names, u.Name())

	r := &remove{make(chan *Data)}
	all[r.Name()] = r
	names = append(names, r.Name())
}

func Names() ([]string, error) {
	return names
}

func Hooks() map[string]*Hook {
	return all
}
