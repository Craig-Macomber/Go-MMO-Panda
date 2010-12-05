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

interface Streamer {
    
}

type StreamSource struct {
    writers IterBag = NewIterBag()
    
}

type TcpRoot struct {
    
}

func NewRoot(halt <-chan int) {
    println("Setting Up TCP")
    const connectedAndWaitingMax=4
    
    conChan:=make(chan *net.Conn, connectedAndWaitingMax)

    listener,err:=net.Listen("tcp","127.0.0.1:6666")
    if err!=nil {
        println(err)
    }

    go getConnections(listener,conChan,halt)
    
    connectedChan :=make(chan *Connected)
    
    go connector(conChan,connectedChan)
    
    bag:=NewIterBag()
    go welcomTestLoop(connectedChan,bag)
    go updateLoop(bag,halt)
    
    println("TCP Setup")

}