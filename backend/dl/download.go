package dl

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"diektronics.com/carter/dl/backend/hook"
	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/types"
)

func (d *Downloader) download(down *types.Download) {
	defer d.n.Notify(down)
	down.Status = types.Running
	if err := d.db.Update(down); err != nil {
		log.Println("download:", down.Name, "error updating:", err)
	}
	ch := make(chan *types.Link, len(down.Links))
	for _, l := range down.Links {
		d.q <- &link{l, down.Destination, ch}
	}

	for i := 0; i < len(down.Links); i++ {
		l := <-ch
		if l.Status != types.Success {
			down.Status = l.Status
			down.Errors = append(down.Errors, fmt.Sprintf("download: %v failed to download", l.URL))
		}
		if err := d.db.Update(down); err != nil {
			log.Println("download:", down.Name, "error updating:", err)
		}
	}
	if down.Status != types.Running {
		return
	}

	log.Println("download:", down.Name, "all downloads complete, about to run posthooks", down.Posthook)
	files := make([]string, len(down.Links))
	for i, l := range down.Links {
		files[i] = l.Filename
	}

	hooks := strings.Split(down.Posthook, ",")
	for _, hookName := range hooks {
		if err := d.applyHook(hookName, files, down); err != nil {
			break
		}
	}
	if down.Status == types.Running {
		down.Status = types.Success
	}
	if err := d.db.Update(down); err != nil {
		log.Println("download: error updating:", err)
	}
	log.Println("download:", down.Name, "all done, going away")
}

func (d *Downloader) applyHook(hookName string, files []string, down *types.Download) error {
	hookName = strings.TrimSpace(hookName)
	if len(hookName) == 0 {
		return nil
	}
	h, ok := hook.All()[hookName]
	if !ok {
		down.Status = types.Error
		str := fmt.Sprintf("download: %v %v does not exist", down.Name, hookName)
		down.Errors = append(down.Errors, str)
		return errors.New(str)
	}
	log.Println("download:", down.Name, "about to run posthook", hookName)
	ch := make(chan error)
	data := &hook.Data{Files: files, Ch: ch, Extra: down.Name}
	h.Channel() <- data
	err := <-data.Ch
	if err != nil {
		down.Status = types.Error
		str := fmt.Sprintf("download: %v failed %v", h.Name(), err)
		down.Errors = append(down.Errors, str)
		return errors.New(str)
	}
	log.Println("download:", down.Name, h.Name(), "ran successfully")
	return nil
}

func (d *Downloader) worker(i int, c *cfg.Configuration) {
	log.Println("download:", i, "ready for action")

	for l := range d.q {
		l.l.Status = types.Running
		if err := d.db.Update(l.l); err != nil {
			log.Println("download: error updating:", err)
		}

		if err := os.MkdirAll(l.destination, 0777); err != nil {
			log.Println("download:", i, "err:", err)
			log.Println("download:", i, "cannot create directory:", l.destination)
			l.l.Status = types.Error
			l.ch <- l.l
			continue
		}
		log.Printf("download: %d getting %q into %q\n", i, l.l.URL, l.destination)
		cmd := []string{c.PlowprobePath,
			"--printf=%f%t%s%n",
			l.l.URL}
		output, err := exec.Command((cmd[0]), cmd[1:]...).CombinedOutput()
		if err != nil {
			log.Println("download:", i, "err:", err)
			log.Println("download:", i, "output:", string(output))
			l.l.Status = types.Error
			l.ch <- l.l
			continue
		}
		parts := strings.Fields(strings.TrimSpace(string(output)))
		fileSize, _ := strconv.ParseInt(parts[len(parts)-1], 10, 64)
		fileName := filepath.Join(l.destination, strings.Join(parts[:len(parts)-1], " "))
		done := make(chan struct{})
		monitorDone := make(chan struct{})
		go d.sizeMonitor(fileName+".part", fileSize, l.l, done, monitorDone)

		cmd = []string{c.PlowdownPath,
			//"--engine=xfilesharing",
			"--output-directory=" + l.destination,
			"--printf=%F",
			"--temp-rename",
			l.l.URL}
		output, err = exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		// Sync with sizeMonitor() goroutine.
		close(done)
		<-monitorDone
		if err != nil {
			log.Println("download:", i, "err:", err)
			log.Println("download:", i, "output:", string(output))
			l.l.Status = types.Error
		} else {
			parts = strings.Split(strings.TrimSpace(string(output)), "\n")
			l.l.Filename = parts[len(parts)-1]
			log.Println("download:", i, l.l.URL, "download complete")
			l.l.Status = types.Success
			l.l.Percent = 100.0
		}

		l.ch <- l.l
	}
}

func (d *Downloader) sizeMonitor(fileName string, fileSize int64, l *types.Link, done, monitorDone chan struct{}) {
	for {
		select {
		default:
			if fi, err := os.Stat(fileName); err == nil {
				l.Percent = float64(fi.Size()*1000/fileSize) / 10.0
				if err := d.db.Update(l); err != nil {
					log.Println("download: error updating:", err)
				}
			}
			time.Sleep(time.Duration(5) * time.Second)
		case <-done:
			close(monitorDone)
			return
		}
	}
}
