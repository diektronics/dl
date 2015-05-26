package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"diektronics.com/carter/dl/cfg"
	"diektronics.com/carter/dl/frontend/tvd/tvd"
)

var cfgFile = flag.String(
	"cfg",
	filepath.Join(os.Getenv("HOME"), ".config", "dl", "config.cfg"),
	"Configuration file in JSON format indicating DB credentials and mailing details.",
)

const waitingTime = time.Duration(5) * time.Minute

func main() {
	flag.Parse()
	c, err := cfg.GetConfig(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	tvd.New(c).Run(waitingTime)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("entering lame duck mode for 2 seconds")
	time.Sleep(time.Duration(2) * time.Second)
}
