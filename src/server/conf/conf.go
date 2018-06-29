package conf

import (
	"time"
	"io/ioutil"
	"encoding/json"
	golog "log"

	"github.com/name5566/leaf/log"
	"path/filepath"
)

var (
	// log conf
	LogFlag = golog.LstdFlags

	// gate conf
	PendingWriteNum        = 2000
	MaxMsgLen       uint32 = 4096
	HTTPTimeout            = 10 * time.Second
	LenMsgLen              = 2
	LittleEndian           = false

	// skeleton conf
	GoLen              = 10000
	TimerDispatcherLen = 10000
	AsynCallLen        = 10000
	ChanRPCLen         = 10000
)

// 配置文件server初始化结构体
var Server struct {
	LogLevel     string
	LogPath      string
	WSAddr       string
	CertFile     string
	KeyFile      string
	TCPAddr      string
	MaxConnNum   int
	ConsolePort  int
	ProfilePath  string
	HTTPAddr     string
	HTTPCertFile string
	HTTPKeyFile  string
}

func InitServerConfig(confPath string) {
	data, err := ioutil.ReadFile(filepath.Join(confPath, "server.json"))
	if err != nil {
		log.Fatal("%v \n", err)
	}
	err = json.Unmarshal(data, &Server)
	if err != nil {
		log.Fatal("%v \n", err)
	}
	log.Release("Server Config: %v \n", &Server)
}
