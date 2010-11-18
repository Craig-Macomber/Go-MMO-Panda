package main

import (
	//"runtime"
	"time"
)

// unordered collection
// Non safe for concurrent use of Add or Remove

const arraySize=256
const clear=false // if Entry contains pointers, set clear to true to allow GC

type IterSet struct {
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


func NewIterSet() *IterSet {
    n:=&node{nil,nil,make([]Entry,0,arraySize)}
	return &IterSet{n, n, 1}
}

func (s *IterSet) Add(x Entry) {
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

func (s *IterSet) remove(n *node, index int) (last bool){
	old := s.end.Data
	endIndex := len(old)-1
	last=n==s.end && index==endIndex
	n.Data[index]=old[endIndex]
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

func (s *IterSet) Length() int{
    return (s.nodeCount-1)*arraySize+len(s.end.Data)
}



type IterBlock struct{
    Set *IterSet
    n   *node
}

func (b *IterBlock) Remove(index int) (last bool){
    return b.Set.remove(b.n,index)
}

func (b *IterBlock) Data() ([]Entry){
    return b.n.Data
}

func (b *IterBlock) AllData() ([]Entry){
    return b.n.Data[:cap(b.n.Data)]
}

// Use this as an upper bound along with all data if you might call Add on the set while iterating
func (b *IterBlock) Length() int{
    return len(b.n.Data)
}

func (s *IterSet) Feed(f func(block IterBlock)) {
    for n:=s.start ; n!=nil ; n=n.Next {
        f(IterBlock{s,n})
    }
}

func (s *IterSet) BlockChan(out chan<- IterBlock) {
    for n:=s.start ; n!=nil ; n=n.Next {
        out<-IterBlock{s,n}
    }
    close(out)
}

func (s *IterSet) GoFeed(count int, f func(in chan []Entry)) {
    c:=make(chan []Entry)
    for i := 0; i < count; i++ {
        go f(c)
    }
    for n:=s.start ; n!=nil ; n=n.Next {
        c <- n.Data
    }
    close(c)
}

func main() {
//     count:=9
//     c:=make(chan int)
//     done:=make(chan int)
//     f:=func(in chan int){
//         i:=0
//         for x:= range in {
//             println(x,i)
//             i++
//         }
//         done<-0
//     }
//     for i := 0; i < count; i++ {
//         go f(c)
//     }
//     for k := 0; k < 10; k++ {
//         c <- k
//     }
//     close(c)
//     for i := 0; i < count; i++ {
//         <-done
//     }
//     println()
    
    s:=NewIterSet()
    t := time.Nanoseconds()
    for i := 0; i < 10000000; i++ {
        s.Add(Entry{i})
    }
    println(time.Nanoseconds()-t)
    
    t = time.Nanoseconds()
    
    bChan:=make(chan IterBlock)
    go s.BlockChan(bChan)
    for b:=range bChan{
        d:=b.AllData()
        //println("Block")
        for i:=0 ; i<b.Length() ; i++ {
            if d[i].Data%7==0 || d[i].Data%5==0 {
                //println("Removed:",d[i].Data,i)
                if b.Remove(i) { 
                    break
                }
                i--
            }
        }
    }
    println(time.Nanoseconds()-t)
    println ("Last")
}