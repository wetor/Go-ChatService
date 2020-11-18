package main

import (
	"ChatService/src/api"
	"ChatService/src/chat"
	"fmt"
	"net/http"
	"time"
)

func test() {

	exit := make(chan bool)
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println(">>开始匹配0")
		chat.Match(0)
	}()
	go func() {
		time.Sleep(2 * time.Second)
		//chat.Cancel(3)
		time.Sleep(1 * time.Second)
		fmt.Println(">>开始匹配1")
		chat.Match(1)
	}()
	go func() {
		time.Sleep(4 * time.Second)
		fmt.Println(">>开始匹配2")
		chat.Match(2)
		exit <- true
	}()
	fmt.Println(">>开始匹配3")
	chat.Match(3)
	<-exit
	time.Sleep(1 * time.Second)
	fmt.Println(">>end")
}

func testws() {
	for key, value := range h {
		fmt.Println(key)
		go value.run()
	}

	server := http.Server{
		Addr: "127.0.0.1:8080",
	}
	http.HandleFunc("/ws/", myws)
	server.ListenAndServe()
}

func main() {
	fmt.Println("Main")
	testws()

	return
	server := http.Server{
		Addr: "127.0.0.1:8080",
	}
	http.HandleFunc("/chat/match", api.Match)
	http.HandleFunc("/chat/cancel", api.Cancel)
	http.HandleFunc("/chat", api.Chat)
	http.HandleFunc("/chat/report", api.Report)
	http.HandleFunc("/chat/webscoket/", api.Webscoket) // /chat/webscoket/{wsid:string}
	server.ListenAndServe()
}
