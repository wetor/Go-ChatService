package chat

import (
	"ChatService/src/data"
	"encoding/json"
)

// Room 一个聊天房间
type Room struct {
	ChatID int                  // 聊天id，用于定位聊天记录
	WsID   string               // 房间wsid
	Users  [2]data.User         // 两个用户
	Flag   map[*Connection]bool //连接是否存活 c
	Conn   chan *Connection     // 连接数据 r
	Stream chan []byte          // 数据交换流 b
}

// Run 执行
func (room *Room) Run() {
	for {
		select {
		case c := <-room.Conn:
			room.Flag[c] = true
			c.Data.IP = c.Ws.RemoteAddr().String()
			c.Data.Type = "handshake"
			dataByte, _ := json.Marshal(c.Data)
			c.Sc <- dataByte
		case data := <-room.Stream:
			for c := range room.Flag {
				select {
				case c.Sc <- data:
				default:
					delete(room.Flag, c)
					close(c.Sc)
				}
			}
		}
	}
}
