package types

import (
	"fmt"
	"strings"
	"time"
)

type Download struct {
	ID         int64
	Name       string
	Status     Status
	Error      string
	Posthook   string
	CreatedAt  time.Time
	ModifiedAt time.Time
	Links      []*Link
}

type Link struct {
	ID         int64
	Url        string
	Status     Status
	CreatedAt  time.Time
	ModifiedAt time.Time
}

type Status string

const (
	Queued  Status = "QUEUED"
	Running Status = "RUNNING"
	Success Status = "SUCCESS"
	Error   Status = "ERROR"
)

func (d *Download) String() string {
	links := []string{}
	for _, l := range d.Links {
		links = append(links, fmt.Sprintf("\tID: %v\n\tUrl: %v\n\tStatus: %v\n\tCreatedAt: %v\n\tModifiedAt: %v",
			l.ID, l.Url, l.Status, l.CreatedAt, l.ModifiedAt))
	}
	return fmt.Sprintf("ID: %v\nName: %v\nStatus: %v\nError: %v\nPosthook: %v\nCreatedAt: %v\nModifiedAt: %v\nLinks: %v\n",
		d.ID, d.Name, d.Status, d.Error, d.Posthook, d.CreatedAt, d.ModifiedAt, strings.Join(links, "\n"))
}
