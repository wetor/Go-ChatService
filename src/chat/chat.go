package chat

import (
	"ChatService/src/data"
	"ChatService/src/rabbitmq"
	"ChatService/src/redis"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sony/sonyflake"
	"github.com/streadway/amqp"
)

// Connection 连接信息
type Connection struct {
	//ChatID int
	Ws *websocket.Conn
	//WsID   string
	//Users  [2]data.User
	Sc   chan []byte // 用于储存发送数据的chan
	Data *WsData     // 当前正要发送的数据
}

// --------对象---------

// Rabbitmq 实例
var Rabbitmq *rabbitmq.RabbitMQ

// Redis 实例
var Redis *redis.RedisPool

// wsid生成器
var flake *sonyflake.Sonyflake

// Wu WebScoket.Upgrader
var wu *websocket.Upgrader

// --------储存---------

// waitPool 等待用户数组【可存redis】
var waitPool map[int]*data.WaitUser

// waitPool的key排序，根据等待时间由大到小排序，增加新用户是更新。[0]userid  [1]time【可存redis】
var waitPoolSort [][]int

// Rooms 连接信息数组【可存redis】
var Rooms map[string]*Room

// wsidMap key:user_id value:wsid【可存redis】
var wsidMap map[int]string

// uidMap key:wsid value:[]userid【可存redis】
var uidMap map[string][]int

// chatID【可存redis】
var chatID int

// --------锁---------

// lock 同步锁，防止同时操作上面变量
var lock sync.Mutex

func init() {
	fmt.Println("chat init")
	Rabbitmq = rabbitmq.NewRabbitMQ("chat")
	Redis = redis.NewRedisPool()
	flake = sonyflake.NewSonyflake(sonyflake.Settings{})
	wu = &websocket.Upgrader{ReadBufferSize: 512, WriteBufferSize: 512, CheckOrigin: func(r *http.Request) bool { return true }}

	waitPool = make(map[int]*data.WaitUser)
	//waitPoolSort = make([][]int, 0)
	Rooms = make(map[string]*Room)
	wsidMap = make(map[int]string)
	uidMap = make(map[string][]int)
	chatID = 0

}

// getChatid 获取数据库唯一id chatid
func getChatid() int {
	chatID++
	return chatID
}

