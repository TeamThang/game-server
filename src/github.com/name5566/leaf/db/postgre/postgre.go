// postgre初始化
package postgre

import (
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/name5566/leaf/conf"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/db/postgre/model"
)

var DB *gorm.DB

// 初始化数据库连接实例
func InitDB() {
	Config := conf.Config.Postgre
	conArgs := "host=%s port=%s user=%s dbname=%s password=%s sslmode=disable"
	conArgs = fmt.Sprintf(conArgs, Config.Host, strconv.Itoa(Config.Port), Config.User, Config.DbName, Config.PassWord)
	log.Debug("PostgreConfig: %v\n", conArgs)
	var err error
	DB, err = gorm.Open("postgres", conArgs)
	if err != nil {
		log.Fatal("初始化Postgre数据库失败: %v", err)
	}
	DB.DB().SetMaxIdleConns(10)
	DB.DB().SetMaxOpenConns(100)
	DB.LogMode(true) // 是否打印sql
	updateTable(DB)
}

// gorm自动同步表结构
// 外键显示添加
func updateTable(db *gorm.DB) {
	db.SingularTable(true)
	db.AutoMigrate(&model.User{})
	// 添加外键关联，注意user是postgre关键字
	// 关联user表

}
