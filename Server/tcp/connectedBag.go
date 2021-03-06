package tcp



//Generic Container via gotgo
//Summery of built in operations:
//Add: O(1) - Adds to end, can be used while iterating, or otherwise
//Remove: O(1) - Moves last item over current, use while iterating
//Pop: O(1) - Returns last item
//Length O(1) - Number of items
//BlockChan O(n) - Feeds slices of data to a channel for concurrent processing and updating
//    remove is not possible while iterating this way. Adding is, but added
//    items may or may not show up in the current pass

//Concurrency:
// Adding while iterating is not well tested,
// but should work except for possible causing the last block to not get in that pass

// similar to an unrolled linked list, but with better memory efficency
// and with faster but order disrupting removes.
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
// shifting blocks from one bag to another for quick balancing should be easy to impliment as well

const arraySize=128 // Number of Entries ber block
const clear=false // if Entry contains pointers, set clear to true to allow GC

type IterBag struct {
    start *node
    end *node
    nodeCount int
    fillIndex int // index of last item in end
}

type node struct {
	Previous *node
	Next *node
	Data [arraySize]LoggedIn
}

func NewIterBag() *IterBag {
    n:=new(node)
	return &IterBag{start: n, end: n, nodeCount: 1, fillIndex: -1}
}

// adds element to end
func (s *IterBag) Add(x LoggedIn) {
	if s.fillIndex>=arraySize-1 {
        n:=&node{Previous: s.end}
        n.Data[0]=x
        s.fillIndex=0
	    s.end.Next=n
	    s.end=n
	    s.nodeCount++
	} else {
	    s.end.Data[s.fillIndex+1]=x
	    s.fillIndex++
	}
}

// overwrites the passed Entry with the last entry, then removes the last entry.
func (s *IterBag) remove(e *LoggedIn) (last bool){
	last=e==&s.end.Data[s.fillIndex]
	*e=s.end.Data[s.fillIndex]
	
	if clear {
	    s.end.Data[s.fillIndex]=*new(LoggedIn)
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
func (s *IterBag) Pop() (e LoggedIn) {
	s.remove(&e) // explotes how remove works
	return
}


func (s *IterBag) Length() int{
    return (s.nodeCount-1)*arraySize+s.fillIndex+1
}

// safe to modify entries in returned blocks, but calling Add or Remove while
// processing will cause unexpected issues
func (s *IterBag) BlockChan(out chan<- []LoggedIn) {
    for n:=s.start ; n.Next!=nil ; n=n.Next {
        out<-n.Data[:]
    }
    out<-s.end.Data[:s.fillIndex+1]
    close(out)
}


type Iterator struct{
    n *node
    Bag *IterBag
    index int // &n.Data[index]==Current
    Current *LoggedIn
}

func (s *IterBag) NewIterator() *Iterator{
    i:=Iterator{s.start,s,-1,nil}
    i.Iter()
    return &i
}

// removes Current, and goes to and returns Next. Effects ordering (last moved to Next)
func (i *Iterator) Remove() (out *LoggedIn){
    if i.Bag.remove(i.Current) {
        i.Current=nil
    }
    return i.Current
}

// adds an item to the end by just calling the set's Add
func (i *Iterator) Add(e LoggedIn) {
    i.Bag.Add(e)
}


// goes back one
func (i *Iterator) GoBack() {
    i.index--
    if i.index>=0 {
        i.Current = &i.n.Data[i.index]
    } else {
        if i.n.Previous==nil {
            i.Current = nil
        } else {
            i.n = i.n.Previous
            i.index=arraySize-1
            i.Current = &i.n.Data[arraySize-1]
        }
    }
}

// increments current
func (i *Iterator) Iter(){
    i.index++
    if i.index<arraySize && (i.n.Next!=nil || i.index<=i.Bag.fillIndex) {
        i.Current = &i.n.Data[i.index]
    } else {
        if i.n.Next==nil {
            i.Current=nil
        } else {
            i.n = i.n.Next
            i.index=0
            i.Current = &i.n.Data[0]
        }
    }
}
// Here we will test that the types parameters are ok...
func testTypes(arg0 LoggedIn) {
    f := func(interface{}) { } // this func does nothing...
    f(arg0)
}
