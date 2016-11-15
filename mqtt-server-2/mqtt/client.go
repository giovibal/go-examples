package mqtt

import (
	"bufio"
	"github.com/giovibal/go-examples/mqtt-server-2/packets"
	"net"
	"log"
)

type Client struct {
	Conn              net.Conn
	ID                string
	CleanSession      bool
	WillMessage       *packets.PublishPacket
	Keepalive         *Keepalive
	reader            *bufio.Reader
	writer            *bufio.Writer
	queue             *Queue
	incoming          chan packets.ControlPacket
	outgoing          chan packets.ControlPacket
	Subscriptions     map[string]*Subscription
	SubscriptionCache map[string]bool
	quit              chan bool
}

func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	client := &Client{
		Conn:              connection,
		reader:            reader,
		writer:            writer,
		queue:             NewQueue(),
		incoming:          make(chan packets.ControlPacket),
		outgoing:          make(chan packets.ControlPacket),
		Subscriptions:     make(map[string]*Subscription),
		SubscriptionCache: make(map[string]bool),
		quit:              make(chan bool),
	}
	return client
}

func (client *Client) Read() {
	for {
		//log.Println("reading ...")
		data, err := packets.ReadPacket(client.reader)
		if err != nil {
			break
		}
		client.incoming <- data
	}
}
func (client *Client) Write() {
	for data := range client.outgoing {
		//log.Println("writing ...")
		//log.Printf(">> %v \n", data)
		err := data.Write(client.writer)
		if err != nil {
			log.Println(err)
		}
		err2 := client.writer.Flush()
		if err2 != nil {
			log.Println(err)

			// enqueue messsage if error and cleanSession is false
			if !client.CleanSession {
				switch data.(type) {
				case *packets.PublishPacket:
					pubMsg := data.(*packets.PublishPacket)
					if pubMsg.Qos == 1 || pubMsg.Qos == 2 {
						client.queue.EnqueueMessage(pubMsg)
					}
					break
				}
			}
		}
	}
}
func (client *Client) Quit() {
	log.Printf("Quitting client %s ...", client.ID)
	client.Keepalive.Stop()
	client.Conn.Close()
	client.quit <- true
}
func (client *Client) waitForQuit() {
	select {
	case ret := <-client.quit:
		if ret {
			return
		}
	}
}
func (c *Client) IsSubscribed(publishingTopic string) bool {
	_, present := c.SubscriptionCache[publishingTopic]
	if present {
		return true
	} else {
		for _, subscription := range c.Subscriptions {
			if subscription.IsSubscribed(publishingTopic) {
				c.SubscriptionCache[publishingTopic] = true
				return true
			}
		}
	}
	return false
}
func (c *Client) WritePublishMessage(msg *packets.PublishPacket) {
	//go func(ch chan packets.ControlPacket) {
	//	ch <- msg
	//}(c.outgoing)
	c.outgoing <- msg
}

func (client *Client) Start(router *Router) {
	go client.Read()
	go client.Write()
	client.waitForQuit()
}

func (client *Client) CopyTo(other *Client) {
	other.SubscriptionCache = client.SubscriptionCache
	other.Subscriptions= client.Subscriptions
	other.queue = client.queue
}

func (client *Client) FlushQueuedMessages() {
	len := client.queue.queue.Len()
	if len > 0 {
		log.Printf("Flush queued messages (%v)", len)
		for msg := client.queue.DequeueMessage(); msg != nil; msg = client.queue.DequeueMessage() {
			client.WritePublishMessage(msg.(*packets.PublishPacket))
		}
	}
}

//func (client *Client) IsConnected() bool {
//	client.Conn.SetReadDeadline(time.Now().Add(500*time.Millisecond))
//	_, err := client.Conn.Read(make([]byte, 0))
//	client.Conn.SetReadDeadline(time.Time{})
//	if err!=io.EOF{
//		return false
//	} else {
//		return true
//	}
//}

