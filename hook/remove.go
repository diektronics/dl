package hook

type remove struct {
	ch chan *Data
}

func (r *remove) Name(*Data) string {
	return "REMOVE"
}

func (u *unrar) Channel() chan *Data {
	return ch
}
