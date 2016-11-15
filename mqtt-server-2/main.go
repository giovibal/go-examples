package main

import (
	"fmt"
	"github.com/giovibal/go-examples/mqtt-server-2/mqtt"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	router := mqtt.NewRouter()
	router.Start()

	go mqtt.StartTcpServer(":1883", router)
	go mqtt.StartWebsocketServer(":11883", router)

	// capture ctrl+c
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	select {
	case <-c:
		fmt.Println("Shutting down ...")
		os.Exit(0)
	}
}
