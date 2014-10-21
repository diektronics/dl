package hook

func init() {
	// removing files is going to be pretty fast,
	// no need to have many iddle threads.
	all["REMOVE"] = New("REMOVE", remove, 1)
	names = append(names, "REMOVE")
}

func remove(i int, h *Hook) {

}
