package dl

import (
	"errors"
	"log"
	"path/filepath"

	"diektronics.com/carter/dl/backend/hook"
	"diektronics.com/carter/dl/types"
)

func (d *Downloader) Download(down *types.Download, _ *string) error {
	if len(down.Destination) == 0 {
		down.Destination = filepath.Join(d.defaultDir, down.Name)
	}
	if err := sanitizeHooks(down); err != nil {
		return err
	}
	if err := d.db.Add(down); err != nil {
		return err
	}
	go d.download(down)

	return nil
}

func (d *Downloader) GetAll(statuses []types.Status, reply *types.GetAllReply) error {
	if reply == nil {
		err := errors.New("GetAll: reply is a nil pointer")
		log.Println(err)
		return err
	}
	var err error
	reply.Downs, err = d.db.GetAll(statuses)
	if err != nil {
		return err
	}
	return nil
}

func (d *Downloader) Get(id int64, down *types.Download) error {
	if down == nil {
		err := errors.New("Get: down is a nil pointer")
		log.Println(err)
		return err
	}
	ret, err := d.db.Get(id)
	if err != nil {
		log.Println(err)
		return err
	}
	*down = *ret
	return nil
}

func (d *Downloader) Del(down *types.Download, _ *string) error {
	return d.db.Del(down)
}

func (d *Downloader) HookNames(_ string, reply *types.HookReply) error {
	if reply == nil {
		err := errors.New("HookNames: reply is a nil pointer")
		log.Println(err)
		return err
	}
	reply.Names = hook.Names()
	return nil
}
