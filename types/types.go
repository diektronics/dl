package types

import (
	"fmt"
	"strings"
	"time"
)

type Download struct {
	ID          int64
	Name        string
	Status      Status
	Errors      []string
	Posthook    string
	Destination string
	CreatedAt   time.Time
	ModifiedAt  time.Time
	Links       []*Link
}

type Link struct {
	ID         int64
	URL        string
	Status     Status
	CreatedAt  time.Time
	ModifiedAt time.Time
	Filename   string
}

type Status string

const (
	Queued  Status = "QUEUED"
	Running Status = "RUNNING"
	Success Status = "SUCCESS"
	Error   Status = "ERROR"
)

func AllStatuses() []Status {
	return []Status{Queued, Running, Success, Error}
}

func (d *Download) String() string {
	links := []string{}
	for _, l := range d.Links {
		links = append(links, fmt.Sprintf("\tID: %v\n\tURL: %v\n\tStatus: %v\n\tCreatedAt: %v\n\tModifiedAt: %v",
			l.ID, l.URL, l.Status, l.CreatedAt, l.ModifiedAt))
	}
	return fmt.Sprintf("ID: %v\nName: %v\nStatus: %v\nErrors: %v\nPosthook: %v\nDestination: %v\nCreatedAt: %v\nModifiedAt: %v\nLinks: %v\n",
		d.ID, d.Name, d.Status, strings.Join(d.Errors, "\n\t"), d.Posthook, d.Destination, d.CreatedAt, d.ModifiedAt, strings.Join(links, "\n"))
}

// RPC types

type HookReply struct {
	Names []string
}

type GetAllReply struct {
	Downs []*Download
}

// tvd types
// TODO(diek): I don't think this should go here...
type Show struct {
	Name string
	Eps  string
	Blob string
	Down *Download
}

type ByAlpha []*Show

func (s ByAlpha) Len() int      { return len(s) }
func (s ByAlpha) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByAlpha) Less(i, j int) bool {
	return s[i].Name < s[j].Name || s[i].Name == s[j].Name && s[i].Eps < s[j].Eps
}
