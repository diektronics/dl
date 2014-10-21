package hook

import (
	"log"
	"os"
)

func init() {
	// removing files is going to be pretty fast,
	// no need to have many iddle threads.
	all["REMOVE"] = New("REMOVE", remove, 1)
	names = append(names, "REMOVE")
}

func remove(i int, h *Hook) {
	log.Println(i, "ready for action")
	for {
		select {
		case data := <-h.ch:
			var err error
			for _, file := range data.files {
				if err = os.Remove(file); err != nil {
					log.Println("unrar: error removing", file, err)
					break
				}
			}
			data.ch <- err
		}
	}
}
