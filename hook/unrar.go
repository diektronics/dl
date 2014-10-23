package hook

import (
	"log"
	"os/exec"
	"path/filepath"
	"sort"
)

func init() {
	// only one worker, unrar takes many resources, better to serialize it
	all["UNRAR"] = New("UNRAR", unrar, 1)
	names = append(names, "UNRAR")
}

func unrar(i int, h *Hook) {
	log.Println("unrar:", i, "ready for action")
	for {
		select {
		case data := <-h.ch:
			var err error
			// sort alphabeticallt data.files, and pass the first one to unrar x
			sort.Strings(data.Files)
			if len(data.Files) > 0 {
				cmd := []string{"/usr/bin/unrar",
					"x", data.Files[0], filepath.Dir(data.Files[0])}
				var output []byte
				output, err = exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
				if err != nil {
					log.Println("unrar:", i, "err:", err)
					log.Println("unrar:", i, "output:", string(output))
				}
			}
			data.Ch <- err
		}
	}
}
