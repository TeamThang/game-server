package network

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/name5566/leaf/log"
	"net"
	"sync"
	"fmt"
)

type WebsocketConnSet map[*websocket.Conn]struct{}

type WSConn struct {
	sync.Mutex
	conn      *websocket.Conn
	writeChan chan []byte
	maxMsgLen uint32
	closeFlag bool
}

// 通过websocket连接conn新建WSConn对象,
// 启动go协程去队列writeChan中取数据写入conn返回给客户端
// 写完之后置该对象关闭标志为true
// pendingWriteNum: 写入队列容量, maxMsgLen: 最大消息长度
func newWSConn(conn *websocket.Conn, pendingWriteNum int, maxMsgLen uint32) *WSConn {
	wsConn := new(WSConn)
	wsConn.conn = conn

	wsConn.writeChan = make(chan []byte, pendingWriteNum)
	wsConn.maxMsgLen = maxMsgLen
	fmt.Printf("ws wrtite chan init: maxMsgLen: %d, pendingWriteNum: %d\n", maxMsgLen, pendingWriteNum)
	go func() {
		for b := range wsConn.writeChan {
			fmt.Sprintf("new ws conn write message: %v\n", b)
			if b == nil {
				break
			}

			err := conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				break
			}
		}

		conn.Close()
		wsConn.Lock()
		wsConn.closeFlag = true
		wsConn.Unlock()
	}()

	return wsConn
}

// 放弃未发送数据并关闭对应WSConn对象的ws连接
func (wsConn *WSConn) doDestroy() {
	wsConn.conn.UnderlyingConn().(*net.TCPConn).SetLinger(0)
	wsConn.conn.Close()

	if !wsConn.closeFlag {
		close(wsConn.writeChan)
		wsConn.closeFlag = true
	}
}

// 加锁调用doDestroy
func (wsConn *WSConn) Destroy() {
	wsConn.Lock()
	defer wsConn.Unlock()

	wsConn.doDestroy()
}

// 置对应WSConn对象的关闭标志未true
func (wsConn *WSConn) Close() {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return
	}

	wsConn.doWrite(nil)
	wsConn.closeFlag = true
}

// 向对应WSConn的写入队列写数据
// 如果WSConn的队列已满，阻塞了，就丢弃并关闭
func (wsConn *WSConn) doWrite(b []byte) {
	if len(wsConn.writeChan) == cap(wsConn.writeChan) {
		log.Debug("close conn: channel full")
		wsConn.doDestroy()
		return
	}
	fmt.Printf("ws do write: write chan: %v,content: %s\n", wsConn.writeChan, string(b))
	wsConn.writeChan <- b
}

func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}

// goroutine not safe
// 从对应WSConn的conn中读取数据
func (wsConn *WSConn) ReadMsg() ([]byte, error) {
	msgType, b, err := wsConn.conn.ReadMessage()
	fmt.Printf("ws read msg: msg type: %v, content: %v,err: %v\n", msgType, string(b), err)
	return b, err
}

// args must not be modified by the others goroutines
// 调用WSConn的doWrite向ws连接中写入一个或多个数据，加锁
func (wsConn *WSConn) WriteMsg(args ...[]byte) error {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return nil
	}

	// get len
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	// check len
	if msgLen > wsConn.maxMsgLen {
		return errors.New(fmt.Sprintf("message too long, current msg len: %d", msgLen))
	} else if msgLen < 1 {
		return errors.New("message too short")
	}
	// don't copy
	if len(args) == 1 {
		wsConn.doWrite(args[0])
		return nil
	}

	// merge the args
	msg := make([]byte, msgLen)
	l := 0
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}
	wsConn.doWrite(msg)

	return nil
}

func (wsConn *WSConn) SetCookie(map[string] string) error {
	return nil
}
