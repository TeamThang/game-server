package network

import (
	"net"
	"time"
	"net/http"
	"github.com/name5566/leaf/log"
	"crypto/tls"
)

type HttpServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint32
	HTTPTimeout     time.Duration
	CertFile        string
	KeyFile         string
	ln              net.Listener
	Handler      	http.ServeMux
}


// 启动http服务
func (server *HttpServer) Start() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}
	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		log.Release("invalid HTTPTimeout, reset to %v", server.HTTPTimeout)
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
	log.Release("http server init: %s", server.Addr)
	server.ln = ln
	httpServer := &http.Server{
		Addr:           server.Addr,
		Handler:        &server.Handler,
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}

	go httpServer.Serve(ln)
}

// websocket服务关闭
func (server *HttpServer) Close() {
}

