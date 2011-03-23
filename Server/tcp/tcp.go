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
	"sync"
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


type Player struct {
    Name string
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
	fmt.Println("Disconnecting", c.Conn.RemoteAddr(),"Error:", reason, err)
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
				return nil, true
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

func login(c *Connected, out chan<- *LoggedIn) {
	name, fail := c.ReadMessage()
	if !fail{
		fmt.Println("Got Name:",name)
		password, fail := c.ReadMessage()
		if !fail {
			fmt.Println("Login:", name[0], password[0], string(name[1:]), string(password[1:]))
			loginSucess:=[...]byte{0,0}[:]
			c.SendMessage(loginSucess)
			
			out <- &LoggedIn{*c,&Player{string(name),string(password)}}
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




type SyncNode interface {
	Write(b *IterBag)
}

type ParentNode struct {
	children []SyncNode
	headFlag byte
}


func newParentNode(children []SyncNode, headFlag byte) *ParentNode{
	return &ParentNode{children,headFlag}
}

func (n *ParentNode) Write(b *IterBag){
	for _,child:=range(n.children){
		child.Write(b)
	}
}


type RamSync struct {
	Data []byte
	oldData []byte
	framesSinceKeyframe int
	headFlag byte
}

func NewRamSync(data []byte,headFlag byte) *RamSync{
	return &RamSync{data,nil,0,headFlag}
}

func (n *RamSync) Write(b *IterBag){
	n.framesSinceKeyframe++
	size:=len(n.Data)
	oldValid:=true
	if n.oldData==nil || len(n.oldData)!=size{
		n.oldData=make([]byte,size)
		oldValid=false
	}
	
	buff := new(bytes.Buffer)
	mtype := uint8(2)
	toSend := n.Data
	
	if !oldValid || n.framesSinceKeyframe%16 == 0 {
		mtype = 1
	} else {
		xOrData:=make([]byte,size)
		for i := 0; i < size; i++ {
			xOrData[i] = n.Data[i] ^ n.oldData[i]
		}
		toSend=xOrData
	}
	for i := 0; i < size; i++ {
		n.oldData[i] = n.Data[i]
	}
	buff.WriteByte(n.headFlag)
	buff.WriteByte(mtype)
	
	zlibw, _ := zlib.NewWriter(buff)
	zlibw.Write(toSend)
	zlibw.Close()

	buffBytes := buff.Bytes()
	
	iterTrySendAll(b,buffBytes)
}




type Event struct {
	Data []byte
}


type EventNode struct {
	sync sync.Mutex
	events []*Event
	headFlag byte
}

func NewEventNode(headFlag byte) *EventNode{
	return &EventNode{events:make([]*Event,0),headFlag:headFlag}
}

func iterTrySendAll(bag *IterBag,data []byte) {
	for iter := bag.NewIterator(); iter.Current != nil; iter.Iter() {
		iterTrySend(iter,data)
	}
}

func iterTrySend(iter *Iterator,data []byte) {
	failed := iter.Current.SendMessage(data)
	if failed {
		iter.Remove()
		iter.GoBack()
		fmt.Println("Removing from Bag")
	}
}

func (e *EventNode) Write(b *IterBag){
	e.sync.Lock()
	oldEvents:=e.events
	e.events=make([]*Event,0)
	e.sync.Unlock()
	for _,event:=range(oldEvents){
		data:=make([]byte,len(event.Data)+1)
		data[0]=e.headFlag
		copy(data[1:],event.Data)
		iterTrySendAll(b,data)
	}
	
}

func (e *EventNode) Add(event *Event){
	e.sync.Lock()
	e.events=append(e.events,event)
	e.sync.Unlock()
}



func broadcast(source *LoggedIn, dst *EventNode){
    for{
        data,failed:=source.ReadMessage()
        if failed{
            return
        } else {
            dst.Add(&Event{append([]byte(source.Name+": "),data...)})
        }
    }
}

func updateLoop(in <-chan *LoggedIn) {
	bag := NewIterBag()
	const size = 100
	data := make([]byte, size)

	for i := 0; i < size; i++ {
		data[i] = uint8(i)
	}
	
	data=[]byte("TEST_DATA")
	
	
	ramSync:=NewRamSync(data,1)
	eventNode:=NewEventNode(2)
	children:=make([]SyncNode,2)
	children[0]=ramSync
	children[1]=eventNode
	root:=newParentNode(children,0)
	
	
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
					go broadcast(newCon,eventNode)
				}
			case _ = <-waitChan:
				break L
			}
		}

		frameCount++
		if frameCount%1 == 0 {
			println(frameCount, " - ", bag.Length())
		}
		if frameCount%20 == 0 {
			ramSync.Data=append(ramSync.Data,'x')
		}
		
		root.Write(bag)
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

	conChan2 := make(chan *LoggedIn, connectedAndWaitingMax)

	go welcomTestLoop(conChan, conChan2)
	go updateLoop(conChan2)

	println("TCP Setup")

}
