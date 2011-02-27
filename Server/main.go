package main

import (
	"http"
	"tcp"
	"io"
	"io/ioutil"
	"os"
	"bufio"
	"bytes"
)

var httpHandlers map[string]*http.ServeMux

func init() {
	httpHandlers = make(map[string]*http.ServeMux)
}


func startServers(serverType string, serverList io.Reader, servers map[string]func([][]byte)) {
	// read serverList, and start servers that are assigned to this serverType
	bType := []byte(serverType)
	r := bufio.NewReader(serverList)
	sig, err := r.ReadString('\n')
	println(sig)
	var line string
	for err == nil {
		line, err = r.ReadString('\n')
		if len(line) > 1 {
			//read columns
			cols := bytes.Split([]byte(line), []byte(" "), 6)
			if len(cols) != 6 {
				println("bad line in serverList:", line)
			} else {
				if bytes.Compare(bType, cols[0]) == 0 {
					// current line assigned to this server type, so start it
					name := string(cols[1])
					data := cols[2:]
					servers[name](data)
				}
			}
		}
	}
}


func LaunchLoginServer(data [][]byte) {
	tcp.SetupTCP()
}

func httpLauncher(addr string, pattern string, handler func(http.ResponseWriter, *http.Request)) func([][]byte) {
	return func(data [][]byte) {
		mux, ok := httpHandlers[addr]
		if !ok {
			mux = http.NewServeMux()
			go http.ListenAndServe(addr, mux)
			httpHandlers[addr] = mux
			println("adding http server for", addr)
		}
		mux.HandleFunc(pattern, handler)
	}
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	serverList, _ := ioutil.ReadFile("serverList.txt")
	io.WriteString(w, string(serverList))
}

func statusHttpHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "running")
}


func main() {
	halt := make(chan int)
	serverList, _ := os.Open("serverList.txt", os.O_RDONLY, 0)
	servers := make(map[string]func([][]byte))
	servers["Login"] = LaunchLoginServer
	servers["ServerList"] = httpLauncher(":8080", "/ServerList", httpHandler)
	servers["Status"] = httpLauncher(":8080", "/Status", statusHttpHandler)
	startServers("master", serverList, servers)
	_ = <-halt
	println("Over")
}
