package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"diektronics.com/carter/dl/backend/dl"
	"diektronics.com/carter/dl/cfg"
)

var cfgFile = flag.String(
	"cfg",
	filepath.Join(os.Getenv("HOME"), ".config", "dl", "config.json"),
	"Configuration file in JSON format indicating DB credentials and mailing details.",
)

func main() {
	flag.Parse()
	c, err := cfg.GetConfig(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	d := dl.New(c, 5)
	if err := rpc.Register(d); err != nil {
		log.Fatal("registering:", err)
	}
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", c.BackendPort))
	if err != nil {
		log.Fatal("listening:", err)
	}
	go http.Serve(l, nil)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("entering lame duck mode for 2 seconds")
	time.Sleep(time.Duration(2) * time.Second)
}
