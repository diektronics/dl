package db

import (
	"database/sql"
	"fmt"
	"sort"

	"diektronics.com/carter/dl/frontend/tvd/show"
	"diektronics.com/carter/dl/protos/cfg"

	_ "github.com/Go-SQL-Driver/MySQL"
)

type Episode struct {
	Title    string
	Episode  string
	Location string
}

type Db struct {
	connectionString string
}

func New(c *cfg.Configuration) *Db {
	return &Db{
		connectionString: fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true&loc=Local",
			c.DbUser, c.DbPassword, c.DbServer, c.DbDatabase),
	}
}

func (d *Db) GetMyShows(titles []string) ([]*Episode, error) {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	dbQuery := "SELECT name, latest_ep, location FROM series where name IN ("
	vals := []interface{}{}
	for _, t := range titles {
		dbQuery += "?,"
		vals = append(vals, t)
	}
	dbQuery = dbQuery[0 : len(dbQuery)-1]
	dbQuery += ")"
	stmt, err := db.Prepare(dbQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	myShows := []*Episode{}
	for rows.Next() {
		eps := &Episode{}
		err = rows.Scan(&eps.Title, &eps.Episode, &eps.Location)
		if err != nil {
			return nil, err
		}
		myShows = append(myShows, eps)
	}
	return myShows, nil
}

func (d *Db) UpdateMyShows(shows []*show.Show) error {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return err
	}
	defer db.Close()
	var lastErr error
	sort.Sort(show.ByAlpha(shows))
	for _, s := range shows {
		dbQuery := fmt.Sprintf("UPDATE series SET latest_ep=%q WHERE name=%q", s.Eps, s.Name)
		_, err = db.Exec(dbQuery)
		if err != nil {
			lastErr = err
			continue
		}
	}

	return lastErr
}
