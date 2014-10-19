package hook

type unrar struct {
	ch chan *Data
}

func (u *unrar) Name(*Data) string {
	return "UNRAR"
}

func (u *unrar) Channel() chan *Data {
	return ch
}
