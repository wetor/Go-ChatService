package data

import "github.com/gorilla/websocket"

// User 用户结构
type User struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Name     string   `json:"name"`
	Sex      string   `json:"sex"`
	Region   string   `json:"region"`
	Avator   string   `json:"avator"`
	Label    []string `json:"label"`
}

// Data 发送和接受的信息
type Data struct {
	IP      string `json:"ip"`
	UserID  int    `json:"userid"`
	FormID  int    `json:"formid"`
	Time    int    `json:"time"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

// Connection 连接信息，房间
type Connection struct {
	Ws    *websocket.Conn
	WsID  string
	Users [2]User
	Sc    chan []byte // 用于储存发送数据的chan
	Data  *Data       // 当前正要发送的数据
}

// WaitUser 正在等待用户
type WaitUser struct {
	User User `json:"user"`
	Time int  `json:"time"` // 已经等待的时间
}

// MqUser RabbitMQ结构
type MqUser struct {
	ID      int `json:"id"`      // userID
	MatchID int `json:"matchid"` // matchID
	Type    int `json:"type"`    // 1:被匹配到 -1:匹配被取消
}
