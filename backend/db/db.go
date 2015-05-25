package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"diektronics.com/carter/dl/cfg"
	dlpb "diektronics.com/carter/dl/protos/dl"
	_ "github.com/Go-SQL-Driver/MySQL"
	"github.com/golang/protobuf/proto"
)

type Db struct {
	connectionString string
	q                chan proto.Message
}

func New(c *cfg.Configuration) *Db {
	d := &Db{
		connectionString: fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true&loc=Local",
			c.DbUser, c.DbPassword, c.DbServer, c.DbDatabase),
		q: make(chan proto.Message, 1000),
	}
	go d.worker(0)

	return d
}

func (d *Db) Add(down *dlpb.Down) (int64, error) {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	now := time.Now()
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	res, err := tx.Exec("INSERT INTO downloads (name, posthook, destination, created_at, modified_at) VALUES (?, ?, ?, ?, ?)",
		down.Name, strings.Join(down.Posthook, ","), down.Destination, now, now)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	down.Id = id
	down.Status = dlpb.Status_QUEUED
	down.CreatedAt = now.Unix()
	down.ModifiedAt = now.Unix()

	for _, link := range down.Links {
		res, err := tx.Exec("INSERT INTO links (download_id, url, created_at, modified_at) VALUES (?, ?, ?, ?)",
			id, link.Url, now, now)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		link_id, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		link.Id = link_id
		link.Status = dlpb.Status_QUEUED
		link.CreatedAt = now.Unix()
		link.ModifiedAt = now.Unix()
	}
	tx.Commit()

	return id, nil
}

