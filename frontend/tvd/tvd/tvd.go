package tvd

import (
	"fmt"
	"log"
	"net/rpc"
	"path/filepath"
	"time"

	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/frontend/tvd/db"
	"diektronics.com/carter/dl/frontend/tvd/feed"
	"diektronics.com/carter/dl/types"
)

type Datamanager struct {
	c       *cfg.Configuration
	backend string
}

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
		log.Println("--mark--", timestamp)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(t)
	}
}

func (dm *Datamanager) doer(timestamp time.Time) (time.Time, error) {
	shows, newTimestamp, err := feed.ScrapeShows(dm.c.Feed)
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
	toDown, err := dm.getLinks(myShows)
	if err != nil {
		return timestamp, fmt.Errorf("getLinks: %v", err)
	}
	client, err := rpc.DialHTTP("tcp", dm.backend)
	if err != nil {
		return timestamp, fmt.Errorf("dialing: %v", err)
	}
	defer client.Close()
	for _, d := range toDown {
		log.Println(d.Down)
		if err := client.Call("Downloader.Download", d.Down, nil); err != nil {
			return timestamp, fmt.Errorf("Download: %v", err)
		}
	}

	return newTimestamp, db.New(dm.c).UpdateMyShows(toDown)
}

func (dm *Datamanager) selectMyShows(shows []*types.Show) ([]*types.Show, error) {
	titles := []string{}
	showMap := make(map[string][]*types.Show)
	for _, s := range shows {
		titles = append(titles, s.Name)
		showMap[s.Name] = append(showMap[s.Name], s)
	}

	// Select among these shows the ones I really watch.
	eps, err := db.New(dm.c).GetMyShows(titles)
	if err != nil {
		return nil, err
	}
	myShows := []*types.Show{}
	for _, ep := range eps {
		for _, s := range showMap[ep.Title] {
			if s.Eps > ep.Episode {
				season, err := feed.Season(s.Eps)
				if err != nil {
					log.Println("selectMyShows:", err)
					continue
				}
				s.Down = &types.Download{
					Name:        fmt.Sprintf("%v - %v", s.Name, s.Eps),
					Destination: filepath.Join(ep.Location, s.Name, season),
					Posthook:    "RENAME",
				}
				myShows = append(myShows, s)
			}
		}
	}

	return myShows, nil
}

func (dm *Datamanager) getLinks(shows []*types.Show) ([]*types.Show, error) {
	toDown := []*types.Show{}
	for _, s := range shows {
		if link := feed.Link(dm.c.LinkRegexp, s); len(link) > 0 {
			s.Down.Links = append(s.Down.Links, &types.Link{URL: link})
			toDown = append(toDown, s)
		}
	}
	return toDown, nil
}
