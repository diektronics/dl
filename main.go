package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/db"
	"diektronics.com/carter/dl/types"
)

var cfgFile = flag.String(
	"cfg",
	os.Getenv("HOME")+"/.config/dl/config.json",
	"Configuration file in JSON format indicating DB credentials and mailing details.",
)

func main() {
	flag.Parse()
	c, err := cfg.GetConfig(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	d := &types.Download{Name: "test1", Links: []*types.Link{
		&types.Link{Url: "http://example.com/1.rar"},
		&types.Link{Url: "http://example.com/2.rar"},
	}}
	dataBase := db.New(c)
	if err := dataBase.Add(d); err != nil {
		log.Fatal(err)
	}
	fmt.Println(d)
	d2, err := dataBase.Get(d.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(d2)
	d2.Status = types.Running
	if err := dataBase.Update(d2); err != nil {
		log.Fatal(err)
	}
	d3, err := dataBase.Get(d.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(d3)
	if err := dataBase.Del(d3); err != nil {
		log.Fatal(err)
	}

	_, err = dataBase.Get(d.ID)
	fmt.Println(err)
}
