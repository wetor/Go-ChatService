package api

import (
	"ChatService/src/chat"
	"ChatService/src/data"
	"encoding/json"
	"fmt"
	"net/http"
)

type userID struct {
	ID int `json:"id"`
}

type chatConn struct {
	WsID   string `json:"wsid"`
	ChatID int    `json:"chatid"`
}

// Response 响应
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Match 匹配
func Match(w http.ResponseWriter, r *http.Request) {
	var userID userID
	if err := json.NewDecoder(r.Body).Decode(&userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	matchID, wsID := chat.Match(userID.ID) // 开始匹配，等待

	var resp Response
	var jsonByte []byte

	switch matchID {
	case -2: // 匹配被取消
		resp = Response{
			Code: 602,
			Msg:  "match cancel",
			Data: nil,
		}
		break
	case -1: //匹配超时
		resp = Response{
			Code: 601,
			Msg:  "match timeout",
			Data: nil,
		}
		break
	default:
		resp = Response{
			Code: 200,
			Msg:  "match success",
			Data: data.GetUser(matchID),
		}
		fmt.Println(wsID)
		break
	}
	jsonByte, _ = json.Marshal(resp)
	w.Write(jsonByte)
	return
}

// Cancel 取消匹配/聊天
func Cancel(w http.ResponseWriter, r *http.Request) {
	var userID userID
	if err := json.NewDecoder(r.Body).Decode(&userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	retn := chat.Cancel(userID.ID)

	var resp Response
	var jsonByte []byte
	switch retn {
	case -1: // 不在匹配或不在房间中
		resp = Response{
			Code: 202,
			Msg:  "not match",
			Data: nil,
		}
		break
	case 0: // 取消匹配
		resp = Response{
			Code: 200,
			Msg:  "match cancel",
			Data: nil,
		}
		break
	case 1: // 退出聊天
		resp = Response{
			Code: 201,
			Msg:  "chat cancel",
			Data: nil,
		}
		break
	}
	jsonByte, _ = json.Marshal(resp)
	w.Write(jsonByte)
	return
}

// Chat 开始聊天
func Chat(w http.ResponseWriter, r *http.Request) {
	var userID userID
	if err := json.NewDecoder(r.Body).Decode(&userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	chatid, wsid := chat.Chat(userID.ID, w, r)

	var resp Response
	var jsonByte []byte
	switch chatid {
	case -1: // 房间不存在
		resp = Response{
			Code: 404,
			Msg:  "room no exist",
			Data: nil,
		}
		break
	default: // 开始聊天
		resp = Response{
			Code: 200,
			Msg:  "chat start",
			Data: chatConn{WsID: wsid, ChatID: chatid},
		}
		break
	}
	jsonByte, _ = json.Marshal(resp)
	w.Write(jsonByte)
	return

}

// Report 举报此聊天用户
func Report(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
}

// Webscoket 聊天webscoket
func Webscoket(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
}
