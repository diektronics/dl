package dl

import (
	"log"
	"path/filepath"

	"diektronics.com/carter/dl/backend/hook"
	dlpb "diektronics.com/carter/dl/protos/dl"
	"golang.org/x/net/context"
)

func (d *Downloader) Download(ctx context.Context, in *dlpb.DownloadRequest) (*dlpb.DownloadResponse, error) {
	down := in.Down
	if len(down.Destination) == 0 {
		down.Destination = filepath.Join(d.defaultDir, down.Name)
	}
	if err := sanitizeHooks(down); err != nil {
		return nil, err
	}
	id, err := d.db.Add(down)
	if err != nil {
		return nil, err
	}
	go d.download(down)

	return &dlpb.DownloadResponse{Id: id}, nil
}

func (d *Downloader) GetAll(ctx context.Context, in *dlpb.GetAllRequest) (*dlpb.GetAllResponse, error) {
	downs, err := d.db.GetAll(in.Statuses)
	if err != nil {
		return nil, err
	}
	return &dlpb.GetAllResponse{Downs: downs}, nil
}

func (d *Downloader) Get(ctx context.Context, in *dlpb.GetRequest) (*dlpb.GetResponse, error) {
	down, err := d.db.Get(in.Id)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &dlpb.GetResponse{Down: down}, nil
}

func (d *Downloader) Del(ctx context.Context, in *dlpb.DelRequest) (*dlpb.DelResponse, error) {
	return &dlpb.DelResponse{}, d.db.Del(in.Down)
}

func (d *Downloader) HookNames(ctx context.Context, in *dlpb.HookNamesRequest) (*dlpb.HookNamesResponse, error) {
	return &dlpb.HookNamesResponse{Names: hook.Names()}, nil
}
