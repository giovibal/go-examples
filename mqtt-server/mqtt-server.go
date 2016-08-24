package main

import (
	"log"
	"net"
	"github.com/giovibal/go-examples/mqtt-server/packets"
)

func main() {
	ln, err := net.Listen("tcp", ":6000")
	if err != nil {
		log.Fatal(err)
	}

	msgchan := make(chan packets.ControlPacket)
	addchan := make(chan Client)
	rmchan := make(chan Client)

	go handleMessages(msgchan, addchan, rmchan)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn, msgchan, addchan, rmchan)
	}
}

func handleConnection(c net.Conn, msgchan chan<- packets.ControlPacket, addchan chan<- Client, rmchan chan<- Client) {
	//bufc := bufio.NewReader(c)
	defer c.Close()

	// create the client instance
	client := Client{
		conn:     c,
		ch:       make(chan packets.ControlPacket),
	}


	// send "new client" message/event
	addchan <- client
	// ensure "remove client" message/event at end
	defer func() {
		//msgchan <- fmt.Sprintf("User %s left the chat room.\n", client.conn)
		log.Printf("Connection from %v closed.\n", c.RemoteAddr())
		rmchan <- client
	}()

	//io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", client.conn))
	//msgchan <- fmt.Sprintf("New user %s has joined the chat room.\n", client.conn)

	msgchan <- packets.NewControlPacket(packets.Connack);

	go client.ReadMessagesInto(msgchan)
	client.WriteMessagesFrom(client.ch)
}


func handleMessages(msgchan <-chan packets.ControlPacket, addchan <-chan Client, rmchan <-chan Client) {
	clients := make(map[net.Conn]chan packets.ControlPacket)
	for {
		select {
		case msg := <-msgchan:
			// example of publish to all clients...
			for _, ch := range clients {
				go func(mch chan packets.ControlPacket) {
					mch <- msg
				}(ch)
			}

		case client := <-addchan:
			clients[client.conn] = client.ch
		case client := <-rmchan:
			delete(clients, client.conn)
		}
	}
}



// Client
type Client struct {
	conn net.Conn
	ch   chan packets.ControlPacket
}

// tcp conn >>> channel
func (c Client) ReadMessagesInto(ch chan<- packets.ControlPacket) {
	var err error
	var cp packets.ControlPacket
	for {
		if cp, err = packets.ReadPacket(c.conn); err != nil {
			break
		}
		log.Printf(">> %v \n", cp.String())
		ch <- cp
	}
}

// tcp conn <<< channel
func (c Client) WriteMessagesFrom(ch <-chan packets.ControlPacket) {
	for msg := range ch {
		//_, err := io.WriteString(c.conn, msg)
		err := msg.Write(c.conn)
		if err != nil {
			return
		}
	}
}
