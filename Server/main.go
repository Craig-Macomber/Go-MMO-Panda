package main

import (
	//"runtime"
	//"time"
	"net"
	"http"
	"io"
	"io/ioutil"
)

func con(listener net.Listener) {
    con,err:=listener.Accept()
    println(err)
    con.Write([]byte("\x00TestMessage"))
    con.Write([]byte("\x00Message2!"))
    //println(con.Read(20))
}


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
	http.HandleFunc("/", httpHandler)
	http.ListenAndServe(":8080", nil)
}

func main() {
    listener,err:=net.Listen("tcp","127.0.0.1:6666")
    println(err)
    go con(listener)
    
    setupHttp()
    
    println("Over")
}