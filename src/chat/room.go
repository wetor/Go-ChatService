package chat

import (
	"ChatService/src/data"
	"encoding/json"
)

// Room 一个聊天房间
type Room struct {
	ChatID int              // 聊天id，用于定位聊天记录
	WsID   string           // 房间wsid
	Users  []data.User      // 两个用户
	Conns  []*Connection    // 两个用户的ws连接
	Conn   chan *Connection // 连接数据 r
	Stream chan []byte      // 数据交换流 b
}

// WsData websocket的消息传递结构
type WsData struct {
	UserID int    `json:"userid"`
	Type   string `json:"type"`
	Time   int64  `json:"time"`
	Data   string `json:"data"`
}

func toData(wsdata *WsData, ip string, sendid int) *data.Data {
	return &data.Data{
		IP:      ip,
		UserID:  wsdata.UserID,
		SendID:  sendid,
		Time:    wsdata.Time,
		Type:    wsdata.Type,
		Content: wsdata.Data,
	}
}
func toWsData(data *data.Data) *WsData {
	return &WsData{
		UserID: data.UserID,
		Time:   data.Time,
		Type:   data.Type,
		Data:   data.Content,
	}
}

// Run 执行
func (room *Room) Run() {
	for {
		select {
		case c := <-room.Conn:
			if c.Data.Type == "start" {
				room.Conns = append(room.Conns, c)
			}
			dataByte, _ := json.Marshal(c.Data)
			c.Sc <- dataByte
		case data := <-room.Stream:
			// 向双方发送数据
			for _, c := range room.Conns {
				c.Sc <- data
			}
		}
	}
}
