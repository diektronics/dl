package main

import (
	"flag"
	"log"
	"os"

	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/dl"
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

	d := dl.New(c, 5)
	d.Download(&types.Download{Name: "test1", Links: []*types.Link{
		&types.Link{URL: "http://example.com/1.rar"},
		&types.Link{URL: "http://example.com/2.rar"},
	}})
}
