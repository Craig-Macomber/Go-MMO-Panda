package tcp

import (
	"bytes"
	"compress/zlib"
	"sync"
)

// stuff related to "nodes" which represent a tree of
// information clients can sync/watch/download depending on type

type SyncNode interface {
	Write(b *IterBag)
}

type ParentNode struct {
	children []SyncNode
	headFlag byte
}


func newParentNode(children []SyncNode, headFlag byte) *ParentNode {
	return &ParentNode{children, headFlag}
}

func (n *ParentNode) Write(b *IterBag) {
	for _, child := range n.children {
		child.Write(b)
	}
}


type RamSync struct {
	Data                []byte
	oldData             []byte
	framesSinceKeyframe int
	headFlag            byte
}

func NewRamSync(data []byte, headFlag byte) *RamSync {
	return &RamSync{data, nil, 0, headFlag}
}

func (n *RamSync) Write(b *IterBag) {
	n.framesSinceKeyframe++
	size := len(n.Data)
	oldValid := true
	if n.oldData == nil || len(n.oldData) != size {
		n.oldData = make([]byte, size)
		oldValid = false
	}

	buff := new(bytes.Buffer)
	mtype := uint8(2)
	toSend := n.Data

	if !oldValid || n.framesSinceKeyframe%16 == 0 {
		mtype = 1
	} else {
		xOrData := make([]byte, size)
		for i := 0; i < size; i++ {
			xOrData[i] = n.Data[i] ^ n.oldData[i]
		}
		toSend = xOrData
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

	iterTrySendAll(b, buffBytes)
}


type Event struct {
	Data []byte
}


type EventNode struct {
	sync     sync.Mutex
	events   []*Event
	headFlag byte
}

func NewEventNode(headFlag byte) *EventNode {
	return &EventNode{events: make([]*Event, 0), headFlag: headFlag}
}


func (e *EventNode) Write(b *IterBag) {
	e.sync.Lock()
	oldEvents := e.events
	e.events = make([]*Event, 0)
	e.sync.Unlock()
	for _, event := range oldEvents {
		data := make([]byte, len(event.Data)+1)
		data[0] = e.headFlag
		copy(data[1:], event.Data)
		iterTrySendAll(b, data)
	}

}

func (e *EventNode) Add(event *Event) {
	e.sync.Lock()
	e.events = append(e.events, event)
	e.sync.Unlock()
}


func broadcast(source *LoggedIn, dst *EventNode) {
	for {
		data, failed := source.ReadMessage()
		if failed {
			return
		} else {
			dst.Add(&Event{append([]byte(source.Name+": "), data...)})
		}
	}
}
