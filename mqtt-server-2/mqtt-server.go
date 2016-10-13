package main

import (
	"github.com/giovibal/go-examples/mqtt-server-2/mqtt"
	"log"
	"net"
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
	"os"
	"os/signal"
	"syscall"
	"io"
)

func main() {

	router := mqtt.NewRouter()
	router.Start()

	go startTcpServer(":1883", router)
	//go startTcpServer(":1885", router)
	//go startWebsocketServer(":11883", router)

	// capture ctrl+c and stop CPU profiler
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

	//router := mqtt.NewRouter()
	//router.Start()

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
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	http.HandleFunc("/mqtt", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		for {
			messageType, reader, err := conn.NextReader()
			if err != nil {
				log.Println(err)
				continue
			}
			writer, err := conn.NextWriter(messageType)
			if err != nil {
				log.Println(err)
				continue
			}
			//if _, err := io.Copy(w, r); err != nil {
			//	return err
			//}
			go handleConnectionRW(reader, writer, router)

			if err := writer.Close(); err != nil {
				log.Println(err)
				continue
			}
		}
	})
	http.ListenAndServe(addr, nil)
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
