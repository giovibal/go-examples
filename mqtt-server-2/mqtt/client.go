package mqtt

import (
	"bufio"
	"github.com/giovibal/go-examples/mqtt-server-2/packets"
	"io"
	"net"
	"log"
)

type Client struct {
	Conn              net.Conn
	ID                string
	reader            *bufio.Reader
	writer            *bufio.Writer
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
		incoming:          make(chan packets.ControlPacket),
		outgoing:          make(chan packets.ControlPacket),
		Subscriptions:     make(map[string]*Subscription),
		SubscriptionCache: make(map[string]bool),
		quit:              make(chan bool),
	}
	return client
}
func NewClientRW(r io.Reader, w io.Writer) *Client {
	reader := bufio.NewReader(r)
	writer := bufio.NewWriter(w)

	client := &Client{
		//Conn:     connection,
		reader:            reader,
		writer:            writer,
		incoming:          make(chan packets.ControlPacket),
		outgoing:          make(chan packets.ControlPacket),
		Subscriptions:     make(map[string]*Subscription),
		SubscriptionCache: make(map[string]bool),
	}
	return client
}

// tcp conn >>> channel
func (c *Client) handleMqttProtocol(router *Router) {
	for {
		select {
		case cp := <-c.incoming:
			//log.Printf(">> %v \n", cp.String())

			//msgType := cp.GetMessageType()
			switch cp.(type) {

			case *packets.ConnectPacket :
				connectMsg := cp.(*packets.ConnectPacket)
				c.ID = connectMsg.ClientIdentifier
				proto := connectMsg.ProtocolName
				if proto == "MQTT" {
					log.Printf("MQTT 3.1.1 (%s)\n", proto)
				} else if proto == "MQIsdp" {
					log.Printf("MQTT 3.1 (%s)\n", proto)
				} else {
					log.Printf("Wrong protocol (%s)\n", proto)
					c.Quit()
					break
				}

				if router.Connected(c) {
					c.Quit()
				} else {
					router.Connected(c);
				}

				connackMsg := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
				connackMsg.SessionPresent = false
				c.outgoing <- connackMsg
				break
			case *packets.SubscribePacket:
				subackMsg := packets.NewControlPacket(packets.Suback).(*packets.SubackPacket)
				subackMsg.MessageID = cp.Details().MessageID
				router.Subscribe(c)

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

				router.RepublishRetainedMessages(c, subMsg)

				c.outgoing <- subackMsg
				break
			case *packets.UnsubscribePacket:
				router.Unsubscribe(c)
				unsubackMsg := packets.NewControlPacket(packets.Unsuback).(*packets.UnsubackPacket)
				unsubackMsg.MessageID = cp.Details().MessageID
				c.outgoing <- unsubackMsg
				break
			case *packets.DisconnectPacket:
				router.Unsubscribe(c)
				router.Disconnect(c)
				c.Quit()
				break
			case *packets.PublishPacket:
				qos := cp.Details().Qos
				pubMsg := cp.(*packets.PublishPacket)

				switch qos {
				case 0:
					router.Publish(pubMsg)
					break
				case 1:
					router.Publish(pubMsg)
					pubAck := packets.NewControlPacket(packets.Puback).(*packets.PubackPacket)
					pubAck.MessageID = cp.Details().MessageID
					c.outgoing <- pubAck
					break
				case 2:
					router.Publish(pubMsg)
					pubRec := packets.NewControlPacket(packets.Pubrec).(*packets.PubrecPacket)
					pubRec.MessageID = cp.Details().MessageID
					c.outgoing <- pubRec
					break
				}
			case *packets.PubrecPacket:
				pubRel := packets.NewControlPacket(packets.Pubrel).(*packets.PubrelPacket)
				pubRel.MessageID = cp.Details().MessageID
				c.outgoing <- pubRel
				break

			case *packets.PubrelPacket:
				pubComp := packets.NewControlPacket(packets.Pubcomp).(*packets.PubcompPacket)
				pubComp.MessageID = cp.Details().MessageID
				c.outgoing <- pubComp
				break

			case *packets.PubackPacket:
				break

			case *packets.PingreqPacket:
				pingresp := packets.NewControlPacket(packets.Pingresp).(*packets.PingrespPacket)
				c.outgoing <- pingresp
				break

			default:
				c.Quit()
			}

		}
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
		//log.Printf("<< %v \n", data.String())
		err := data.Write(client.writer)
		if err != nil {
			log.Fatal(err)
			return
		}
		err2 := client.writer.Flush()
		if err2 != nil {
			log.Fatal(err2)
			return
		}
	}
}
func (client *Client) Quit() {
	client.quit <- true
}
func (client *Client) waitForQuit() bool {
	select {
	case ret := <-client.quit:
		return ret
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
	//c.outgoing <- msg
}

func (client *Client) Start(router *Router) {
	go client.handleMqttProtocol(router)
	go client.Read()
	go client.Write()
	client.waitForQuit()
}
