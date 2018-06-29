package user

import "time"

// 当前连接用户数据
type UserData struct {
	UserID  uint
	Token   string
	Expired time.Time
}

