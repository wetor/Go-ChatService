package data

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
	UserID  int    `json:"userid"` // 发送者
	SendID  int    `json:"sendid"` // 发送对象
	Time    int64  `json:"time"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

// Chat 一个聊天记录
type Chat struct {
}

// WaitUser 正在等待用户
type WaitUser struct {
	User User
	Time int // 已经等待的时间，单位：查询次数
}

// MqUser RabbitMQ结构
type MqUser struct {
	ID      int    `json:"id"`      // userID
	MatchID int    `json:"matchid"` // matchID
	WsID    string `json:"wsid"`    // webscoket id
	Type    int    `json:"type"`    // 1:被匹配到 -1:匹配被取消
}
