package main

import (
	"github.com/giovibal/go-examples/mqtt-server/client"
	"github.com/giovibal/go-examples/mqtt-server/packets"
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":1883")
	if err != nil {
		log.Fatal(err)
	}

	publishchan := make(chan *packets.PublishPacket, 8)
	addchan := make(chan *client.Client, 4)
	rmchan := make(chan *client.Client, 4)

	go handleMessages(publishchan, addchan, rmchan)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn, publishchan, addchan, rmchan)
	}
}

func handleConnection(conn net.Conn, publishchan chan<- *packets.PublishPacket, addchan chan<- *client.Client, rmchan chan<- *client.Client) {
	//bufconn := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	c := client.New(conn)
	defer func() {
		conn.Close()
		log.Printf("Connection from %v closed.\n", conn.RemoteAddr())
		rmchan <- c
	}()

	log.Printf("Connection from %v.\n", conn.RemoteAddr())
	go c.ReadMessagesInto(publishchan, addchan, rmchan)
	c.WriteMessagesFrom(c.Ch)
}

func handleMessages(publishchan <-chan *packets.PublishPacket, addchan <-chan *client.Client, rmchan <-chan *client.Client) {
	clients := make(map[net.Conn]*client.Client)
	for {
		select {
		case msg := <-publishchan:
			// example of publish to all clients...
			for _, c := range clients {
				// check subscription match
				publishingTopic := msg.TopicName
				ok, _ := c.IsSubscribed(publishingTopic)
				if ok {
					// TODO: look at mqtt specs for qos handling...
					//msg.Qos = subscriptionQos
					c.WritePublishMessage(msg)
				}
			}

		case client := <-addchan:
			clients[client.Conn] = client
		case client := <-rmchan:
			delete(clients, client.Conn)
		}
	}
}
