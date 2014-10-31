package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/types"
	_ "github.com/Go-SQL-Driver/MySQL"
)

type Db struct {
	connectionString string
	q                chan interface{}
}

func New(c *cfg.Configuration) *Db {
	d := &Db{
		connectionString: fmt.Sprintf("%s:%s@%s/%s?charset=utf8&parseTime=true&loc=Local",
			c.DbUser, c.DbPassword, c.DbServer, c.DbDatabase),
		q: make(chan interface{}, 1000),
	}
	go d.worker(0)

	return d
}

func (d *Db) Add(down *types.Download) error {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	now := time.Now()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec("INSERT INTO downloads (name, posthook, created_at, modified_at) VALUES (?, ?, ?, ?)",
		down.Name, down.Posthook, now, now)
	if err != nil {
		tx.Rollback()
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	down.ID = id
	down.Status = types.Queued
	down.CreatedAt = now
	down.ModifiedAt = now

	for _, link := range down.Links {
		res, err := tx.Exec("INSERT INTO links (download_id, url, created_at, modified_at) VALUES (?, ?, ?, ?)",
			id, link.URL, now, now)
		if err != nil {
			tx.Rollback()
			return err
		}
		link_id, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return err
		}
		link.ID = link_id
		link.Status = types.Queued
		link.CreatedAt = now
		link.ModifiedAt = now
	}
	tx.Commit()

	return nil
}

func (d *Db) Get(id int64) (*types.Download, error) {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	down := &types.Download{ID: id}
	var status string
	var errStr string
	if err := db.QueryRow("SELECT name, status, error, posthook, created_at, modified_at FROM downloads WHERE id=?", id).Scan(
		&down.Name, &status, &errStr, &down.Posthook, &down.CreatedAt, &down.ModifiedAt); err != nil {
		return nil, err
	}
	down.Status = types.Status(status)
	for _, e := range strings.Split(errStr, "\n") {
		if len(e) > 0 {
			down.Errors = append(down.Errors, e)
		}
	}

	rows, err := db.Query("SELECT id, url, status, created_at, modified_at FROM links WHERE download_id=?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		l := &types.Link{}
		if err := rows.Scan(&l.ID, &l.URL, &status, &l.CreatedAt, &l.ModifiedAt); err != nil {
			return nil, err
		}
		l.Status = types.Status(status)
		down.Links = append(down.Links, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return down, nil
}

func (d *Db) GetAll(statuses []types.Status) ([]*types.Download, error) {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	downs := []*types.Download{}
	var status string
	if statuses == nil || len(statuses) == 0 {
		statuses = types.AllStatuses()
	}
	query := "SELECT id, name, status, error, posthook, created_at, modified_at FROM downloads WHERE status IN ("
	vals := []interface{}{}
	for _, s := range statuses {
		query += "?,"
		vals = append(vals, string(s))
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
		down := &types.Download{}
		var errStr string
		if err := rows.Scan(&down.ID, &down.Name, &status, &errStr, &down.Posthook, &down.CreatedAt, &down.ModifiedAt); err != nil {
			return nil, err
		}
		down.Status = types.Status(status)
		for _, e := range strings.Split(errStr, "\n") {
			if len(e) > 0 {
				down.Errors = append(down.Errors, e)
			}
		}

		rowsLinks, err := db.Query("SELECT id, url, status, created_at, modified_at FROM links WHERE download_id=?", down.ID)
		if err != nil {
			return nil, err
		}
		defer rowsLinks.Close()
		for rowsLinks.Next() {
			l := &types.Link{}
			if err := rowsLinks.Scan(&l.ID, &l.URL, &status, &l.CreatedAt, &l.ModifiedAt); err != nil {
				return nil, err
			}
			l.Status = types.Status(status)
			down.Links = append(down.Links, l)
		}
		if err := rowsLinks.Err(); err != nil {
			return nil, err
		}

		downs = append(downs, down)
	}
	return downs, nil
}

func (d *Db) Del(down *types.Download) error {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec("DELETE FROM links WHERE download_id=?", down.ID)
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

	res, err = tx.Exec("DELETE FROM downloads WHERE id=?", down.ID)
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

func (d *Db) Update(data interface{}) error {
	// check data is of the supported types
	switch data := data.(type) {
	default:
		return errors.New("Update: unexpected data")

	case *types.Download, *types.Link:
		d.q <- data
	}

	return nil
}

func (d *Db) QueueRunning() ([]*types.Download, error) {
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

	return d.GetAll([]types.Status{types.Queued})
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
		case *types.Download:
			if err := updateDownload(tx, data); err != nil {
				tx.Rollback()
				log.Println("db:", err)
				continue
			}
		case *types.Link:
			if err := updateLink(tx, []*types.Link{data}); err != nil {
				tx.Rollback()
				log.Println("db:", err)
				continue
			}
		}

		tx.Commit()
	}
}

func updateDownload(tx *sql.Tx, down *types.Download) error {
	if err := updateLink(tx, down.Links); err != nil {
		return err
	}

	now := time.Now()
	res, err := tx.Exec("UPDATE downloads SET status=?, error=?, modified_at=? WHERE id=?",
		string(down.Status), strings.Join(down.Errors, "\n"), now, down.ID)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 1 {
		down.ModifiedAt = now

	}

	return nil
}

func updateLink(tx *sql.Tx, links []*types.Link) error {
	now := time.Now()
	for _, l := range links {
		res, err := tx.Exec("UPDATE links SET status=?, modified_at=? WHERE id=?",
			string(l.Status), now, l.ID)
		if err != nil {
			return err
		}
		if n, err := res.RowsAffected(); err != nil {
			return err
		} else if n == 1 {
			l.ModifiedAt = now
		}
	}

	return nil
}
