package db

import (
	"database/sql"
	"fmt"
	"strings"

	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/types"

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

	dbQuery := fmt.Sprintf("SELECT name, latest_ep, location FROM series where name IN (%s)", strings.Join(titles, ","))
	rows, err := db.Query(dbQuery)
	if err != nil {
		return nil, err
	}

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

func (d *Db) UpdateMyShows(shows []*types.Show) error {
	db, err := sql.Open("mysql", d.connectionString)
	if err != nil {
		return err
	}
	defer db.Close()
	var lastErr error
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
