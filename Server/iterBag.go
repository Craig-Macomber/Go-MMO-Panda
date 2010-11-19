package main

import (
	"time"
	"fmt"
)

// similar to an unrolled linked list, but with better memory efficency
// and with faster but order disrupting removes.
// collection, remove disrupts order,
// iterable (bidirectionally, allows removing and adding while iterating)
// non indexable
// Non safe for concurrent use of Add or Remove
// use your own mutex if needed
// optimized for gorwing/shrinking as needed and good cache perforamnce
// constant time remove disrupts order,
// but Pop can be used to remove the last while maintaining order
// supports efficent multi core parallel iteration (untested!) without add and remove.

// potentially split-able (fast, but lienar time)
// and merge-able (constant time, but disrupts order) quickly,
// but not yet implemented

const arraySize=512
const clear=false // if Entry contains pointers, set clear to true to allow GC

type IterBag struct {
    start *node
    end *node
    nodeCount int
    fillIndex int // index of last item in end
}

type Entry int

type node struct {
	Previous *node
	Next *node
	Data [arraySize]Entry
}


func NewIterBag() *IterBag {
    n:=new(node)
	return &IterBag{start: n, end: n, nodeCount: 1, fillIndex: -1}
}

// adds element to end
func (s *IterBag) Add(x Entry) {
	if s.fillIndex>=arraySize-1 {
        n:=&node{Previous: s.end}
        n.Data[0]=x
	    s.end.Next=n
	    s.end=n
	    s.nodeCount++
	    s.fillIndex=0
	} else {
	    s.fillIndex++
	    s.end.Data[s.fillIndex]=x
	}
}

// overwrites the passed Entry with the last entry, then removes the last entry.
func (s *IterBag) remove(e *Entry) (last bool){
	last=e==&s.end.Data[s.fillIndex]
	*e=s.end.Data[s.fillIndex]
	
	if clear {
	    s.end.Data[s.fillIndex]=*new(Entry)
	}
	
	s.fillIndex--
	
	if s.fillIndex<0 && s.start!=s.end {
	    s.fillIndex=arraySize-1
	    s.end=s.end.Previous
	    s.end.Next=nil
	    s.nodeCount--
	}
	return
}

// remove and return the last element for stack like use
func (s *IterBag) Pop() (e Entry) {
	s.remove(&e) // explotes how remove works
	return
}


func (s *IterBag) Length() int{
    return (s.nodeCount-1)*arraySize+s.fillIndex+1
}

// safe to modify entries in returned blocks, but calling Add or Remove while
// processing will cause unexpected issues
func (s *IterBag) BlockChan(out chan<- []Entry) {
    for n:=s.start ; n.Next!=nil ; n=n.Next {
        out<-n.Data[:]
    }
    out<-s.end.Data[:s.fillIndex+1]
    close(out)
}


type Iterator struct{
    n *node
    Bag *IterBag
    index int // pointing to Next
    Current *Entry
    Next *Entry
}

func (s *IterBag) NewIterator() *Iterator{
    i:=Iterator{s.start,s,-1,nil,nil}
    i.Iter()
    i.Iter()
    return &i
}

// removes Current, and goes to and returns Next. Effects ordering (last moved to Next)
func (i *Iterator) Remove() (out *Entry){
    if i.Bag.remove(i.Current) {
        i.Current=nil
    } else {
        *i.Current, *i.Next = *i.Next, *i.Current
    }
    return i.Current
}

// adds an item to the end by just calling the set's Add
func (i *Iterator) Add(e Entry) {
    i.Bag.Add(e)
}


// goes back one
func (i *Iterator) GoBack() {
    i.Next=i.Current
    i.index--
    if i.index>0 {
        i.Current = &i.n.Data[i.index-1]
    } else if i.index<0 {
        i.n = i.n.Previous
        i.index=arraySize-1
    } else {
        if i.n.Previous!=nil {
            d:=i.n.Previous.Data
            i.Current = &d[arraySize-1]
        } else {
            i.Current = nil
        }
    }
}

// increments and returns Current
func (i *Iterator) Iter() (*Entry){
    i.Current=i.Next
    i.index++
    if i.index<arraySize && (i.n.Next!=nil || i.index<=i.Bag.fillIndex) {
        i.Next = &i.n.Data[i.index]
    } else {
        if i.n.Next==nil {
            i.Next=nil
        } else {
            i.n = i.n.Next
            i.index=0
            i.Next = &i.n.Data[0]
        }
    }
    return i.Current
}



func main() {
    s:=NewIterBag()
    t := time.Nanoseconds()

    fmt.Println("IterBag Test")
    
    const count=1000000
    s=NewIterBag()
    t = time.Nanoseconds()
    for i := 0; i < count; i++ {
        s.Add(Entry(i))
    }
    fmt.Println("Add Time:",float(time.Nanoseconds()-t)/count)
    
    
    iter:=s.NewIterator()
    t = time.Nanoseconds()
    for e:=iter.Current ; e!=nil ; e=iter.Iter() {
    }
    fmt.Println("Iter Time:",float(time.Nanoseconds()-t)/count)
    
    t = time.Nanoseconds()
    iter.GoBack()
    for ; iter.Current!=nil ; iter.GoBack() {
    }
    fmt.Println("Reverse Iter Time:",float(time.Nanoseconds()-t)/count)
    
    
    iter=s.NewIterator()
    t = time.Nanoseconds()
    e:=iter.Current
    for e!=nil {
        e=iter.Remove()
    }
    fmt.Println("Remove Time:",float(time.Nanoseconds()-t)/count)

    
    println("Slice Test")
    ss:=make([]Entry,0)
    t = time.Nanoseconds()
    for i := 0; i < count; i++ {
        ss=append(ss,Entry(i))
    }
    fmt.Println("Add Time:",float(time.Nanoseconds()-t)/count)
    
    ss=make([]Entry,count)
    t = time.Nanoseconds()
    for i := 0; i < count; i++ {
        ss[i]=Entry(i)
    }
    fmt.Println("Add Time:",float(time.Nanoseconds()-t)/count)
    
    println("Map Test")
    m:=map[Entry] int{}
    t = time.Nanoseconds()
    for i := 0; i < count; i++ {
        m[Entry(i)]=0
    }
    fmt.Println("Add Time:",float(time.Nanoseconds()-t)/count)
    
    
    println ("Last")
}