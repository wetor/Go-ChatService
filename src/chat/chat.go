package chat

import (
	"ChatService/src/data"
	"ChatService/src/rabbitmq"
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

// WaitPool 等待用户数组
var WaitPool map[int]*data.WaitUser

// Room 连接信息数组
var Room map[string]*data.Connection

// Rabbitmq 实例
var Rabbitmq *rabbitmq.RabbitMQ

func init() {
	fmt.Println("chat init")
	WaitPool = make(map[int]*data.WaitUser)
	Room = make(map[string]*data.Connection)
	Rabbitmq = rabbitmq.NewRabbitMQ("chat")
}

// Chat 接口
func Chat() {
	fmt.Println("Chat Service")
	bytes, _ := json.Marshal(data.GetUser(1))
	fmt.Println(string(bytes))

}

// sort 排序WaitPool
func sort() {

}

// getWsID 返回短uuid作为wsid
func getWsID() string {
	return "ad125sdf"
}

// updateTime 更新等待时间
func updateTime(id int) {
	WaitPool[id].Time++
}

// match 根据标签的匹配算法
func match(userA data.User, userB data.User) bool {
	for _, labelA := range userA.Label {
		for _, labelB := range userB.Label {
			if labelA == labelB {
				fmt.Printf("匹配成功：%s %s\n", labelA, labelB)
				return true
			}
		}
	}
	return false
}

// Match 匹配用户 id 自身id
func Match(id int) int {

	matchID := -2 // 匹配到的user_id，初始值-2
	user := data.GetUser(id)
	sort() // 按照等待时间排序
	for waitID, waitUser := range WaitPool {
		if match(waitUser.User, user) {
			fmt.Printf("%d 从WaitPool中匹配到 %d\n", id, waitID)
			matchID = waitID
			delete(WaitPool, waitID) // 从等待队列中删除
			jsonByte, _ := json.Marshal(data.MqUser{ID: waitID, MatchID: id, Type: 1})
			Rabbitmq.PublishPub(string(jsonByte)) // 发送广播，匹配到waitID
			break
		} else {
			updateTime(waitID) // 更新等待时间
		}
	}

	var recv data.MqUser
	if matchID == -2 { // 没有从WaitPool中匹配到，则进入等待匹配状态
		fmt.Printf("%d 等待他人匹配...\n", id)
		WaitPool[id] = &data.WaitUser{User: user, Time: 0} // 加入到等待队列

		wait := make(chan int)
		Rabbitmq.ReceiveSub(func(message <-chan amqp.Delivery) {
			//开启rabbitmq接收模式，如果被他人匹配到，就会执行
			for d := range message {
				json.Unmarshal(d.Body, &recv) // 类型转换
				if recv.ID == id {            // 被他人匹配到，或匹配被取消
					if recv.Type == 1 {
						wait <- recv.MatchID // 被他人匹配到
					} else if recv.Type == -1 {
						wait <- -1 // 匹配被取消
					}
					break
				}
			}

		})
		matchID = <-wait
		fmt.Printf("%d 等待结束，被匹配到 %d\n", id, matchID)
		if matchID == -1 {
			return -1 //返回602取消响应
		}
	}

	wsid := getWsID()
	Room[wsid] = &data.Connection{
		//Ws:    null,
		WsID:  wsid,
		Users: [2]data.User{user, data.GetUser(matchID)},
		Sc:    make(chan []byte),
		//Data:  null,
	}
	return matchID //返回200响应
}
