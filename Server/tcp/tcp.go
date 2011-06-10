package tcp

import (
	"encoding/binary"
	"fmt"
)


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



func iterTrySendAll(bag *IterBag, data []byte) {
	for iter := bag.NewIterator(); iter.Current != nil; iter.Iter() {
		iterTrySend(iter, data)
	}
}

func iterTrySend(iter *Iterator, data []byte) {
	failed := iter.Current.SendMessage(data)
	if failed {
		iter.Remove()
		iter.GoBack()
		fmt.Println("Removing from Bag")
	}
}


