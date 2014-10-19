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

type Downloader interface {
	Download(*types.Download) error
}

type downloader struct {
	q  chan *link
	db *db.Db
	h  []*hook.Hook
}

type link struct {
	l       *types.Link
	dirName string
	ch      chan *types.Link
}

func New(c *cfg.Configuration, nWorkers int) *downloader {
	d := &downloader{
		q:  make(chan *link, 1000),
		db: db.New(c),
		h:  hook.Hooks(),
	}
	for i := 0; i < nWorkers; i++ {
		go d.worker(i)
	}

	return d
}

func (d *downloader) Download(down *types.Download) error {
	if err := d.db.Add(down); err != nil {
		return err
	}
	go d.download(down)

	return nil
}

func (d *downloader) download(down *types.Download) {
	down.Status = types.Running
	if err := d.db.Update(down); err != nil {
		log.Println("download: error updating:", err)
	}
	ch := make(chan types.Link, len(down.Links))
	for _, l := range down.Links {
		d.q <- &link{l, down.Name, ch}
	}
	i := 0
	for i < len(down.Links) {
		l <- ch
		if l.Status != types.Success {
			down.Status = l.Status
			down.Error += fmt.Sprintln("download:", l.URL, "failed to download")
		}
		if err := d.db.Update(down); err != nil {
			log.Println("download: error updating:", err)
		}
	}
	if down.Status != types.Running {
		return
	}

	files := make([]string, len(down.Links))
	for i, l := range down.Links {
		files[i] = l.Filename
	}

	hooks := strings.Split(down.Posthook, ",")
	for _, hook := range hooks {
		hook = strings.TrimSpace(hook)
		if h, ok := d.h[hook]; !ok {
			down.Status = types.Error
			down.Error += fmt.Sprintln("download:", hook, "does not exist")
			break
		}
		ch := make(chan error)
		data := &hook.Data{files, ch}
		h.Channel() <- data
		err := <-data.ch
		if err != nil {
			down.Status = types.Error
			down.Error += fmt.Sprintln("download:", hook, "failed", err)
			break
		}

	}
	if down.Status == types.Running {
		down.Status = types.Success
	}
	if err := d.db.Update(down); err != nil {
		log.Println("download: error updating:", err)
	}
}

func (d *downloader) worker(i int) {
	log.Println(i, "ready for action")

	for l := range d.q {
		// TODO(diek): make this into a downloader var, and get it from cfg.Configuration
		destination := "~/Downloads/" + l.dirName
		if err := os.MkdirAll(destination, 0777); err != nil {
			log.Println(i, "err:", err)
			log.Println(i, "cannot create directory:", destination)
			l.l.Status = types.Error
			l.ch <- l.l
			continue
		}
		log.Printf("%d getting %q into %q\n", i, l.l.URL, destination)
		cmd := []string{"/home/carter/bin/plowdown",
			"--engine=xfilesharing",
			"--output-directory=" + destination,
			"--printf=%F",
			"--temp-rename",
			l.l.URL}
		output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			log.Println(i, "err:", err)
			log.Println(i, "output:", string(output))
			l.l.Status = types.Error
		} else {
			parts = strings.Split(strings.TrimSpace(string(output)), "\n")
			l.l.Filename = parts[len(parts)-1]
			log.Println(i, l.l.URL, "download complete")
			l.l.Status = types.Success
		}
		l.ch <- l.l
	}
}