func (d *Db) Get(id int64) (*dlpb.Down, error) {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	down := &dlpb.Down{Id: id}
	var status string
	var errStr string
	var hooks string
	var createdAt time.Time
	var modifiedAt time.Time
	if err := db.QueryRow("SELECT name, status, error, posthook, destination, created_at, modified_at FROM downloads WHERE id=?", id).Scan(
		&down.Name, &status, &errStr, &hooks, &down.Destination, &createdAt, &modifiedAt); err != nil {
		return nil, err
	}
	down.Status = dlpb.Status(dlpb.Status_value[status])
	for _, e := range strings.Split(errStr, "\n") {
		if len(e) > 0 {
			down.Errors = append(down.Errors, e)
		}
	}
	for _, h := range strings.Split(hooks, ",") {
		if len(h) > 0 {
			down.Posthook = append(down.Posthook, h)
		}
	}
	down.CreatedAt = createdAt.Unix()
	down.ModifiedAt = modifiedAt.Unix()

	rows, err := db.Query("SELECT id, url, status, percent, created_at, modified_at FROM links WHERE download_id=?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		l := &dlpb.Link{}
		if err := rows.Scan(&l.Id, &l.Url, &status, &l.Percent, &createdAt, &modifiedAt); err != nil {
			return nil, err
		}
		l.Status = dlpb.Status(dlpb.Status_value[status])
		l.CreatedAt = createdAt.Unix()
		l.ModifiedAt = modifiedAt.Unix()
		down.Links = append(down.Links, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return down, nil
}

func allStatuses() []dlpb.Status {
	out := make([]dlpb.Status, 0, len(dlpb.Status_name))
	for key := range dlpb.Status_name {
		out = append(out, dlpb.Status(key))
	}
	return out
}

func (d *Db) GetAll(statuses []dlpb.Status) ([]*dlpb.Down, error) {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	downs := []*dlpb.Down{}
	var status string
	if statuses == nil || len(statuses) == 0 {
		statuses = allStatuses()
	}
	query := "SELECT id, name, status, error, posthook, destination, created_at, modified_at FROM downloads WHERE status IN ("
	vals := []interface{}{}
	for _, s := range statuses {
		query += "?,"
		vals = append(vals, s.String())
	}
	query = query[0 : len(query)-1]
	query += ")"

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		down := &dlpb.Down{}
		var errStr string
		var hooks string
		var createdAt time.Time
		var modifiedAt time.Time
		if err := rows.Scan(&down.Id, &down.Name, &status, &errStr, &hooks, &down.Destination, &createdAt, &modifiedAt); err != nil {
			return nil, err
		}
		down.Status = dlpb.Status(dlpb.Status_value[status])
		for _, e := range strings.Split(errStr, "\n") {
			if len(e) > 0 {
				down.Errors = append(down.Errors, e)
			}
		}
		for _, h := range strings.Split(hooks, ",") {
			if len(h) > 0 {
				down.Posthook = append(down.Posthook, h)
			}
		}
		down.CreatedAt = createdAt.Unix()
		down.ModifiedAt = modifiedAt.Unix()

		rowsLinks, err := db.Query("SELECT id, url, status, percent, created_at, modified_at FROM links WHERE download_id=?", down.Id)
		if err != nil {
			return nil, err
		}
		defer rowsLinks.Close()
		for rowsLinks.Next() {
			l := &dlpb.Link{}
			if err := rowsLinks.Scan(&l.Id, &l.Url, &status, &l.Percent, &createdAt, &modifiedAt); err != nil {
				return nil, err
			}
			l.Status = dlpb.Status(dlpb.Status_value[status])
			l.CreatedAt = createdAt.Unix()
			l.ModifiedAt = modifiedAt.Unix()
			down.Links = append(down.Links, l)
		}
		if err := rowsLinks.Err(); err != nil {
			return nil, err
		}

		downs = append(downs, down)
	}
	return downs, nil
}

func (d *Db) Del(down *dlpb.Down) error {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec("DELETE FROM links WHERE download_id=?", down.Id)
	if err != nil {
		tx.Rollback()
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if n != int64(len(down.Links)) {
		tx.Rollback()
		return fmt.Errorf("Del: unexpected rows affected %v != %v", n, len(down.Links))
	}

	res, err = tx.Exec("DELETE FROM downloads WHERE id=?", down.Id)
	if err != nil {
		tx.Rollback()
		return err
	}
	n, err = res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if n != 1 {
		tx.Rollback()
		return fmt.Errorf("Del: unexpected rows affected %v != 1", n)
	}

	tx.Commit()
	return nil
}

func (d *Db) Update(data proto.Message) error {
	// check data is of the supported types
	switch data := data.(type) {
	default:
		return errors.New("Update: unexpected data")

	case *dlpb.Down, *dlpb.Link:
		d.q <- data
	}

	return nil
}

func (d *Db) QueueRunning() ([]*dlpb.Down, error) {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec("UPDATE downloads SET status='QUEUED' where status='RUNNING'")
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec("UPDATE links SET status='QUEUED' where status='RUNNING'")
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return d.GetAll([]dlpb.Status{dlpb.Status_QUEUED})
}

// FIXME(diek): even if there is no change, these updates will always change the data
// because modified_at is set to time.Now().
func (d *Db) worker(i int) {
	log.Println("db:", i, "ready for action")

	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		log.Println("db:", err)
		return
	}
	defer db.Close()

	for data := range d.q {
		tx, err := db.Begin()
		if err != nil {
			log.Println("db:", err)
			continue
		}
		switch data := data.(type) {
		case *dlpb.Down:
			if err := updateDownload(tx, data); err != nil {
				tx.Rollback()
				log.Println("db:", err)
				continue
			}
		case *dlpb.Link:
			if err := updateLink(tx, []*dlpb.Link{data}); err != nil {
				tx.Rollback()
				log.Println("db:", err)
				continue
			}
		}

		tx.Commit()
	}
}

func updateDownload(tx *sql.Tx, down *dlpb.Down) error {
	if err := updateLink(tx, down.Links); err != nil {
		return err
	}

	now := time.Now()
	res, err := tx.Exec("UPDATE downloads SET status=?, error=?, modified_at=? WHERE id=?",
		down.Status.String(), strings.Join(down.Errors, "\n"), now, down.Id)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 1 {
		down.ModifiedAt = now.Unix()

	}

	return nil
}

func updateLink(tx *sql.Tx, links []*dlpb.Link) error {
	now := time.Now()
	for _, l := range links {
		res, err := tx.Exec("UPDATE links SET status=?, percent=?, modified_at=? WHERE id=?",
			l.Status.String(), l.Percent, now, l.Id)
		if err != nil {
			return err
		}
		if n, err := res.RowsAffected(); err != nil {
			return err
		} else if n == 1 {
			l.ModifiedAt = now.Unix()
		}
	}

	return nil
}
