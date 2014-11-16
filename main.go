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

	d := dl.New(c, 5)
	if err := d.Recovery(); err != nil {
		log.Fatal(err)
	}
	server.New(d, c).Run()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("entering lame duck mode for 2 seconds")
	time.Sleep(2 * time.Second)
}
