package main

import (
	"time"
)

// unordered collection
// Non safe for concurrent use of Add or Remove
// use your own mutex if needed

const arraySize=256
const clear=false // if Entry contains pointers, set clear to true to allow GC

type IterBag struct {
    start *node
    end *node
    nodeCount int
}

type Entry struct {
    Data int //SomeType
}

type node struct {
	Previous *node
	Next *node
	Data []Entry
}


func NewIterBag() *IterBag {
    n:=&node{nil,nil,make([]Entry,0,arraySize)}
	return &IterBag{n, n, 1}
}

func (s *IterBag) Add(x Entry) {
	if len(s.end.Data)==arraySize {
	    n:=&node{s.end,nil,make([]Entry,1,arraySize)}
	    n.Data[0]=x
	    s.end.Next=n
	    s.end=n
	    s.nodeCount++
	} else {
	    slice := s.end.Data[0:1+len(s.end.Data)]
	    slice[len(slice)-1]=x
	    s.end.Data = slice
	}
}

func (s *IterBag) remove(e *Entry) (last bool){
	old := s.end.Data
	endIndex := len(old)-1
	last=e==&old[endIndex]
	*e=old[endIndex]
	s.end.Data = old[:endIndex]
	if clear {
	    old[endIndex]=*new(Entry)
	}
	if len(s.end.Data)==0 && s.start!=s.end {
	    s.end=s.end.Previous
	    s.end.Next=nil
	    s.nodeCount--
	}
	return
}

func (s *IterBag) Length() int{
    return (s.nodeCount-1)*arraySize+len(s.end.Data)
}

// safe to modify entries in returned blocks, but calling Add or Remove while
// processing will cause unexpected issues
func (s *IterBag) BlockChan(out chan<- []Entry) {
    for n:=s.start ; n!=nil ; n=n.Next {
        out<-n.Data
    }
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
func (i *Iterator) Remove() *Entry{
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
    if i.index<0 {
        i.n = i.n.Previous
        i.index=len(i.n.Data)-1
    }
    if i.index>0 { // next is first in node
        i.Current = &i.n.Data[i.index-1]
    } else {
        if i.n.Previous!=nil {
            d:=i.n.Previous.Data
            i.Current = &d[len(d)-1]
        } else {
            i.Current = nil
        }
    }
}

// increments and returns Current
func (i *Iterator) Iter() (*Entry){
    i.Current=i.Next
    i.index++
    if i.index<len(i.n.Data) {
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
//     for i := 0; i < 10000000; i++ {
//         s.Add(Entry{i})
//     }
//     println(time.Nanoseconds()-t)
// 
//     
//     t = time.Nanoseconds()
//     
//     bChan:=make(chan IterBlock)
//     go s.BlockChan(bChan)
//     L: for b:=range bChan{
//         d:=b.AllData()
//         //println("Block")
//         for i:=0 ; i<b.Length() ; i++ {
//             if d[i].Data%7==0 || d[i].Data%5==0 {
//                 //println("Removed:",d[i].Data,i)
//                 if b.Remove(i) { 
//                     break L
//                 }
//                 i--
//             }
//         }
//     }
//     println(time.Nanoseconds()-t,s.Length())
    

    
    println("IterTest")
    
    s=NewIterBag()
    t = time.Nanoseconds()
    for i := 0; i < 10000000; i++ {
        s.Add(Entry{i})
    }
    println(time.Nanoseconds()-t)
    t = time.Nanoseconds()
    
    iter:=s.NewIterator()
    println(iter.Current,iter.Next)
    for e:=iter.Current ; e!=nil ; e=iter.Iter() {
        for e!=nil && (e.Data%7==0 || e.Data%5==0) {
            e=iter.Remove()
        }
    }    
    println(time.Nanoseconds()-t,s.Length())
//     iter=s.NewIterator()
//     for e:=iter.Current ; e!=nil ; e=iter.Iter() {
//         println(e.Data)
//     }    
    
    
    
//     s=NewIterBag()
//     t = time.Nanoseconds()
//     for i := 0; i < 10000000; i++ {
//         s.Add(Entry{i})
//     }
//     println(time.Nanoseconds()-t)
//     t = time.Nanoseconds()
//     
//     s.Feed (func(b IterBlock) {
//         d:=b.AllData()
//         for i:=0 ; i<b.Length() ; i++ {
//             if d[i].Data%7==0 || d[i].Data%5==0 {
//                 if b.Remove(i) { 
//                     return
//                 }
//                 i--
//             }
//         }
//     })
//     
//     println(time.Nanoseconds()-t)
//     
//     s=NewIterBag()
//     s.Add(Entry{1})
//     last:=1
//     s.Feed (func(b IterBlock) {
//         d:=b.AllData()
//         for i:=0 ; i<b.Length() ; i++ {
//             x:=d[i].Data
//             s.Add(Entry{last+x})
//             last=x
//             if last>100000 {
//                 return
//             }
//         }
//     })
//     println(time.Nanoseconds()-t)
//     
//     println("Remove 2s")
//     s=NewIterBag()
//     t = time.Nanoseconds()
//     for i := 0; i < 10000; i++ {
//         s.Add(Entry{i})
//     }
//     println(time.Nanoseconds()-t)
//     t = time.Nanoseconds()
//     
//     s.Feed (func(b IterBlock) {
//         d:=b.AllData()
//         for i:=0 ; i<b.Length() ; i++ {
//             x:=d[i].Data
//             if x%2==0 {
//                 d[i].Data=x+1
//                 //s.Add(Entry{x+1})
//                 s.Add(Entry{x-1})
// //                 if b.Remove(i) { 
// //                     return
// //                 }
// //                 i--
//             }
//         }
//     })
//     println(time.Nanoseconds()-t)
//     
//     
//     //bench ItemChan
//     println("ItemChan")
//     s=NewIterBag()
//     t = time.Nanoseconds()
//     for i := 0; i < 10000000; i++ {
//         s.Add(Entry{i})
//     }
//     println(time.Nanoseconds()-t)
//     t = time.Nanoseconds()
//     
//     iChan:=make(chan *Entry)
//     oChan:=make(chan *Entry)
//     go s.ItemChan(iChan,oChan)
//     for e:= range(oChan) {
//         if e.Data%7==0 || e.Data%5==0 {
//             iChan<-nil
//         }
//     }
//     println(time.Nanoseconds()-t)
    println ("Last")
}