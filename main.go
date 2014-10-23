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
	server.New(dl.New(c, 5)).Run()

	// d.Download(&types.Download{Name: "The Hobbit The Desolation of Smaug (2013) EXTENDED",
	// 	Links: []*types.Link{
	// 		&types.Link{URL: "http://www.uploadable.ch/file/HFvfndwqKKdj/thtdos13.ex.1080p-gec.part01.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/RshFX8MnFcYB/thtdos13.ex.1080p-gec.part02.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/v2xUG3bV3dR4/thtdos13.ex.1080p-gec.part03.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/8phcAxphQEJR/thtdos13.ex.1080p-gec.part04.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/cUvU9z8jdmJ5/thtdos13.ex.1080p-gec.part05.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/PfV3QpwNuCJV/thtdos13.ex.1080p-gec.part06.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/yRkh2NkUhxDP/thtdos13.ex.1080p-gec.part07.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/U7VtJBwK4r4x/thtdos13.ex.1080p-gec.part08.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/PB3rG9R5N7tr/thtdos13.ex.1080p-gec.part09.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/NrcrC9Ps9YKB/thtdos13.ex.1080p-gec.part10.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/CVFrewUMWQ2B/thtdos13.ex.1080p-gec.part11.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/t8pVzjrKvMcp/thtdos13.ex.1080p-gec.part12.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/hX2cHG5c5NNu/thtdos13.ex.1080p-gec.part13.rar"},
	// 		&types.Link{URL: "http://www.uploadable.ch/file/fGzxbgEPspBD/thtdos13.ex.1080p-gec.part14.rar"},
	// 	}})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("entering lame duck mode for 2 seconds")
	time.Sleep(2 * time.Second)
}
