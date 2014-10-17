package dl

import (
	"log"

	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/db"
	"diektronics.com/carter/dl/types"
)

type Downloader interface {
	Download(*types.Download) error
}

type downloader struct {
	q  chan *types.Download
	db *db.Db
}

func New(c *cfg.Configuration, nWorkers int) *downloader {
	d := &downloader{
		q:  make(chan *types.Download, 1000),
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
	// TODO(diek): check len(d.q) != cap(d.q) and return error if so.
	d.q <- down

	return nil
}

func (d *downloader) worker(i int) {
	log.Println(i, "ready for action")
}
