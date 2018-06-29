package main

import (
	"github.com/name5566/leaf"
	lconf "github.com/name5566/leaf/conf"
	"server/conf"
	"server/game"
	"server/gate"
	"server/login"
)

var confPath = "bin/conf/"

// 初始化config.json配置文件
func init() {
	conf.InitServerConfig(confPath)
	leaf.InitDB(confPath)
}

func main() {
	lconf.LogLevel = conf.Server.LogLevel
	lconf.LogPath = conf.Server.LogPath
	lconf.LogFlag = conf.LogFlag
	lconf.ConsolePort = conf.Server.ConsolePort
	lconf.ProfilePath = conf.Server.ProfilePath

	leaf.Run(
		game.Module,
		gate.Module,
		login.Module,
	)
}
