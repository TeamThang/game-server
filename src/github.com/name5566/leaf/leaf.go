package leaf

import (
	"os"
	"os/signal"
	"github.com/name5566/leaf/cluster"
	"github.com/name5566/leaf/conf"
	"github.com/name5566/leaf/console"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/module"
	"github.com/name5566/leaf/db/postgre"
	"github.com/name5566/leaf/db/redis"
)

func Run(mods ...module.Module) { //...不定参数语法，参数类型都为module.Module
	// logger
	if conf.LogLevel != "" { //日志级别不为空
		logger, err := log.New(conf.LogLevel, conf.LogPath, conf.LogFlag) //创建一个logger
		if err != nil {
			panic(err)
		}
		log.Export(logger)   //替换默认的gLogger
		defer logger.Close() //Run函数返回,关闭logger
	}

	log.Release("Leaf %v starting up", version) //启动日志

	// module
	for i := 0; i < len(mods); i++ { //遍历传入的所有module
		module.Register(mods[i]) //注册module
	}
	module.Init() //初始化模块，并执行各个模块(在各个不同的goroutine里)

	// cluster
	cluster.Init() //初始化集群

	// console
	console.Init() //初始化控制台

	// close
	c := make(chan os.Signal, 1)                       //新建一个管道用于接收系统Signal
	signal.Notify(c, os.Interrupt, os.Kill)            //监听SIGINT和SIGKILL信号(linux下叫这个名字)
	sig := <-c                                         //读信号，没有信号时会阻塞goroutine
	log.Release("Leaf closing down (signal: %v)", sig) //关键日志 服务器关闭
	console.Destroy()                                  //销毁控制台
	cluster.Destroy()                                  //销毁集群
	module.Destroy()                                   //销毁模块
}

// 根据config.json初始化配置
// 初始化postgre和redis
func InitDB(confPath string) {
	conf.InitConfig(confPath)
	log.Release("Config: %v \n", conf.Config)
	// 初始化postgre
	postgre.InitDB()
	// 初始化redis
	redis.InitPool()
}

