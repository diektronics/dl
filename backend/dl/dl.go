// Package dl includes types and functions to download files.
package dl

import (
	"log"
	"sort"

	"diektronics.com/carter/dl/backend/db"
	"diektronics.com/carter/dl/backend/hook"
	"diektronics.com/carter/dl/backend/notifier"
	"diektronics.com/carter/dl/cfg"
	dlpb "diektronics.com/carter/dl/protos/dl"
)

// Downloader exports five functions that are made available through an RPC interface
// to add downloads to a working queue, get information of what is happening, delete
// old downloads, and get a list of available commands to run after completing a download.
type Downloader struct {
	q          chan *link
	db         *db.Db
	n          *notifier.Client
	defaultDir string
}

type link struct {
	l           *dlpb.Link
	destination string
	ch          chan *dlpb.Link
}

// New returns a pointer to Downloader provided a configuration and the number of workers
// to use.
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

func sanitizeHooks(down *dlpb.Down) error {
	if len(down.Posthook) == 0 {
		return nil
	}
	hooks := down.Posthook
	for _, h := range hooks {
		if _, err := hook.Order(h); err != nil {
			return err
		}
	}
	sort.Sort(hook.ByOrder(hooks))
	down.Posthook = hooks
	return nil
}
