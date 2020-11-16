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

// Response 响应
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// ChatMatch 匹配
func ChatMatch(w http.ResponseWriter, r *http.Request) {
	var userID userID
	if err := json.NewDecoder(r.Body).Decode(&userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	matchID, wsID := chat.Match(userID.ID) // 开始匹配，等待

	var resp Response
	var jsonByte []byte
	if matchID == -2 { // 取消
		resp = Response{
			Code: 602,
			Msg:  "match cancel",
			Data: nil,
		}
		jsonByte, _ = json.Marshal(resp)
	} else if matchID == -1 { // 超时
		resp = Response{
			Code: 601,
			Msg:  "match timeout",
			Data: nil,
		}
		jsonByte, _ = json.Marshal(resp)
	} else {
		resp = Response{
			Code: 200,
			Msg:  "match success",
			Data: data.GetUser(matchID),
		}
		jsonByte, _ = json.Marshal(resp)
		fmt.Println(wsID)
	}
	w.Write(jsonByte)
	return
}
