package tcp

import (
	"time"
	"net"
	//"/iterBag"
	//"io"
	//"io/ioutil"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"os"
	"fmt"
)


func wait(ns int64, halt <-chan int) (halted bool) {
	waitChan := make(chan int)
	go func() {
		time.Sleep(ns)
		close(waitChan)
	}()
	select {
	case _ = <-waitChan:
		halted = false
	case _ = <-halt:
		halted = true
	}
	return
}

func getConnections(listener net.Listener, out chan<- net.Conn) {
	for {
		println("listener listening")
		con, err := listener.Accept()
		if err == nil {
			select {
			//             case _ = <- halt:
			//                 close(out)
			//                 return
			case out <- con:
			}
		} else {
			println(err.String())
			time.Sleep(100000000) //some error trying to get a socket, so wait before retry
		}
	}
}

type Connected struct {
	Conn net.Conn
	In   <-chan []byte
}

type Error int

const (
	ConnectionError Error = iota
	InvalidHeader
)

func (c *Connected) Disconnect(reason Error, err os.Error) {
	c.Conn.Close()
	close(c.In)
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


func newConnected(con net.Conn) *Connected {
	in := make(chan []byte)
	c := &Connected{con, in}
	//con.SetReadTimeout(1e9) // 1 second time out
	go func() {
		// read messages from con to in
		var length uint32
		var lengthCheck uint32
		var err os.Error = nil
		for {
			err = binary.Read(con, binary.LittleEndian, &length)
			if err != nil {
				break
			}
			err = binary.Read(con, binary.LittleEndian, &lengthCheck)
			if err != nil {
				break
			}

			if lengthCheck != ^(1 + length) {
				c.Disconnect(InvalidHeader, nil)
				return
			}

			data := make([]byte, length-8)
			_, err = con.Read(data)
			if err == nil {
				//fmt.Println("Got Data:", string(data))
				in <- data
			} else {
				break
			}
		}
		c.Disconnect(ConnectionError, err)
	}()
	return c
}

func connector(in <-chan net.Conn, out chan<- *Connected) {
	for c := range in {
		println("Adding newConnected")
		out <- newConnected(c)
	}
	close(out)
}


func welcomTestLoop(in <-chan *Connected, bag *IterBag) {
	for c := range in {
		go func() {
			name := <-c.In
			password := <-c.In
			fmt.Println("Login:", string(name[1:]), string(password[1:]))
			bag.Add(*c)
		}()
	}
}

func updateLoop(bag *IterBag) {
	const size = 100
	data := make([]byte, size)

	for i := 0; i < size; i++ {
		data[i] = uint8(i)
	}

	keyFrame := make([]byte, size)
	xOrData := make([]byte, size)
	frameCount := 0
	for {
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

		iter := bag.NewIterator()
		for ; iter.Current != nil; iter.Iter() {
			failed := iter.Current.SendMessage(buffBytes)

			if failed {
				iter.Remove()
				iter.GoBack()
			}
		}
		time.Sleep(500000000)
	}
}

func SetupTCP() {
	println("Setting Up TCP")
	const connectedAndWaitingMax = 4
	conChan := make(chan net.Conn, connectedAndWaitingMax)
	listener, err := net.Listen("tcp", "127.0.0.1:6666")
	if err != nil {
		println(err)
	}

	go getConnections(listener, conChan)

	connectedChan := make(chan *Connected)

	go connector(conChan, connectedChan)

	bag := NewIterBag()
	go welcomTestLoop(connectedChan, bag)
	go updateLoop(bag)

	println("TCP Setup")

}
