package mqtt

import (
	"bufio"
	"github.com/giovibal/go-examples/mqtt-server-2/packets"
	"net"
	"io"
)

type Client struct {
	Conn          net.Conn
	ID            string
	reader        *bufio.Reader
	writer        *bufio.Writer
	incoming      chan packets.ControlPacket
	outgoing      chan packets.ControlPacket
	Subscriptions map[string]*Subscription
	SubscriptionCache map[string]bool
}

func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	client := &Client{
		Conn:     connection,
		reader:   reader,
		writer:   writer,
		incoming: make(chan packets.ControlPacket),
		outgoing: make(chan packets.ControlPacket),
		Subscriptions: make(map[string]*Subscription),
		SubscriptionCache: make(map[string]bool),
	}
	return client
}
func NewClientRW(r io.Reader, w io.Writer) *Client {
	reader := bufio.NewReader(r)
	writer := bufio.NewWriter(w)

	client := &Client{
		//Conn:     connection,
		reader:   reader,
		writer:   writer,
		incoming: make(chan packets.ControlPacket),
		outgoing: make(chan packets.ControlPacket),
		Subscriptions: make(map[string]*Subscription),
		SubscriptionCache: make(map[string]bool),
	}
	return client
}

// tcp conn >>> channel
func (c *Client) handleMqttProtocol(router *Router) {
	for {
		select {
		case cp := <- c.incoming:
			//log.Printf(">> %v \n", cp.String())

			msgType := cp.GetMessageType()
			switch msgType {

			case packets.Connect:
				connectMsg := cp.(*packets.ConnectPacket)
				c.ID = connectMsg.ClientIdentifier

				connackMsg := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
				c.outgoing <- connackMsg
				break
			case packets.Subscribe:
				//router.Addchan <- c
				router.Subscribe(c)
				subackMsg := packets.NewControlPacket(packets.Suback).(*packets.SubackPacket)
				subackMsg.MessageID = cp.Details().MessageID

				subMsg := cp.(*packets.SubscribePacket)
				topics := subMsg.Topics
				qoss := subMsg.Qoss

				for i := 0; i < len(topics); i++ {
					topic := topics[i]
					var qos byte
					if i < len(qoss) {
						qos = qoss[i]
					} else {
						qos = 0x0
					}
					s := NewSubscription(topic, qos)
					c.Subscriptions[topic] = s
				}

				c.outgoing <- subackMsg
				break
			case packets.Unsubscribe:
				//router.Rmchan <- c
				router.Unsubscribe(c)
				c.Subscriptions = make(map[string]*Subscription)
				unsubackMsg := packets.NewControlPacket(packets.Unsuback).(*packets.UnsubackPacket)
				unsubackMsg.MessageID = cp.Details().MessageID
				c.outgoing <- unsubackMsg
				break
			case packets.Disconnect:
				break
			case packets.Publish:
				qos := cp.Details().Qos
				pubMsg := cp.(*packets.PublishPacket)
				switch qos {
				case 0:
					//router.Publishchan <- pubMsg
					router.Publish(pubMsg)
					break
				case 1:
					//router.Publishchan <- pubMsg
					router.Publish(pubMsg)
					pubAck := packets.NewControlPacket(packets.Puback).(*packets.PubackPacket)
					pubAck.MessageID = cp.Details().MessageID
					c.outgoing <- pubAck
					break
				case 2:
					//router.Publishchan <- pubMsg
					router.Publish(pubMsg)
					pubRec := packets.NewControlPacket(packets.Pubrec).(*packets.PubrecPacket)
					pubRec.MessageID = cp.Details().MessageID
					c.outgoing <- pubRec
					break
				}
			case packets.Pubrec:
				pubRel := packets.NewControlPacket(packets.Pubrel).(*packets.PubrelPacket)
				pubRel.MessageID = cp.Details().MessageID
				c.outgoing <- pubRel
				break

			case packets.Pubrel:
				pubComp := packets.NewControlPacket(packets.Pubcomp).(*packets.PubcompPacket)
				pubComp.MessageID = cp.Details().MessageID
				c.outgoing <- pubComp
				break

			case packets.Puback:
				break

			case packets.Pingreq:
				pingresp := packets.NewControlPacket(packets.Pingresp).(*packets.PingrespPacket)
				c.outgoing <- pingresp
				break

			default:
			}

		}
	}
}


// tcp conn <<< channel
func (client *Client) WriteMessagesFrom(ch <-chan *packets.PublishPacket) {
	for msg := range ch {
		err := msg.Write(client.writer)
		if err != nil {
			return
		}
		client.writer.Flush()
	}
}

func (client *Client) Read() {
	for {
		data, err := packets.ReadPacket(client.reader)
		if err != nil {
			break
		}
		client.incoming <- data
	}
}
func (client *Client) Write() {
	for data := range client.outgoing {
		err := data.Write(client.writer)
		if err != nil {
			return
		}
		client.writer.Flush()
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
	go func(ch chan packets.ControlPacket) {
		ch <- msg
	}(c.outgoing)
}

func (client *Client) Start(router *Router) {
	go client.handleMqttProtocol(router)
	go client.Read()
	client.Write()
}
