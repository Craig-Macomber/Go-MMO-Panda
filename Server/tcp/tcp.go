package tcp

import (
	"time"
	"net"
	//"/iterBag"
	//"io"
	//"io/ioutil"
	"bytes"
	"compress/zlib"
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

func getConnections(listener net.Listener, out chan<- *net.Conn) {
	for {
		println("listener listening")
		con, err := listener.Accept()
		if err == nil {
			select {
			//             case _ = <- halt:
			//                 close(out)
			//                 return
			case out <- &con:
			}
		} else {
			println(err.String())
			time.Sleep(100000000) //some error trying to get a socket, so wait before retry
		}
	}
}

type Connected struct {
	Conn *net.Conn
}

func newConnected(con *net.Conn) *Connected {
	return &Connected{con}
}

func connector(in <-chan *net.Conn, out chan<- *Connected) {
	for c := range in {
		println("Adding newConnected")
		out <- newConnected(c)
	}
	close(out)
}


func welcomTestLoop(in <-chan *Connected, bag *IterBag) {
	for c := range in {
		bag.Add(*c)
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
		if frameCount%100 == 0 {
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

		length := uint32(buff.Len() + 8)
		const max = 1<<32 - 1
		compLen := (1 + length) ^ (max)

		header := make([]byte, 8)
		for i := uint(0); i < 4; i++ {
			header[i] = byte(length >> (i * 8))
			header[i+4] = byte(compLen >> (i * 8))
		}

		buffBytes := buff.Bytes()

		iter := bag.NewIterator()
		for ; iter.Current != nil; iter.Iter() {
			//println("con")
			_, err := (*(iter.Current.Conn)).Write(header)
			if err == nil {
				_, err = (*(iter.Current.Conn)).Write(buffBytes)
			}

			if err != nil {
				println(err.String())
				iter.Remove()
				iter.GoBack()
				println("removed a connection")
			}
		}
		//time.Sleep(1000)
	}
}

func SetupTCP() {
	println("Setting Up TCP")
	const connectedAndWaitingMax = 4
	conChan := make(chan *net.Conn, connectedAndWaitingMax)
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
