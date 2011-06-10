package tcp

import (
	"time"
	"net"
	"os"
	"fmt"
)

func getConnections(listener net.Listener, out chan<- net.Conn, halt <-chan int) {
	for {
		con, err := listener.Accept()
		println("listener.Accept")
		if err == nil {
			/*
			fmt.Println("Handshake start")
						err=con.(*tls.Conn).Handshake()
						fmt.Println("Handshake err:",err)
			*/

			select {
			case _ = <-halt:
				close(out)
				return
			case out <- con:
			}
		} else {
			println("listener.Accept error:", err.String())
			_ = <- time.After(100000000) //some error trying to get a socket, so wait before retry
		}
	}
}

type Connected struct {
	Conn net.Conn
}


type Player struct {
	Name         string
	passwordHash string
}


type LoggedIn struct {
	Connected
	*Player
}

type Error int

const (
	ConnectionError Error = iota
	InvalidHeader
)

func (c *Connected) Disconnect(reason Error, err os.Error) {
	c.Conn.Close()
	fmt.Println("Disconnecting", c.Conn.RemoteAddr(), "Error:", reason, err)
}

func newConnected(con net.Conn) *Connected {
	c := &Connected{con}
	return c
}

func login(c *Connected, out chan<- *LoggedIn) {
	name, fail := c.ReadMessage()
	if !fail {
		fmt.Println("Got Name:", name)
		password, fail := c.ReadMessage()
		if !fail {
			fmt.Println("Login:", name[0], password[0], string(name[1:]), string(password[1:]))
			loginSucess := [...]byte{0, 0}[:]
			c.SendMessage(loginSucess)

			out <- &LoggedIn{*c, &Player{string(name), string(password)}}
			return
		}
	}

	fmt.Println("LOGIN FAIL")
}

func welcomTestLoop(in <-chan net.Conn, out chan<- *LoggedIn) {
	for c := range in {
		go login(newConnected(c), out)
	}
	close(out)
}

