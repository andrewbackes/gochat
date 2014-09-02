package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Client Object:

type Client struct {
	name    string
	DataIn  chan string
	scanner *bufio.Scanner
	writer  *bufio.Writer
}

func (C *Client) Send(text string) {
	C.writer.WriteString(fmt.Sprintln(text))
	C.writer.Flush()
}

func (C *Client) Listen() {
	for {
		C.scanner.Scan()
		data := C.scanner.Text()
		C.DataIn <- data
		if strings.HasPrefix(data, "quit") {
			break
		}
	}
}

func InitClient(c net.Conn) *Client {
	s := bufio.NewScanner(c)
	w := bufio.NewWriter(c)
	newClient := &Client{
		name:    c.RemoteAddr().String(),
		DataIn:  make(chan string),
		scanner: s,
		writer:  w,
	}
	go newClient.Listen()
	return newClient
}

// Client Handling Object:

type ClientManager struct {
	Connecting chan *Client
	Messages   chan string
	List       []*Client
}

func (CM *ClientManager) ListenTo(c *Client) {
	c.Send("You are being listened to.")
	for {
		message := <-c.DataIn
		fmt.Println(c.name + ": " + message)
	}
}

func (CM *ClientManager) Add(c net.Conn) {
	newClient := InitClient(c)
	CM.Connecting <- newClient
}

// Communication Hub:
func Broadcast(CM *ClientManager, stop chan struct{}) {
	fmt.Println("Broadcasting...")
BroadcastLoop:
	for {
		select {
		// Send message to all connected clients:
		case message := <-CM.Messages:
			for i, _ := range CM.List {
				go CM.List[i].Send(message)
			}
		// Add connecting clients to the broadcast list:
		case newClient := <-CM.Connecting:
			fmt.Println(newClient.name, "connected.")
			CM.List = append(CM.List, newClient)
			go CM.ListenTo(newClient)
		// Fan in:
		case <-stop:
			break BroadcastLoop
		}
	}
	fmt.Println("Stopped broadcasting.")
	return
}

// Host Worker:
func Serve(CM *ClientManager, stop chan struct{}) {
	fmt.Println("Awaiting connections on port 1337...")
	server, err := net.Listen("tcp", ":1337")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer server.Close()
	for {
		// Wait for a connection:
		conn, _ := server.Accept()
		// Handle the incomming connection:
		go CM.Add(conn)
	}
}

// Client Worker:
func Connect(CM *ClientManager, address string) {
	fmt.Print("Connecting to " + address + "... ")
	host, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	CM.Add(host)
	fmt.Println("Success.")
}

func main() {
	fmt.Println("\nGo Chat")
	fmt.Println("Type 'connect [ip]:[port]' or 'quit' to exit.\n")
	CM := ClientManager{
		Connecting: make(chan *Client),
		Messages:   make(chan string),
	}
	// We will close this channel to fan in all of our workers:
	stop := make(chan struct{})

	// Fan out workers:
	go Serve(&CM, stop)
	go Broadcast(&CM, stop)

	// REPL:
	console := bufio.NewScanner(os.Stdin)
	for {
		console.Scan()
		command := console.Text()
		words := strings.SplitN(command, " ", 2)

		switch words[0] {
		case "quit":
			CM.Messages <- "quit"
			close(stop)
			return
		case "connect":
			if len(words) > 0 {
				Connect(&CM, words[1])
			}
		default:
			CM.Messages <- command
		}
	}
	return
}
