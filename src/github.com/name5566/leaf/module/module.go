package module

import (
	"github.com/name5566/leaf/conf"
	"github.com/name5566/leaf/log"
	"runtime"
	"sync"
)

//模块接口定义
type Module interface {
	OnInit()                // 初始化函数
	OnDestroy()             // 销毁函数
	Run(closeSig chan bool) // 运行函数
}

//模块类型定义
type module struct {
	mi       Module         //实现了模块接口的对象
	closeSig chan bool      //传输关闭信号的管道
	wg       sync.WaitGroup //等待组
}

//模块数组，用于保存注册的模块
var mods []*module

// 注册一个新模块
func Register(mi Module) {
	m := new(module)                //新建一个模块
	m.mi = mi                       //保存实现了模块接口的某对象
	m.closeSig = make(chan bool, 1) //创建传输关闭信号的管道

	mods = append(mods, m) //保存模块到模块数组中
}

//初始化函数，注意不是init
func Init() {
	for i := 0; i < len(mods); i++ { //遍历所有注册的模块(从前往后)
		mods[i].mi.OnInit() //调用各个模块的OnInit函数
	}

	for i := 0; i < len(mods); i++ { //遍历所有注册的模块(从前往后)
		m := mods[i]
		m.wg.Add(1) //等待goroutine数加1
		go run(m)   //在一个新的goroutine中运行模块
	}
}

//销毁函数
func Destroy() {
	for i := len(mods) - 1; i >= 0; i-- { //遍历所有注册的模块(反序，从后往前)
		m := mods[i]       // 取得对应索引的模块
		m.closeSig <- true //向管道发送关闭信号(导致Run内的死循环介绍，继续执行到m.wg.Done() 60行)
		m.wg.Wait()        //等待该模块所在goroutine执行完成
		destroy(m)         //销毁该模块
	}
}

//运行模块函数定义
func run(m *module) {
	m.mi.Run(m.closeSig) //调用模块的Run函数(skeleton内实现，一个死循环)
	m.wg.Done()          //等待goroutine数减1
}

//销毁模块
func destroy(m *module) {
	defer func() { //延迟执行
		if r := recover(); r != nil { //捕获异常
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf) //创建一个字节切片用于存储格式化后的stack trace
				l := runtime.Stack(buf, false)        //格式化调用Stack函数的goroutine的stack trace
				log.Error("%v: %s", r, buf[:l])       //打印错误消息和stack trace
			} else {
				log.Error("%v", r) //只打印错误消息
			}
		}
	}()

	m.mi.OnDestroy() //先调用模块的销毁函数，再执行上面的延迟函数
}
