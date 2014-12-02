package dl

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"diektronics.com/carter/dl/backend/db"
	"diektronics.com/carter/dl/backend/hook"
	"diektronics.com/carter/dl/backend/notifier"
	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/types"
)

type Downloader struct {
	q          chan *link
	db         *db.Db
	n          *notifier.Client
	defaultDir string
}

type link struct {
	l           *types.Link
	destination string
	ch          chan *types.Link
}

func New(c *cfg.Configuration, nWorkers int) *Downloader {
	d := &Downloader{
		q:          make(chan *link, 1000),
		db:         db.New(c),
		n:          notifier.New(c),
		defaultDir: c.DownloadDir,
	}
	for i := 0; i < nWorkers; i++ {
		go d.worker(i, c)
	}

	if err := d.recovery(); err != nil {
		log.Fatal("recovery:", err)
	}

	return d
}

func (d *Downloader) recovery() error {
	downs, err := d.db.QueueRunning()
	if err != nil {
		return err
	}

	for _, down := range downs {
		go d.download(down)
	}

	return nil
}

func (d *Downloader) Download(down *types.Download, _ *string) error {
	if len(down.Destination) == 0 {
		down.Destination = filepath.Join(d.defaultDir, down.Name)
	}
	if err := d.db.Add(down); err != nil {
		return err
	}
	go d.download(down)

	return nil
}

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
		data := &hook.Data{files, ch, down.Name}
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
		cmd := []string{c.PlowdownPath,
			"--engine=xfilesharing",
			"--output-directory=" + l.destination,
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

func (d *Downloader) GetAll(statuses []types.Status, reply *types.GetAllReply) error {
	var err error
	reply.Downs, err = d.db.GetAll(statuses)
	if err != nil {
		return err
	}
	return nil
}

func (d *Downloader) Get(id int64, down *types.Download) error {
	ret, err := d.db.Get(id)
	*down = *ret
	return err
}

func (d *Downloader) Del(down *types.Download, _ *string) error {
	return d.db.Del(down)
}

func (d *Downloader) HookNames(_ string, reply *types.HookReply) error {
	reply.Names = hook.Names()
	return nil
}
