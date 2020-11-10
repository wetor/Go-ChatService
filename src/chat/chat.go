package chat

import (
	"ChatService/src/data"
	"encoding/json"
	"fmt"
)

// WaitPool 等待用户数组
var WaitPool map[int]data.WaitUser

// Room 连接信息数组
var Room map[int]data.Connection

// Chat 接口
func Chat() {
	fmt.Println("Chat Service")
	bytes, _ := json.Marshal(data.GetUser(1))
	fmt.Println(string(bytes))

}