// getWsid 取得唯一的wsid，并使用双方id建立映射表
func getWsid(idA int, idB int) string {
	id, _ := flake.NextID()
	wsid := fmt.Sprintf("%x", id)
	wsidMap[idA] = wsid
	wsidMap[idB] = wsid
	uidMap[wsid] = []int{idA, idB}
	return wsid
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

// sort 排序WaitPool
func sortPool() [][]int {
	waitPoolSort = make([][]int, 0)
	for waitID, waitUser := range waitPool {
		temp := []int{waitID, waitUser.Time}
		waitPoolSort = append(waitPoolSort, temp)
	}
	sort.Slice(waitPoolSort, func(i, j int) bool {
		return waitPoolSort[i][1] > waitPoolSort[j][1]
	}) // 按时间由大到小排序
	//fmt.Println(waitPoolSort)
	return waitPoolSort
}

// Match 匹配用户 id 自身id
func Match(id int) (int, string) {
	//matchID := -2 // 匹配到的user_id，初始值-2
	user := data.GetUser(id)

	lock.Lock() // 在等待队列中匹配时上锁
	for _, val := range waitPoolSort {
		waitID := val[0]
		waitUser := waitPool[waitID]
		//for waitID, waitUser := range waitPool {
		if match(waitUser.User, user) {
			fmt.Printf("%d 从WaitPool中匹配到 %d\n", id, waitID)

			delete(waitPool, waitID) // 从等待队列中删除
			sortPool()               // 排序等待池
			// 创建房间，暂时不初始化chan，因为可能会取消
			wsid := getWsid(id, waitID)
			Rooms[wsid] = &Room{
				ChatID: -1, //未初始化的房间为-1
				WsID:   wsid,
				Users:  []data.User{user, data.GetUser(waitID)},
			}
			lock.Unlock()
			// 将匹配者、被匹配者、wsid和消息类型发送广播
			jsonByte, _ := json.Marshal(data.MqUser{ID: waitID, MatchID: id, WsID: wsid, Type: 1})
			Rabbitmq.PublishPub(string(jsonByte)) // 发送广播，匹配到waitID

			return waitID, wsid
		}
		waitPool[waitID].Time++ // 更新等待时间
	}
	lock.Unlock()

	var recv data.MqUser
	// 没有从WaitPool中匹配到，则进入等待匹配状态
	fmt.Printf("%d 等待他人匹配...\n", id)

	go func() { // 超时计时
		time.Sleep(time.Second * 10) // 10s
		jsonByte, _ := json.Marshal(data.MqUser{ID: id, Type: -2})
		Rabbitmq.PublishPub(string(jsonByte)) // 发送广播，id 匹配超时
	}()
	lock.Lock() //上锁
	_, ok := waitPool[id]
	if ok {
		waitPool[id].Time++ // 更新等待时间
	} else {
		waitPool[id] = &data.WaitUser{User: user, Time: 0} // 加入到等待队列
	}
	sortPool() // 排序等待池
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
					delete(waitPool, id) // 从等待队列中删除
					lock.Unlock()
					wsID = recv.WsID
					wait <- recv.MatchID // 被他人匹配到
				} else if recv.Type == -1 {
					wait <- -1 // 匹配被取消
				} else if recv.Type == -2 {
					wait <- -2 // 匹配超时
				}
				break
			}
		}
	})
	matchID := <-wait

	if matchID == -1 {
		fmt.Printf("%d 等待结束，匹配被取消\n", id)
		return -2, "cancel" //返回602取消响应
	} else if matchID == -2 {
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
	for waitID := range waitPool {
		if waitID == id {
			delete(waitPool, waitID) // 从等待队列中删除
			jsonByte, _ := json.Marshal(data.MqUser{ID: waitID, Type: -1})
			Rabbitmq.PublishPub(string(jsonByte)) // 发送广播，取消匹配waitID
			return 0                              // 匹配取消
		}
	}

	// 在Room中查找，若其中一人取消，则删除整个房间
	wsid, ok := wsidMap[id]
	if ok && roomExist(wsid) {
		delete(wsidMap, uidMap[wsid][0])
		delete(wsidMap, uidMap[wsid][1])
		delete(uidMap, wsid)
		delete(Rooms, wsid) // 删除房间
		return 1            // 房间删除
	}
	return -1
}

// Chat 开始聊天
func Chat(id int) (int, string) {
	lock.Lock()         // 此方法上锁
	defer lock.Unlock() //解锁
	wsid, ok := wsidMap[id]
	if !ok || (ok && !roomExist(wsid)) {
		// 房间不存在
		return -1, ""
	}
	if Rooms[wsid].ChatID >= 0 { // 一方已经开始聊天，ws连接已创建，无需创建
		fmt.Printf("%d WebScoket连接已经创建 %s\n", id, wsid)
		return Rooms[wsid].ChatID, wsid
	}

	// 初始化chan
	Rooms[wsid].ChatID = getChatid()
	Rooms[wsid].Conns = make([]*Connection, 0)
	Rooms[wsid].Conn = make(chan *Connection)
	Rooms[wsid].Stream = make(chan []byte, 256)
	go Rooms[wsid].Run()
	fmt.Printf("%d WebScoket连接创建 %s\n", id, wsid)
	return Rooms[wsid].ChatID, wsid
}

