package data

import (
	"fmt"
	"strconv"
)

// GetUser 通过ID获取User
func GetUser(ID int) User {
	fmt.Println("getUser " + strconv.Itoa(ID))
	// 通过rpc从数据服务获取数据
	return UserList[ID]
}
