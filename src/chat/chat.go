package chat

import (
	"ChatService/src/data"
	"ChatService/src/rabbitmq"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sony/sonyflake"
	"github.com/streadway/amqp"
)

// Connection 连接信息，房间
type Connection struct {
	Ws    *websocket.Conn
	WsID  string
	Users [2]data.User
	Sc    chan []byte // 用于储存发送数据的chan
	Data  *data.Data  // 当前正要发送的数据
}

// WaitPool 等待用户数组
var WaitPool map[int]*data.WaitUser

// Room 连接信息数组
var Room map[string]*Connection

// Rabbitmq 实例
var Rabbitmq *rabbitmq.RabbitMQ

// wsid生成器
var flake *sonyflake.Sonyflake

// wsidMap key:user_id value:wsid
var wsidMap map[int]string

// Wu WebScoket.Upgrader
var Wu *websocket.Upgrader

// lock 同步锁，防止同时操作上面变量
var lock sync.Mutex

func init() {
	fmt.Println("chat init")
	WaitPool = make(map[int]*data.WaitUser)
	Room = make(map[string]*Connection)
	Rabbitmq = rabbitmq.NewRabbitMQ("chat")
	flake = sonyflake.NewSonyflake(sonyflake.Settings{})
	wsidMap = make(map[int]string)

	Wu = &websocket.Upgrader{ReadBufferSize: 512, WriteBufferSize: 512, CheckOrigin: func(r *http.Request) bool { return true }}
}

// sort 排序WaitPool
func sort() {

}

// getWsid 通过双方的uid取得唯一的wsid
func getWsid(idA int, idB int) string {
	id, _ := flake.NextID()
	wsid := fmt.Sprintf("%x", id)
	wsidMap[idA] = wsid
	wsidMap[idB] = wsid
	return wsid
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
func Match(id int) (int, string) {

	//matchID := -2 // 匹配到的user_id，初始值-2
	user := data.GetUser(id)
	sort()      // 按照等待时间排序 TODO:未实现
	lock.Lock() // 在等待队列中匹配时上锁
	for waitID, waitUser := range WaitPool {
		if match(waitUser.User, user) {
			fmt.Printf("%d 从WaitPool中匹配到 %d\n", id, waitID)

			delete(WaitPool, waitID) // 从等待队列中删除
			// 创建连接（房间）
			wsid := getWsid(id, waitID)
			Room[wsid] = &Connection{
				//Ws:    null,
				WsID:  wsid,
				Users: [2]data.User{user, data.GetUser(waitID)},
				Sc:    make(chan []byte),
				//Data:  null,
			}
			lock.Unlock()
			// 将匹配者、被匹配者、wsid和消息类型发送广播
			jsonByte, _ := json.Marshal(data.MqUser{ID: waitID, MatchID: id, WsID: wsid, Type: 1})
			Rabbitmq.PublishPub(string(jsonByte)) // 发送广播，匹配到waitID

			return waitID, wsid
		}
		updateTime(waitID) // 更新等待时间
	}
	lock.Unlock()

	var recv data.MqUser
	// 没有从WaitPool中匹配到，则进入等待匹配状态
	fmt.Printf("%d 等待他人匹配...\n", id)

	go func() { // 超时计时
		time.Sleep(time.Second * 10) // 10s
		jsonByte, _ := json.Marshal(data.MqUser{ID: id, Type: 0})
		Rabbitmq.PublishPub(string(jsonByte)) // 发送广播，id 匹配超时
	}()
	lock.Lock()                                        //上锁
	WaitPool[id] = &data.WaitUser{User: user, Time: 0} // 加入到等待队列
	lock.Unlock()
	var wsID string
	wait := make(chan int)
	Rabbitmq.ReceiveSub(func(message <-chan amqp.Delivery) {
		//开启rabbitmq接收模式，如果被他人匹配到，就会执行
		for d := range message {
			json.Unmarshal(d.Body, &recv) // 类型转换
			if recv.ID == id {            // 被他人匹配到，或匹配被取消
				if recv.Type == 1 {
					lock.Lock()          // 上锁
					delete(WaitPool, id) // 从等待队列中删除
					lock.Unlock()
					wsID = recv.WsID
					wait <- recv.MatchID // 被他人匹配到
				} else if recv.Type == -1 {
					wait <- -1 // 匹配被取消
				} else if recv.Type == 0 {
					wait <- 0 // 匹配超时
				}
				break
			}
		}
	})
	matchID := <-wait

	if matchID == -1 {
		fmt.Printf("%d 等待结束，匹配被取消\n", id)
		return -2, "cancel" //返回602取消响应
	} else if matchID == 0 {
		fmt.Printf("%d 等待结束，匹配超时\n", id)
		return -1, "timeout" //返回601超时响应
	}
	fmt.Printf("%d 等待结束，被匹配到 %d，wsid %s\n", id, matchID, wsID)

	return matchID, wsID //返回200响应
}

// Cancel 取消匹配
func Cancel(id int) int {
	lock.Lock() // 此方法上锁
	defer lock.Unlock()
	// id 在等待队列WaitPool中
	for waitID := range WaitPool {
		if waitID == id {
			delete(WaitPool, waitID) // 从等待队列中删除
			jsonByte, _ := json.Marshal(data.MqUser{ID: waitID, Type: -1})
			Rabbitmq.PublishPub(string(jsonByte)) // 发送广播，取消匹配waitID
			return 0
		}
	}

	// 在Room中查找，若其中一人取消，则删除整个房间
	wsid := ""
	for wsID, conn := range Room {
		for _, user := range conn.Users {
			if user.ID == id {
				wsid = wsID
				break
			}
		}
	}
	if wsid != "" {
		delete(wsidMap, Room[wsid].Users[0].ID)
		delete(wsidMap, Room[wsid].Users[1].ID)
		delete(Room, wsid) // 删除房间
		return 1
	}
	return -1
}

// Chat 开始聊天
func Chat(id int, w http.ResponseWriter, r *http.Request) (int, string) {
	lock.Lock() // 此方法上锁
	defer lock.Unlock()
	wsid, ok := wsidMap[id]
	if !ok {
		// 房间不存在
		return -1, ""
	}
	_, ok = Room[wsid]
	if !ok {
		// 房间不存在
		return -1, ""
	}
	if Room[wsid].Ws != nil { // 一方已经开始聊天，ws连接已创建，无需创建
		fmt.Printf("%d WebScoket连接已经创建 %s\n", id, wsid)
		return 0, wsid
	}

	ws, err := Wu.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	Room[wsid].Ws = ws
	Room[wsid].Data = &data.Data{}
	Room[wsid].Sc = make(chan []byte, 256)
	fmt.Printf("%d WebScoket连接创建 %s\n", id, wsid)
	return 1, wsid

}

func (c *Connection) writer() {
	c.Ws.Close()
}

func (c *Connection) reader() {
	c.Ws.Close()
}
