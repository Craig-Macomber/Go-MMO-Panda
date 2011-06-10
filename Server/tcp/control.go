package tcp

import (
	"time"
	"net"
	"os"
	"fmt"
	"crypto/tls"
)

// main loop and code to start main loop in this file

func updateLoop(in <-chan *LoggedIn) {
	bag := NewIterBag()
	const size = 100
	data := make([]byte, size)

	for i := 0; i < size; i++ {
		data[i] = uint8(i)
	}

	data = []byte("TEST_DATA")

	ramSync := NewRamSync(data, 1)
	eventNode := NewEventNode(2)
	children := make([]SyncNode, 2)
	children[0] = ramSync
	children[1] = eventNode
	root := newParentNode(children, 0)

	frameCount := 0
	for {

		// Add any new connections
		waitChan := time.After(500000000)
	L:
		for {
			select {
			case newCon := <-in:
				if newCon == nil {
					// in is closed!
					break L
				} else {
					bag.Add(*newCon)
					go broadcast(newCon, eventNode)
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
			ramSync.Data = append(ramSync.Data, 'x')
		}

		root.Write(bag)
	}
}

func SetupTCP(useTls bool, address string) {
	println("Setting Up TCP at:", address)
	const connectedAndWaitingMax = 0
	conChan := make(chan net.Conn, connectedAndWaitingMax)
	halt := make(chan int)

	var listener net.Listener
	var err os.Error
	if useTls {
		certs := make([]tls.Certificate, 1)
		c0, errx := tls.LoadX509KeyPair("cert/cert.pem", "cert/key.pem")
		certs[0] = c0
		fmt.Println(errx)
		config := tls.Config{Certificates: certs, ServerName: "TestServer"}
		listener, err = tls.Listen("tcp", ":6666", &config)
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