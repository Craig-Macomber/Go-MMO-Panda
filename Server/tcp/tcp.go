package tcp

import (
	"time"
	"net"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"os"
	"fmt"
	"crypto/tls"
)


func makeWaitChan(ns int64) <-chan int {
	waitChan := make(chan int)
	go func() {
		time.Sleep(ns)
		close(waitChan)
	}()
	return waitChan
}

func wait(ns int64, halt <-chan int) (halted bool) {
	waitChan := makeWaitChan(ns)
	select {
	case _ = <-waitChan:
		halted = false
	case _ = <-halt:
		halted = true
	}
	return
}

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
			time.Sleep(100000000) //some error trying to get a socket, so wait before retry
		}
	}
}

type Connected struct {
	Conn net.Conn
}

type Error int

const (
	ConnectionError Error = iota
	InvalidHeader
)

func (c *Connected) Disconnect(reason Error, err os.Error) {
	c.Conn.Close()
	fmt.Println("Disconnecting", c.Conn.RemoteAddr(), reason, err)
}

func (c *Connected) SendMessage(data []byte) (failed bool) {
	length := uint32(len(data) + 8)
	lengthCheck := ^(1 + length)

	header := make([]byte, 8)
	for i := uint(0); i < 4; i++ {
		header[i] = byte(length >> (i * 8))
		header[i+4] = byte(lengthCheck >> (i * 8))
	}
	_, err := c.Conn.Write(header)
	if err == nil {
		_, err = c.Conn.Write(data)
	}

	if err != nil {
		c.Disconnect(ConnectionError, err)
		return true
	}
	return false
}

func (c *Connected) ReadMessage() (data []byte, failed bool) {
	var length uint32
	var lengthCheck uint32
	err := binary.Read(c.Conn, binary.LittleEndian, &length)
	if err == nil {
		err = binary.Read(c.Conn, binary.LittleEndian, &lengthCheck)
		if err == nil {
			if lengthCheck != ^(1 + length) {
				c.Disconnect(InvalidHeader, nil)
				return
			}

			data := make([]byte, length-8)
			_, err = c.Conn.Read(data)
			if err == nil {
				return data, false
			}
		}
	}

	c.Disconnect(ConnectionError, err)
	return nil, true

}
func newConnected(con net.Conn) *Connected {
	c := &Connected{con}
	return c
}

func login(c *Connected, out chan<- *Connected) {
	c.Conn.SetReadTimeout(10)
	c.Conn.SetTimeout(10)
	//fmt.Println(c.Conn.(*tls.Conn).ConnectionState())
	name, fail := c.ReadMessage()
	if !fail{
		password, fail := c.ReadMessage()
		if !fail {
			fmt.Println("Login:", name[0], password[0], string(name[1:]), string(password[1:]))
			out <- c
			return
		}
	}
	
	fmt.Println("LOGIN FAIL")
}

func welcomTestLoop(in <-chan net.Conn, out chan<- *Connected) {
	for c := range in {
		go login(newConnected(c), out)
	}
	close(out)
}

func updateLoop(in <-chan *Connected) {
	bag := NewIterBag()
	const size = 100
	data := make([]byte, size)

	for i := 0; i < size; i++ {
		data[i] = uint8(i)
	}

	keyFrame := make([]byte, size)
	xOrData := make([]byte, size)
	frameCount := 0
	for {

		// Add any new connections
		waitChan := makeWaitChan(500000000)
	L:
		for {
			select {
			case newCon := <-in:
				if newCon == nil {
					// in is closed!
					break L
				} else {
					bag.Add(*newCon)
				}
			case _ = <-waitChan:
				break L
			}
		}

		frameCount++
		if frameCount%1 == 0 {
			println(frameCount, " - ", bag.Length())
		}
		for n := 0; n < size; n++ {
			xOrData[n] = data[n] ^ keyFrame[n]
			keyFrame[n] = data[n]
		}

		buff := new(bytes.Buffer)

		mtype := uint8(2)
		toSend := xOrData
		if frameCount%16 == 0 {
			mtype = 1
			toSend = keyFrame
		}
		buff.WriteByte(mtype)
		zlibw, _ := zlib.NewWriter(buff)
		zlibw.Write(toSend)
		zlibw.Close()

		buffBytes := buff.Bytes()

		for iter := bag.NewIterator(); iter.Current != nil; iter.Iter() {
			failed := iter.Current.SendMessage(buffBytes)

			if failed {
				iter.Remove()
				iter.GoBack()
				fmt.Println("Removing from Bag")
			}
		}
	}
}

func SetupTCP(useTls bool,address string) {
	println("Setting Up TCP at:",address)
	const connectedAndWaitingMax = 0
	conChan := make(chan net.Conn, connectedAndWaitingMax)
	halt := make(chan int)
	
	var listener net.Listener
	var err os.Error
	if useTls{
		certs:=make([]tls.Certificate,1)
		c0,errx:=tls.LoadX509KeyPair("cert/cert.pem", "cert/key.pem")
		certs[0]=c0
		fmt.Println(errx)
		config:=tls.Config{Certificates:certs,ServerName:"TestServer"}
		listener, err = tls.Listen("tcp", ":6666",&config)
		println("TLS")
	} else {
		listener, err = net.Listen("tcp", ":6666")
		println("TCP")
	}
	
	
	if err != nil {
		println(err)
	}

	go getConnections(listener, conChan, halt)

	conChan2 := make(chan *Connected, connectedAndWaitingMax)

	go welcomTestLoop(conChan, conChan2)
	go updateLoop(conChan2)

	println("TCP Setup")

}
