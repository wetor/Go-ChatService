package main

import (
	"encoding/json"
	"fmt"
)

var h map[string]hub

type hub struct {
	c map[*connection]bool
	b chan []byte
	r chan *connection
	u chan *connection
}

func init() {
	fmt.Println("main init")
	h = make(map[string]hub)
	h["test"] = hub{
		c: make(map[*connection]bool),
		u: make(chan *connection),
		b: make(chan []byte),
		r: make(chan *connection),
	}
	h["test2"] = hub{
		c: make(map[*connection]bool),
		u: make(chan *connection),
		b: make(chan []byte),
		r: make(chan *connection),
	}
	h["test3"] = hub{
		c: make(map[*connection]bool),
		u: make(chan *connection),
		b: make(chan []byte),
		r: make(chan *connection),
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.r:
			h.c[c] = true
			c.data.Ip = c.ws.RemoteAddr().String()
			c.data.Type = "handshake"
			c.data.UserList = user_list
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
