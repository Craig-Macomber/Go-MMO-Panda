package iterBag

import (
	"testing"
    "time"
	"fmt"
)


// This isn't really a test, I just stuck some code in here for now.
func TestAll(testingX *testing.T) {
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
    for ; iter.Current!=nil ; iter.Iter() {
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
    
    t = time.Nanoseconds()
    for _= range(ss) {
    }
    fmt.Println("Iter Time:",float(time.Nanoseconds()-t)/count)
    
    ss=make([]Entry,count)
    t = time.Nanoseconds()
    for i := 0; i < count; i++ {
        ss[i]=Entry(i)
    }
    fmt.Println("Assign Time:",float(time.Nanoseconds()-t)/count)
    
    println("Map Test")
    m:=map[Entry] int{}
    t = time.Nanoseconds()
    for i := 0; i < count; i++ {
        m[Entry(i)]=0
    }
    fmt.Println("Add Time:",float(time.Nanoseconds()-t)/count)
    
    t = time.Nanoseconds()
    for _,_= range(m) {
    }
    fmt.Println("Iter Time:",float(time.Nanoseconds()-t)/count)
    
    fmt.Println("Last")
}