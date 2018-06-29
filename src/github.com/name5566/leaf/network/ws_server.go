package network

import (
	"net"
	"net/http"
	"sync"
	"time"
	"fmt"
	"crypto/tls"
	"github.com/gorilla/websocket"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	tk "github.com/name5566/leaf/db/redis/token"
	"github.com/name5566/leaf/gate/user"
)

type WSServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint32
	HTTPTimeout     time.Duration
	CertFile        string
	KeyFile         string
	NewAgent        func(*WSConn) Agent
	ln              net.Listener
	handler         *WSHandler
}

type WSHandler struct {
	maxConnNum      int
	pendingWriteNum int
	maxMsgLen       uint32
	newAgent        func(*WSConn) Agent
	upgrader        websocket.Upgrader
	conns           WebsocketConnSet
	mutexConns      sync.Mutex
	wg              sync.WaitGroup
}

// websocket handler
// 新建连接则调用改方法初始化一个handler
// 升级http到websocket
func (handler *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	responseHeader := http.Header{}
	log.Debug("INIT HANDLER:  handler adr: %p, \n", handler)
	var userData *user.UserData
	cookies := r.Cookies()
	fmt.Println("Cookies: ", cookies)
	if cookies != nil {
		var err error
		if userData, err = checkCookies(cookies); err != nil {
			log.Error("check cookies error: %v", err)
		}
	}
	conn, err := handler.upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		log.Debug("upgrade error: %v", err)
		return
	}
	conn.SetReadLimit(int64(handler.maxMsgLen))

	handler.wg.Add(1)
	defer handler.wg.Done()

	handler.mutexConns.Lock()
	if handler.conns == nil { // 连接池是nil，close
		handler.mutexConns.Unlock()
		conn.Close()
		return
	}
	if len(handler.conns) >= handler.maxConnNum { // 当前连接大于最大连接返回错误
		handler.mutexConns.Unlock()
		conn.Close()
		log.Debug("too many connections")
		return
	}
	handler.conns[conn] = struct{}{}
	handler.mutexConns.Unlock()

	wsConn := newWSConn(conn, handler.pendingWriteNum, handler.maxMsgLen)
	agent := handler.newAgent(wsConn) // 调用gate的gate.Run中实现的NewAgent方法，创建当前连接的network.Agent
	if userData != nil {
		agent.OnInit(userData)
	}
	agent.Run() // gate中的agent.Run方法，该方法处理这个连接的请求

	// cleanup
	wsConn.Close()
	handler.mutexConns.Lock()
	delete(handler.conns, conn)
	handler.mutexConns.Unlock()
	agent.OnClose()
}

// 调用gorilla.websocket启动websocket服务
func (server *WSServer) Start() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Release("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		log.Release("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.MaxMsgLen <= 0 {
		server.MaxMsgLen = 4096
		log.Release("invalid MaxMsgLen, reset to %v", server.MaxMsgLen)
	}
	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		log.Release("invalid HTTPTimeout, reset to %v", server.HTTPTimeout)
	}
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	if server.CertFile != "" || server.KeyFile != "" {
		config := &tls.Config{}
		config.NextProtos = []string{"http/1.1"}

		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(server.CertFile, server.KeyFile)
		if err != nil {
			log.Fatal("%v", err)
		}

		ln = tls.NewListener(ln, config)
	}
	log.Release("ws server init: maxMsgLen: %d, pendingWriteNum: %d\n", server.MaxMsgLen, server.PendingWriteNum)
	server.ln = ln
	server.handler = &WSHandler{
		maxConnNum:      server.MaxConnNum,
		pendingWriteNum: server.PendingWriteNum,
		maxMsgLen:       server.MaxMsgLen,
		newAgent:        server.NewAgent,
		conns:           make(WebsocketConnSet),
		upgrader: websocket.Upgrader{
			HandshakeTimeout: server.HTTPTimeout,
			CheckOrigin:      func(_ *http.Request) bool { return true },
		},
	}

	httpServer := &http.Server{
		Addr:           server.Addr,
		Handler:        server.handler,
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}

	go httpServer.Serve(ln)
}

// websocket服务关闭
func (server *WSServer) Close() {
	server.ln.Close()

	server.handler.mutexConns.Lock()
	for conn := range server.handler.conns {
		conn.Close()
	}
	server.handler.conns = nil
	server.handler.mutexConns.Unlock()

	server.handler.wg.Wait()
}

// 检查cookies
func checkCookies(cookies [] *http.Cookie) (*user.UserData, error) {
	for _, cookie := range cookies {
		switch cookie.Name {
		case "token":
			token := cookie.Value
			userID, _, maxAge, err := tk.GetTokenValByID(token)
			if err != nil {
				return nil, err
			}
			userData := user.UserData{
				UserID:  userID,
				Token:   token,
				Expired: util.GetExpiredTime(int(maxAge)),
			}
			return &userData, nil
		}
	}
	return nil, nil
}

// 设置跨域资源共享(CORS)请求头
func setCORS(r *http.Request, h *http.Header) {
	reqOrigin := r.Header.Get("Origin")
	log.Debug("req origin: %s", reqOrigin)
	if reqOrigin != "" {
		h.Set("Access-Control-Allow-Origin", reqOrigin)
		//responseHeader.Set("Access-Control-Allow-Headers", "*")
		//responseHeader.Set("Access-Control-Allow-Methods", "GET")
		h.Set("Access-Control-Allow-Credentials", "true")
	}

}
