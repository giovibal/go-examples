package main

import (
	"github.com/giovibal/go-examples/mqtt-server-2/mqtt"
	"log"
	"net"
	"fmt"
)

func main() {
	ln, err := net.Listen("tcp", ":1883")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Listen on %s\n", "1883")

	router := mqtt.NewRouter()
	go router.RouteMessages()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn, router)
	}
}

func handleConnection(conn net.Conn, router *mqtt.Router) {
	c := mqtt.NewClient(conn)
	defer func() {
		conn.Close()
		log.Printf("Connection from %v closed.\n", conn.RemoteAddr())
		router.Unsubscribe(c)
	}()

	log.Printf("Connection from %v.\n", conn.RemoteAddr())

	go c.HandleMqttProtocol(router)
	c.Listen()
}

