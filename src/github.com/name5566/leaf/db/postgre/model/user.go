// 表结构定义
// gorm
package model

import (
	"github.com/jinzhu/gorm"
)

// 用户表
type User struct {
	gorm.Model
	UserID uint `gorm:"unique_index"` // 对应登陆服务的userID，统一userID管理用户

}
