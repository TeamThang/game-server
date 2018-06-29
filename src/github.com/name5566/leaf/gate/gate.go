package gate

import (
	"net"
	"reflect"
	"time"
	"fmt"

	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
	"github.com/name5566/leaf/gate/user"
	"net/http"
)

//websocket和tcp协议网关服务器定义
type Gate struct {
	MaxConnNum      int               //最大连接数
	PendingWriteNum int               //发送缓冲区长度
	MaxMsgLen       uint32            //最大消息长度
	Processor       network.Processor //json或protobuf处理器
	AgentChanRPC    *chanrpc.Server   //RPC服务器

	// websocket
	WSAddr      string // websocket监听地址
	HTTPTimeout time.Duration
	CertFile    string
	KeyFile     string

	// tcp
	TCPAddr      string // tcp监听地址
	LenMsgLen    int    // 消息长度占用字节数
	LittleEndian bool   // 大小端标志

	// http
	HTTPAddr     string // http监听地址
	HTTPCertFile string
	HTTPKeyFile  string
	ServeMux	 http.ServeMux
}

//实现了Module接口的Run
func (gate *Gate) Run(closeSig chan bool) {
	var wsServer *network.WSServer
	if gate.WSAddr != "" {
		wsServer = new(network.WSServer) //创建websocket服务对象
		//设置websocket服务器相关参数
		wsServer.Addr = gate.WSAddr
		wsServer.MaxConnNum = gate.MaxConnNum
		wsServer.PendingWriteNum = gate.PendingWriteNum
		wsServer.MaxMsgLen = gate.MaxMsgLen
		wsServer.HTTPTimeout = gate.HTTPTimeout
		wsServer.CertFile = gate.CertFile
		wsServer.KeyFile = gate.KeyFile
		wsServer.NewAgent = func(conn *network.WSConn) network.Agent { // 设置创建代理函数, 关联gate和conn
			a := &agent{conn: conn, gate: gate}
			fmt.Printf("gate.Agent: %v\n", a)
			fmt.Printf("gate.AgentChanRPC: %v\n", gate.AgentChanRPC)
			if gate.AgentChanRPC != nil {
				gate.AgentChanRPC.Go("NewAgent", a)
				log.Debug("new agent: %v\n", a)
			}
			return a
		}
	}

	var tcpServer *network.TCPServer
	if gate.TCPAddr != "" {
		tcpServer = new(network.TCPServer) //创建TCP服务对象
		//设置TCP服务器相关参数
		tcpServer.Addr = gate.TCPAddr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.PendingWriteNum = gate.PendingWriteNum
		tcpServer.LenMsgLen = gate.LenMsgLen
		tcpServer.MaxMsgLen = gate.MaxMsgLen
		tcpServer.LittleEndian = gate.LittleEndian
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent { //设置创建代理函数
			a := &agent{conn: conn, gate: gate}
			if gate.AgentChanRPC != nil {
				gate.AgentChanRPC.Go("NewAgent", a)
			}
			return a
		}
	}

	var httpServer *network.HttpServer
	if gate.HTTPAddr != "" {
		httpServer = new(network.HttpServer) //创建http服务对象
		//设置http服务器相关参数
		httpServer.Addr = gate.HTTPAddr
		httpServer.HTTPTimeout = gate.HTTPTimeout
		httpServer.CertFile = gate.HTTPCertFile
		httpServer.KeyFile = gate.HTTPKeyFile
		httpServer.Handler = gate.ServeMux
	}

	if wsServer != nil {
		wsServer.Start()
	}
	if tcpServer != nil {
		tcpServer.Start()
	}
	if httpServer != nil {
		httpServer.Start()
	}
	<-closeSig
	if wsServer != nil {
		wsServer.Close()
	}
	if tcpServer != nil {
		tcpServer.Close()
	}
	if httpServer != nil {
		httpServer.Close()
	}
}

//Module接口的OnInit
func (gate *Gate) OnInit() {}

//Module接口的OnDestroy
func (gate *Gate) OnDestroy() {}

//代理类型定义
type agent struct {
	conn     network.Conn // 连接接口
	gate     *Gate        // 网关类型
	userData interface{}  // 用户数据
}

// 实现代理接口(network.Agent)OnInit函数
// 当前主要为初始化websocket连接时用户数据赋值
func (a *agent) OnInit(data interface{}) {
	if userData, ok := data.(*user.UserData); ok {
		a.SetUserData(*userData)
		log.Debug("UserData Set: %v", a.userData)
	}
}

//实现代理接口(network.Agent)Run函数
func (a *agent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		if a.gate.Processor != nil {
			fmt.Printf("agent run, conn: %v, gate: %v, data: %v \n", a.conn, a.gate, data)
			msg, err := a.gate.Processor.Unmarshal(data)
			if err != nil {
				log.Debug("unmarshal message error: %v", err)
				break
			}
			fmt.Printf("Route: a: %v, a.userData: %v\n", a, a.userData)
			err = a.gate.Processor.Route(msg, a)
			if err != nil {
				log.Debug("route message error: %v", err)
				break
			}
		}
	}
}

//实现代理接口(gate.Agent)OnClose函数
func (a *agent) OnClose() {
	if a.gate.AgentChanRPC != nil {
		err := a.gate.AgentChanRPC.Call0("CloseAgent", a)
		if err != nil {
			log.Error("chanrpc error: %v", err)
		}
	}
}

//实现代理接口(gate.Agent)WriteMsg函数
//发送消息
func (a *agent) WriteMsg(msg interface{}) {
	if a.gate.Processor != nil {
		data, err := a.gate.Processor.Marshal(msg)
		if err != nil {
			log.Error("marshal message %v error: %v", reflect.TypeOf(msg), err)
			return
		}
		err = a.conn.WriteMsg(data...)
		if err != nil {
			log.Error("write message %v error: %v", reflect.TypeOf(msg), err)
		}
	}
}

func (a *agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

//实现代理接口(gate.Agent)Close函数
//关闭代理
func (a *agent) Close() {
	a.conn.Close()
}

//实现代理接口(network.Agent)Destroy函数
func (a *agent) Destroy() {
	a.conn.Destroy()
}

//实现代理接口(gate.Agent)UserData函数
//获取用户数据
func (a *agent) UserData() interface{} {
	if a.userData == nil {
		return nil
	}
	userData, ok := a.userData.(user.UserData)
	if !ok {
		log.Error("user data %v is not valid:  ", userData)
		return nil
	}
	if time.Now().After(userData.Expired) {
		log.Release("user data %v is expired: ", userData)
		return nil
	}
	return a.userData
}

//实现代理接口(gate.Agent)SetUserData函数
//设置用户数据
func (a *agent) SetUserData(data interface{}) {
	a.userData = data
}
