package hook

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	// no need for more than 1 goroutine: or either it is super fast moving things around,
	// or it is going to take so much I/O that it makes no sense to have another thread
	// doing the same.
	all["RENAME"] = New("RENAME", rename, 1)
	names = append(names, "RENAME")
	order["RENAME"] = 2
}

// rename moves data.Files[0] into data.extra, keeping the same directory and extension.
func rename(i int, h *Hook) {
	log.Println("rename:", i, "ready for action")
	for {
		select {
		case data := <-h.ch:
			var err error
			if len(data.Files) != 1 {
				err = fmt.Errorf("can only rename one file, not %d", len(data.Files))
			} else {
				file := data.Files[0]
				fn := filepath.Base(file)
				extension := ""
				if idx := strings.LastIndex(fn, "."); idx != -1 {
					extension = fn[idx:len(fn)]
				}
				newName := filepath.Join(filepath.Dir(file), data.Extra.(string)+extension)
				if err = os.Rename(file, newName); err != nil {
					log.Println("rename: error renaming", file, err)
				}
			}
			data.Ch <- err
		}
	}
}
