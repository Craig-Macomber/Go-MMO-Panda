package main

import (
	"http"
	"tcp"
	"io"
	"io/ioutil"
	//"os"
)


func httpHandler(w http.ResponseWriter, r *http.Request) {
    serverList,_:=ioutil.ReadFile("serverList.txt")
    loc:=r.URL.Path[1:]
    if loc=="serverList"{
        io.WriteString(w, string(serverList))
    } else {
        io.WriteString(w, "hello, world!\n"+loc)
    }
}

func setupHttp() {
    println("hosting http")
	http.HandleFunc("/", httpHandler)
	http.ListenAndServe(":8080", nil)
}

func main() {
    halt:=make(chan int)
    tcp.SetupTCP()
    go setupHttp()
    
    _ = <-halt
    println("Over")
}