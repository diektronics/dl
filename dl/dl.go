package dl

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/db"
	"diektronics.com/carter/dl/hook"
	"diektronics.com/carter/dl/types"
)

type Downloader struct {
	q  chan *link
	db *db.Db
	h  []*hook.Hook
}

type link struct {
	l       *types.Link
	dirName string
	ch      chan *types.Link
}

func New(c *cfg.Configuration, nWorkers int) *Downloader {
	d := &Downloader{
		q:  make(chan *link, 1000),
		db: db.New(c),
	}
	for i := 0; i < nWorkers; i++ {
		go d.worker(i)
	}

	return d
}

func (d *Downloader) Download(down *types.Download) error {
	if err := d.db.Add(down); err != nil {
		return err
	}
	go d.download(down)

	return nil
}

func (d *Downloader) download(down *types.Download) {
	down.Status = types.Running
	if err := d.db.Update(down); err != nil {
		log.Println("download:", down.Name, "error updating:", err)
	}
	ch := make(chan *types.Link, len(down.Links))
	for _, l := range down.Links {
		d.q <- &link{l, down.Name, ch}
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
		hookName = strings.TrimSpace(hookName)
		if len(hookName) == 0 {
			continue
		}
		h, ok := hook.All()[hookName]
		if !ok {
			down.Status = types.Error
			down.Errors = append(down.Errors, fmt.Sprintf("download: %v %v does not exist", down.Name, hookName))
			break
		}
		log.Println("download:", down.Name, "about to run posthook", hookName)
		ch := make(chan error)
		data := &hook.Data{files, ch}
		h.Channel() <- data
		err := <-data.Ch
		if err != nil {
			down.Status = types.Error
			down.Errors = append(down.Errors, fmt.Sprintf("download: %v failed %v", h.Name(), err))
			break
		}
		log.Println("download:", down.Name, h.Name(), "ran successfully")
	}
	if down.Status == types.Running {
		down.Status = types.Success
	}
	if err := d.db.Update(down); err != nil {
		log.Println("download: error updating:", err)
	}
	log.Println("download:", down.Name, "all done, going away")
}

func (d *Downloader) worker(i int) {
	log.Println("download:", i, "ready for action")

	for l := range d.q {
		l.l.Status = types.Running
		if err := d.db.Update(l.l); err != nil {
			log.Println("download: error updating:", err)
		}

		// TODO(diek): make this into a Downloader var, and get it from cfg.Configuration
		destination := "/mnt/data/video/downs/" + l.dirName
		if err := os.MkdirAll(destination, 0777); err != nil {
			log.Println("download:", i, "err:", err)
			log.Println("download:", i, "cannot create directory:", destination)
			l.l.Status = types.Error
			l.ch <- l.l
			continue
		}
		log.Printf("download: %d getting %q into %q\n", i, l.l.URL, destination)
		cmd := []string{"/home/carter/bin/plowdown",
			"--engine=xfilesharing",
			"--output-directory=" + destination,
			"--printf=%F",
			"--temp-rename",
			l.l.URL}
		output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			log.Println("download:", i, "err:", err)
			log.Println("download:", i, "output:", string(output))
			l.l.Status = types.Error
		} else {
			parts := strings.Split(strings.TrimSpace(string(output)), "\n")
			l.l.Filename = parts[len(parts)-1]
			log.Println("download:", i, l.l.URL, "download complete")
			l.l.Status = types.Success
		}
		l.ch <- l.l
	}
}

func (d *Downloader) Db() *db.Db { return d.db }
