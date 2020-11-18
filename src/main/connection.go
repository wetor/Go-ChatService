package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type connection struct {
	ws   *websocket.Conn
	sc   chan []byte
	data *Data
}

var wu = &websocket.Upgrader{ReadBufferSize: 512,
	WriteBufferSize: 512, CheckOrigin: func(r *http.Request) bool { return true }}

func myws(w http.ResponseWriter, r *http.Request) {
	wsid := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	fmt.Println("myws wsid:" + wsid)
	ws, err := wu.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("ws创建失败")
		return
	}
	c := &connection{sc: make(chan []byte, 256), ws: ws, data: &Data{}}
	h[wsid].r <- c
	go c.writer()
	c.reader(wsid)
	defer func() {
		c.data.Type = "logout"
		user_list[wsid] = del(user_list[wsid], c.data.User)
		c.data.UserList = user_list[wsid]
		c.data.Content = c.data.User
		data_b, _ := json.Marshal(c.data)
		h[wsid].b <- data_b
		h[wsid].r <- c
	}()
}

func (c *connection) writer() {
	for message := range c.sc {
		c.ws.WriteMessage(websocket.TextMessage, message)
	}
	c.ws.Close()
}

func (c *connection) reader(wsid string) {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			h[wsid].r <- c
			break
		}
		json.Unmarshal(message, &c.data)
		switch c.data.Type {
		case "login":
			c.data.User = c.data.Content
			c.data.From = c.data.User
			user_list[wsid] = append(user_list[wsid], c.data.User)
			c.data.UserList = user_list[wsid]
			data_b, _ := json.Marshal(c.data)
			h[wsid].b <- data_b
		case "user":
			c.data.Type = "user"
			data_b, _ := json.Marshal(c.data)
			h[wsid].b <- data_b
		case "logout":
			c.data.Type = "logout"
			user_list[wsid] = del(user_list[wsid], c.data.User)
			data_b, _ := json.Marshal(c.data)
			h[wsid].b <- data_b
			h[wsid].r <- c
		default:
			fmt.Print("========default================")
		}
	}
}

func del(slice []string, user string) []string {
	count := len(slice)
	if count == 0 {
		return slice
	}
	if count == 1 && slice[0] == user {
		return []string{}
	}
	var n_slice = []string{}
	for i := range slice {
		if slice[i] == user && i == count {
			return slice[:count]
		} else if slice[i] == user {
			n_slice = append(slice[:i], slice[i+1:]...)
			break
		}
	}
	fmt.Println(n_slice)
	return n_slice
}