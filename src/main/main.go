package main

import (
	"ChatService/src/chat"
	"ChatService/src/rabbitmq"
	"fmt"
)

func main() {
	fmt.Println("Main")
	chat.Chat()
	rabbitmq.Test()
}
