// http://www.topgoer.com/网络编程/WebSocket编程.html
package main

import (
	"encoding/json"
	"fmt"
)

var h map[string]*hub
var user_list map[string][]string

type hub struct {
	c    map[*connection]bool
	b    chan []byte
	r    chan *connection
	u    chan *connection
	wsid string
}



func init() {
	fmt.Println("main init")
	user_list = make(map[string][]string)

	h = make(map[string]*hub)
	h["test"] = &hub{
		c:    make(map[*connection]bool),
		u:    make(chan *connection),
		b:    make(chan []byte),
		r:    make(chan *connection),
		wsid: "test",
	}
	user_list["test"] = []string{}
	h["test2"] = &hub{
		c:    make(map[*connection]bool),
		u:    make(chan *connection),
		b:    make(chan []byte),
		r:    make(chan *connection),
		wsid: "test2",
	}
	user_list["test2"] = []string{}
	h["test3"] = &hub{
		c:    make(map[*connection]bool),
		u:    make(chan *connection),
		b:    make(chan []byte),
		r:    make(chan *connection),
		wsid: "test3",
	}
	user_list["test3"] = []string{}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.r:
			h.c[c] = true
			c.data.Ip = c.ws.RemoteAddr().String()
			c.data.Type = "handshake"
			c.data.UserList = user_list[h.wsid]
			data_b, _ := json.Marshal(c.data)
			c.sc <- data_b
		case c := <-h.u:
			fmt.Println("<-h.u")
			if _, ok := h.c[c]; ok {

				delete(h.c, c)
				close(c.sc)
			}
		case data := <-h.b:
			for c := range h.c {
				select {
				case c.sc <- data:
				default:
					delete(h.c, c)
					close(c.sc)
				}
			}
		}
	}
}
