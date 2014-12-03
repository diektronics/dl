package tvd

import (
	"fmt"
	"log"
	"net/rpc"
	"time"

	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/types"
)

type Datamanager struct {
	c       *cfg.Configuration
	backend string
}

type Show struct{}

func New(c *cfg.Configuration) *Datamanager {
	return &Datamanager{
		c:       c,
		backend: fmt.Sprintf("localhost:%v", c.BackendPort),
	}
}

func (dm *Datamanager) Run(t time.Duration) {
	go dm.worker(t)
}

func (dm *Datamanager) worker(t time.Duration) {
	var timestamp time.Time
	for {
		var err error
		timestamp, err = dm.doer(timestamp)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(t)
	}
}

func (dm *Datamanager) doer(timestamp time.Time) (time.Time, error) {
	shows, newTimestamp, err := dm.scrapeShows()
	if err != nil {
		return timestamp, fmt.Errorf("scrapeShows: %v", err)
	}
	if !newTimestamp.After(timestamp) {
		return timestamp, nil
	}
	myShows, err := dm.selectMyShows(shows)
	if err != nil {
		return timestamp, fmt.Errorf("selectMyShows: %v", err)
	}
	downs, err := dm.getLinks(myShows)
	if err != nil {
		return timestamp, fmt.Errorf("getLinks: %v", err)
	}
	client, err := rpc.DialHTTP("tcp", dm.backend)
	if err != nil {
		return timestamp, fmt.Errorf("dialing: %v", err)
	}
	defer client.Close()
	for _, d := range downs {
		if err := client.Call("Downloader.Download", d, nil); err != nil {
			return timestamp, fmt.Errorf("Download: %v", err)
		}
		if err := dm.updateLastEpisode(d); err != nil {
			return timestamp, fmt.Errorf("updateLastEpisode: %v", err)
		}
	}

	return newTimestamp, nil
}

func (dm *Datamanager) scrapeShows() ([]*Show, time.Time, error)          { return nil, *new(time.Time), nil }
func (dm *Datamanager) selectMyShows(shows []*Show) ([]*Show, error)      { return nil, nil }
func (dm *Datamanager) getLinks(shows []*Show) ([]*types.Download, error) { return nil, nil }
func (dm *Datamanager) updateLastEpisode(down *types.Download) error      { return nil }
