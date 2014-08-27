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

func (C *Client) Send(message string) {
	C.writer.WriteString(fmt.Sprintln(message))
	C.writer.Flush()
}

func (C *Client) Listen() {
	for {
		C.scanner.Scan()
		data := C.scanner.Text()
		C.DataIn <- data
	}
}

// Client Handling Object:

type ClientManager struct {
	Connecting chan *Client
	Messages   chan string
}

func (CM *ClientManager) ListenTo(c *Client) {
	c.Send(fmt.Sprintln("You are being listened to."))
	for {
		message := <-c.DataIn
		CM.Messages <- c.name + ": " + message
	}
}

func (CM *ClientManager) Add(c net.Conn) {
	s := bufio.NewScanner(c)
	w := bufio.NewWriter(c)
	newClient := &Client{
		name:    c.RemoteAddr().String(),
		DataIn:  make(chan string),
		scanner: s,
		writer:  w,
	}
	go newClient.Listen()
	CM.Connecting <- newClient
}

// Workers:

func Broadcast(CM *ClientManager, stop chan struct{}) {
	fmt.Println("Broadcasting...")
	var ClientList []*Client
	for {
		select {
		// Send message to all connected clients:
		case message := <-CM.Messages:
			fmt.Println(message)
			for i, _ := range ClientList {
				go ClientList[i].Send(message)
			}
		// Add connecting clients to the broadcast list:
		case newClient := <-CM.Connecting:
			fmt.Println(newClient.name, "connected.")
			ClientList = append(ClientList, newClient)
			go CM.ListenTo(newClient)
		// Fan in:
		case <-stop:
			break
		}
	}
	fmt.Println("Stopped broadcasting.")
	return
}

func Serve(CM *ClientManager, stop chan struct{}) {
	fmt.Println("Awaiting connections on port 1337...")
	server, _ := net.Listen("tcp", ":1337")
	defer server.Close()
	for {
		// Wait for a connection:
		conn, _ := server.Accept()
		// Handle the incomming connection:
		go CM.Add(conn)
	}
}

func Connect(CM *ClientManager, address string) {
	fmt.Print("Connecting to " + address + "... ")
	host, err := net.Dial("tcp", address)
	if err == nil {
		CM.Add(host)
		fmt.Println("Success.")
	} else {
		fmt.Println(err.Error())
	}
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

	go Broadcast(&CM, stop)

	// REPL:
	//console := bufio.NewReader(os.Stdin)
	console := bufio.NewScanner(os.Stdin)
	for {
		console.Scan()
		command := console.Text()
		//command, _ := console.ReadString('\n')
		//command = strings.Trim(command, "\n")
		words := strings.SplitN(command, " ", 2)

		switch words[0] {
		case "quit":
			close(stop)
			return
		case "connect":
			go Connect(&CM, words[1])
		case "host":
			go Serve(&CM, stop)
		case "say":
			CM.Messages <- words[1]
		default:
			CM.Messages <- command
		}
	}
	return
}
