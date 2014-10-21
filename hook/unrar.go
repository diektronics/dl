package hook

func init() {
	// only one worker, unrar takes many resources, better to serialize it
	all["UNRAR"] = New("UNRAR", unrar, 1)
	names = append(names, "UNRAR")
}

func unrar(i int, h *Hook) {

}
