package dl

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/db"
	"diektronics.com/carter/dl/types"
)

type Downloader interface {
	Download(*types.Download) error
}

type downloader struct {
	q  chan *link
	db *db.Db
}

type link struct {
	l    *types.Link
	name string
	ch   chan *types.Link
}

func New(c *cfg.Configuration, nWorkers int) *downloader {
	d := &downloader{
		q:  make(chan *link, 1000),
		db: db.New(c),
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
			down.Error += fmt.Sprintln(l.Url, "failed to download")
		}
		if err := d.db.Update(down); err != nil {
			log.Println("download: error updating:", err)
		}
	}
	if down.Status == types.Running {
		down.Status = types.Success
		if err := d.db.Update(down); err != nil {
			log.Println("download: error updating:", err)
		}
	}

	// TODO(diek): run posthooks
}

func (d *downloader) worker(i int) {
	log.Println(i, "ready for action")

	for l := range d.q {
		// TODO(diek): make this into a downloader var, and get it from cfg.Configuration
		destination := "~/Downloads/" + l.name
		if err := os.MkdirAll(destination, 0777); err != nil {
			log.Println(i, "err:", err)
			log.Println(i, "cannot create directory:", destination)
			l.l.Status = types.Error
			l.ch <- l.l
			continue
		}
		log.Printf("%d getting %q into %q\n", i, l.l.Url, destination)
		cmd := []string{"/home/carter/bin/plowdown",
			"--engine=xfilesharing",
			"--output-directory=" + destination,
			"--printf=%F",
			"--temp-rename",
			l.l.Url}
		output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			log.Println(i, "err:", err)
			log.Println(i, "output:", string(output))
			l.l.Status = types.Error
		} else {
			log.Println(i, l.l.Url, "download complete")
			l.l.Status = types.Success
		}
		l.ch <- l.l
	}
}
