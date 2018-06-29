package conf

import (
	"io/ioutil"
	"encoding/json"
	"github.com/name5566/leaf/log"
	"path/filepath"
)

type Postgre struct {
	Host     string
	Port     int
	User     string
	DbName   string
	PassWord string
}

type Redis struct {
	Host     string
	Port     int
	PassWord string
	DB       int
}


// 配置文件config初始化结构体
var Config struct {
	Postgre      Postgre
	Redis        Redis
	LoginServer  string
}

var (
	LenStackBuf = 4096

	// log
	LogLevel string
	LogPath  string
	LogFlag  int

	// console
	ConsolePort   int
	ConsolePrompt string = "Leaf# "
	ProfilePath   string

	// cluster
	ListenAddr      string
	ConnAddrs       []string
	PendingWriteNum int
)

func InitConfig(confPath string) {
	data, err := ioutil.ReadFile(filepath.Join(confPath, "config.json"))
	if err != nil {
		log.Fatal("%v \n", err)
	}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		log.Fatal("parse config.json failed: %v \n %v\n", err, string(data))
	}
}

