package hook

import (
	"log"
	"os"
)

func init() {
	// removing files is going to be pretty fast,
	// no need to have many idle threads.
	all["REMOVE"] = New("REMOVE", remove, 1)
	names = append(names, "REMOVE")
	order["REMOVE"] = 1
}

func remove(i int, h *Hook) {
	log.Println("remove:", i, "ready for action")
	for {
		select {
		case data := <-h.ch:
			var err error
			for _, file := range data.Files {
				if err = os.Remove(file); err != nil {
					log.Println("remove: error removing", file, err)
					break
				}
			}
			data.Ch <- err
		}
	}
}