// Webscoket webscoket接口实现
func Webscoket(wsid string, w http.ResponseWriter, r *http.Request) {
	defer fmt.Println("关闭连接 Webscoket")
	if !roomExist(wsid) {
		// 房间不存在
		fmt.Println("房间不存在，可能对方已取消")
		return
	}

	ws, err := wu.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("ws创建失败 " + wsid)
		return
	}
	conn := &Connection{
		Ws:   ws,
		Sc:   make(chan []byte, 256),
		Data: &WsData{},
	}
	conn.Data.Type = "start"
	conn.Data.Time = time.Now().Unix()
	Rooms[wsid].Conn <- conn
	go conn.writer(wsid)
	conn.reader(wsid)

	close := make(chan bool)
	go func() {
		for {
			val, ok := Rooms[wsid]
			if !ok || (ok && val.ChatID == -3) { // 判断房间是否已经被删除
				close <- true
				return
			}
		}
	}()
	<-close
	// 退出
	if roomExist(wsid) { // 双方中的一方删除房间
		delete(wsidMap, uidMap[wsid][0])
		delete(wsidMap, uidMap[wsid][1])
		delete(uidMap, wsid)
		delete(Rooms, wsid)
	}
	return
}

func (c *Connection) writer(wsid string) {
	defer fmt.Println("关闭连接 writer")
	for message := range c.Sc {
		c.Ws.WriteMessage(websocket.TextMessage, message)
		val, ok := Rooms[wsid]
		if !ok || (ok && val.ChatID == -2) { // 判断房间是否已经被删除
			if ok {
				lock.Lock()
				Rooms[wsid].ChatID = -3 //-3 关闭writer
				lock.Unlock()
			}
			return
		}
	}
}

// getOtherID 获取房间的另一个人的id
func getOtherID(wsid string, id int) int {
	uids, ok := uidMap[wsid]
	if ok {
		if uids[0] == id {
			return uids[1]
		}
		return uids[0]
	}
	return -1
}

// 判断房间是否存在，同时判断userID-wsID   wsID-userID映射
func roomExist(wsid string) bool {
	var ok [4]bool
	var uids []int
	_, ok[0] = Rooms[wsid]
	uids, ok[1] = uidMap[wsid]
	if ok[1] {
		_, ok[2] = wsidMap[uids[0]]
		_, ok[3] = wsidMap[uids[1]]
	}
	return ok[0] && ok[1] && ok[2] && ok[3]
}

func (c *Connection) reader(wsid string) {
	defer fmt.Println("关闭连接 reader")
	for {
		if !roomExist(wsid) { // 判断房间是否已经被删除
			break
		}
		_, message, err := c.Ws.ReadMessage()
		if err != nil {
			break
		}
		json.Unmarshal(message, c.Data)

		rWsid, ok := wsidMap[c.Data.UserID]

		if !ok || rWsid != wsid {
			// 用户id和wsid不匹配
			fmt.Printf("用户id和wsid不匹配  userid:%d   wsid:%s\n", c.Data.UserID, rWsid)
			c.Data.Type = "close"
		}

		c.Data.Time = time.Now().Unix()
		switch c.Data.Type {
		case "ping":
			c.Data.Type = "pong"
			dataByte, _ := json.Marshal(c.Data)
			Rooms[wsid].Stream <- dataByte
		case "message":
			c.Data.Type = "message"
			dataByte, _ := json.Marshal(c.Data)

			dataByte2, _ := json.Marshal(toData(c.Data, c.Ws.RemoteAddr().String(), getOtherID(wsid, c.Data.UserID)))
			Redis.SetList("chat:history:"+wsid, dataByte2)
			// data储存到Redis作为聊天记录
			Rooms[wsid].Stream <- dataByte
		case "close":
			c.Data.Type = "close"
			dataByte, _ := json.Marshal(c.Data)
			Rooms[wsid].Stream <- dataByte
			Rooms[wsid].Conn <- c
			lock.Lock()
			Rooms[wsid].ChatID = -2 //-2 关闭reader
			lock.Unlock()
			return
		}
	}
}
