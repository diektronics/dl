package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/dl"
	"diektronics.com/carter/dl/server"
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

	// TODO(diek): do recovery, all that is RUNNING becomes QUEUED, then all that is QUEUED goes to the queue.
	d := dl.New(c, 5)
	server.New(d).Run()

	d.Download(&types.Download{Name: "Test 1",
		Links: []*types.Link{
			&types.Link{URL: "http://www.uploadable.ch/file/sNfB6fVsMmy7/test1.part1.rar"},
			&types.Link{URL: "http://www.uploadable.ch/file/29nM4NpTJ8HH/test1.part2.rar"},
			&types.Link{URL: "http://www.uploadable.ch/file/AHzPVTgAZExM/test1.part3.rar"},
		},
		Posthook: "UNRAR,REMOVE",
	})

	d.Download(&types.Download{Name: "Test 2",
		Links: []*types.Link{
			&types.Link{URL: "http://www.uploadable.ch/file/sNfB6fVsMmy7/test1.part1.rar"},
			&types.Link{URL: "http://www.uploadable.ch/file/29nM4NpTJ8HH/test1.part2.rar"},
			&types.Link{URL: "http://www.uploadable.ch/file/AHzPVTgAZExM/test1.part3.rar"},
		},
		Posthook: "UNRAR,REMOVE",
	})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("entering lame duck mode for 2 seconds")
	time.Sleep(2 * time.Second)
}
