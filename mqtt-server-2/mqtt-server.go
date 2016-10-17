package main

import (
	"fmt"
	"github.com/giovibal/go-examples/mqtt-server-2/mqtt"
	"log"
	"net"
	"net/http"
	//"github.com/gorilla/websocket"
	"golang.org/x/net/websocket"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	router := mqtt.NewRouter()
	router.Start()

	go startTcpServer(":1883", router)
	//go startWebsocketServer(":11883", router)

	// capture ctrl+c
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	select {
	case <-c:
		fmt.Println("Shutting down ...")
		os.Exit(0)
	}
}

func startTcpServer(laddr string, router *mqtt.Router) {
	ln, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Listen on %s\n", laddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn, router)
	}
}

func startWebsocketServer(addr string, router *mqtt.Router) {
	// websocket
	fmt.Printf("Listen on %s (websocket)\n", addr)

	acceptConnection := func(ws *websocket.Conn) {
		go handleConnection(ws, router)
	}

	http.Handle("/mqtt", websocket.Handler(acceptConnection))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
	//http.ListenAndServe(addr, nil)
}

func handleConnection(conn net.Conn, router *mqtt.Router) {
	c := mqtt.NewClient(conn)

	defer func() {
		conn.Close()
		log.Printf("Connection from %v closed.\n", conn.RemoteAddr())
		router.Unsubscribe(c)
	}()

	log.Printf("Connection from %v.\n", conn.RemoteAddr())

	c.Start(router)
}
func handleConnectionRW(r io.Reader, w io.WriteCloser, router *mqtt.Router) {
	c := mqtt.NewClientRW(r, w)

	defer func() {
		w.Close()
		log.Printf("Connection from %v closed.\n", c.ID)
		router.Unsubscribe(c)
	}()

	c.Start(router)

	log.Printf("Connection from %v.\n", c.ID)
}
