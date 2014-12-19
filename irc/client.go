package irc

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type Client struct {
	// Provides Nick, Host, and User.
	*Prefix

	// Returns whether the client is connected.
	IsConnected bool

	// Realname.
	Name string

	// Logging.
	//logger log.Logger

	// Connection to server.
	conn net.Conn

	// Messages to be sent to server.
	outQueue chan *Message

	// Channel for stopping the listener.
	stopListener chan int

	// Channel for indicating a successful handshake.
	hsChan chan int

	// State variables
	listenerRunning bool
	quitSent        bool
}

func NewClient(user *string, nick *string, name *string, host *string) *Client {
	log.Println("Making Client")
	c := new(Client)
	c.Name = *name
	c.Prefix = &Prefix{*nick, *host, *user}
	// TODO configure buffer size
	c.outQueue = make(chan *Message, 25)
	//c.logger = l
	c.stopListener = make(chan int, 1)
	c.hsChan = make(chan int)
	return c
}

// Connects to the server specified by address. A client can connect
// to exactly one server.
// Messages from the server will go to in.
func (c *Client) Connect(addr *string, pw *string, in chan<- Message) error {
	if c.IsConnected {
		// TODO better error
		return fmt.Errorf("tirc: Client is already connected.")
	}
	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		return err
	}
	c.conn = conn
	// Start listening on the server.
	go c.listener(in)
	// Start the message sender.
	go c.sender()

	c.sendHandshake(pw)
	select {
	case <-c.hsChan:
		log.Println("Connection established")
		c.IsConnected = true
	case <-time.After(3 * time.Second):
		return fmt.Errorf("tirc: Handshake received no reply.")
	}

	return nil
}

func (c *Client) sendHandshake(pw *string) {
	if *pw != "" {
		c.Send(*NewPassMessage(*c.Prefix, *pw))

	}
	c.Send(*NewNickMessage(*c.Prefix, c.Nick))
	c.Send(*NewUserMessage(*c.Prefix, c.User, "0", c.Name))
}

// Send message to server. Blocks on outQueue
func (c *Client) Send(m Message) {
	c.outQueue <- &m
}

func (c *Client) sender() {
	var out string
	for m := range c.outQueue {
		out = (*m).String()
		log.Printf("SEND: %s", out)
		fmt.Fprintf(c.conn, out)
	}
}

// Send QUIT with qmessage to server and close the connection.
// TODO: Should this block?
func (c *Client) Quit(qmessage string) {
	c.Send(*NewQuitMessage(*c.Prefix, qmessage))
	c.quitSent = true
}

func (c *Client) Disconnect() (err error) {
	if c.listenerRunning {
		c.stopListener <- 1
	}
	close(c.outQueue)
	err = c.conn.Close()
	c.conn = nil
	log.Print("Disconnected from server")
	return
}

// Read and parse messages. Send them into out.
// Replies to PING with PONG.
func (c *Client) listener(out chan<- Message) {
	r := bufio.NewReader(c.conn)
	c.listenerRunning = true
	for {
		select {
		case <-c.stopListener:
			c.listenerRunning = false
			close(out)
			log.Print("Listener stopped")
			return
		default:
			c.readAndHandleLine(r, out)
		}
	}
}

func (c *Client) readAndHandleLine(r *bufio.Reader, out chan<- Message) {
	line, err := r.ReadString('\n')
	switch err {
	case nil:
		m, err := ParseMessage(line)
		if err != nil {
			log.Printf("%v\n", err)
		}
		forward := true
		switch m.Command {
		case "001": // RPL_WELCOME
			c.hsChan <- 1
		case "PING":
			c.Send(*NewPongMessage(*c.Prefix, *m.Params["server1"], ""))
			forward = false
		case "ERROR":
			if c.quitSent {
				c.Disconnect()
			}
			forward = false
		}
		if forward {
			out <- *m
		}
	case io.EOF:
		log.Printf("Reached EOF, disconnecting.\n")
		c.IsConnected = false
		c.Disconnect()
		return
	default:
		log.Printf("Reader error: %s\n", err)
	}
}
