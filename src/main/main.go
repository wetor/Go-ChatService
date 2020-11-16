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

func main() {
	fmt.Println("Main")

	server := http.Server{
		Addr: "127.0.0.1:8080",
	}
	http.HandleFunc("/match", api.ChatMatch)
	server.ListenAndServe()
}
