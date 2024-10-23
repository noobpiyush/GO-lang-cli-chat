package main

import (
	"log"
	"net"
)

const PORT = "6969"

type MessageType int

const (
	ClientConnected MessageType = iota + 1
	NewMessage
	ClientDisconnect
)

type Message struct {
	Type MessageType
	Conn net.Conn
	Text string
}

func server(messages chan Message) {
	conns := map[string]net.Conn{}

	for {
		msg := <-messages
		switch msg.Type {
		case ClientConnected:
			conns[msg.Conn.RemoteAddr().String()] = msg.Conn
		case ClientDisconnect:
			delete(conns, msg.Conn.RemoteAddr().String())
		case NewMessage:
			log.Printf("Client %s send message %s ", msg.Conn.RemoteAddr(), msg.Text)
			for _, conn := range conns {
				if conn.RemoteAddr().String() != msg.Conn.RemoteAddr().String() {

					_, err := conn.Write([]byte(msg.Text))

					if err != nil {
						log.Printf("Could not send data to %s : %s", conn.RemoteAddr().String(), err)
					}
				}
			}
		}
	}
}

func client(conn net.Conn, messages chan Message) {

	buffer := make([]byte, 512)

	for {
		n, err := conn.Read(buffer)

		if err != nil {
			conn.Close()
			messages <- Message{
				Type: ClientDisconnect,
				Conn: conn,
			}
			return
		}

		messages <- Message{
			Type: NewMessage,
			Text: string(buffer[0:n]),
			Conn: conn,
		}
	}

}

func main() {
	ln, err := net.Listen("tcp", ":"+PORT)

	if err != nil {
		log.Fatalf("Error:  could not listne to epic port %s: %s\n", PORT, err)
	}

	log.Printf("Listening to TCP connections on port %s ..\n", PORT)

	messages := make(chan Message)
	go server(messages)

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Printf("Error: could not accept a connection %s\n", err)
		}
		messages <- Message{
			Type: ClientConnected,
			Conn: conn,
		}

		log.Printf("The ips are %s ", conn.RemoteAddr())
		go client(conn, messages)
	}
}
