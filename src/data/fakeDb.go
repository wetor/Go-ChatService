package data

import (
	"fmt"
)

// UserList user列表
var UserList []User

func init() {
	fmt.Println("fakeDb init")
	UserList = append(UserList,
		User{0, "lisi@163.com", "LiSi", "男", "河南", "/avator/1.jpg", []string{"科技", "音乐"}},
		User{1, "zhangsan@163.com", "Zhang", "男", "河南", "/avator/2.jpg", []string{"游戏", "艺术", "音乐"}},
		User{2, "wangwu@163.com", "Wang", "男", "河北", "/avator/4.jpg", []string{"游戏", "音乐"}},
	)
}
