package chanrpc

import (
	"errors"
	"fmt"
	"github.com/name5566/leaf/conf"
	"github.com/name5566/leaf/log"
	"runtime"
)

// one server per goroutine (goroutine not safe)
// one client per goroutine (goroutine not safe)
//rpc服务器定义
type Server struct {
	// id -> function
	//
	// function:
	// func(args []interface{})
	// func(args []interface{}) interface{}
	// func(args []interface{}) []interface{}
	functions map[interface{}]interface{} // id->func映射
	ChanCall  chan *CallInfo              // 调用管道（用于传递调用信息）
}

//调用信息
type CallInfo struct {
	f       interface{}   // 函数
	args    []interface{} // 参数
	chanRet chan *RetInfo // 返回值管道，用于传输返回值，可能是同步返回值管道也可能是异步返回值管道
	cb      interface{}   // 回调
}

//返回信息
type RetInfo struct {
	// nil
	// interface{}
	// []interface{}
	ret interface{} // 返回值
	err error       // 错误
	// callback:
	// func(err error)
	// func(ret interface{}, err error)
	// func(ret []interface{}, err error)
	cb interface{} //回调
}

//rpc客户端定义
type Client struct {
	s               *Server       // rpc服务器引用
	chanSyncRet     chan *RetInfo // 同步函数结果返回管道，大小为1
	ChanAsynRet     chan *RetInfo // 异步函数结果返回管道，大小为n
	pendingAsynCall int           // 待处理的异步调用
}

// 初始化rpc服务器
func NewServer(l int) *Server {
	s := new(Server)                                // 初始化Server结构体
	s.functions = make(map[interface{}]interface{}) // 初始化functions属性
	s.ChanCall = make(chan *CallInfo, l)            // 初始化ChanCall
	return s
}

// 断言i是否为
func assert(i interface{}) []interface{} {
	if i == nil {
		return nil
	} else {
		return i.([]interface{})
	}
}

// you must call the function before calling Open and Go
// 注册f(函数)
// 注册函数到Server实例的fuctions属性中
func (s *Server) Register(id interface{}, f interface{}) {
	switch f.(type) { //判断f的类型
	case func([]interface{}): //参数是切片，值任意。无返回值
	case func([]interface{}) interface{}: //参数是切片，值任意。返回值为一个任意值
	case func([]interface{}) []interface{}: //参数是切片，返回值也是切片，值均为任意
	default:
		panic(fmt.Sprintf("function id %v: definition of function is invalid", id)) //id对应的函数定义非法
	}

	if _, ok := s.functions[id]; ok { //判断映射是否存在
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	s.functions[id] = f //存储映射
}

// 将结果RetInfo写入CallInfo实例的返回值管道chanRet
// 将CallInfo的回调函数cb保存到返回信息RetInfo的回调函数cb中
func (s *Server) ret(ci *CallInfo, ri *RetInfo) (err error) {
	if ci.chanRet == nil { //返回管道不能为空
		return
	}

	//延迟捕获异常
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	ri.cb = ci.cb    //将回调函数保存到返回信息中
	ci.chanRet <- ri //将返回信息发送到返回值管道中
	return
}

// 执行RPC调用
// 执行CallInfo实例的f函数，对应结果调用ret函数写入CallInfo实例的返回管道chanRet
func (s *Server) exec(ci *CallInfo) (err error) {
	//延迟处理异常
	defer func() {
		if r := recover(); r != nil {
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				err = fmt.Errorf("%v: %s", r, buf[:l])
			} else {
				err = fmt.Errorf("%v", r)
			}

			s.ret(ci, &RetInfo{err: fmt.Errorf("%v", r)})
		}
	}()

	// execute
	switch ci.f.(type) { //判断f类型
	case func([]interface{}):               //无返回值
		ci.f.(func([]interface{}))(ci.args) //执行调用
		return s.ret(ci, &RetInfo{})        //返回值为空
	case func([]interface{}) interface{}:                      //一个返回值
		ret := ci.f.(func([]interface{}) interface{})(ci.args) //执行调用
		return s.ret(ci, &RetInfo{ret: ret})                   //一个返回值
	case func([]interface{}) []interface{}:                      //多个返回值
		ret := ci.f.(func([]interface{}) []interface{})(ci.args) //执行调用
		return s.ret(ci, &RetInfo{ret: ret})                     //多个返回值
	}

	panic("bug")
}

// rpc服务器实例根据调用信息CallInfo调用相应方法
func (s *Server) Exec(ci *CallInfo) {
	err := s.exec(ci)
	if err != nil {
		log.Error("%v", err)
	}
}

// goroutine safe
// RPC服务器根据id作为key或取fuctions属性里面对应函数
// 并将函数信息注册到自己的调用管道ChanCall中
func (s *Server) Go(id interface{}, args ...interface{}) {
	f := s.functions[id] // 根据id取得对应的f
	if f == nil {
		return
	}

	defer func() {
		recover()
	}()

	s.ChanCall <- &CallInfo{ //将调用消息传给rpc服务器的调用管道ChanCall
		f: f,
		args: args,
	}
}

// goroutine safe
func (s *Server) Call0(id interface{}, args ...interface{}) error {
	return s.Open(0).Call0(id, args...)
}

// goroutine safe
func (s *Server) Call1(id interface{}, args ...interface{}) (interface{}, error) {
	return s.Open(0).Call1(id, args...)
}

// goroutine safe
func (s *Server) CallN(id interface{}, args ...interface{}) ([]interface{}, error) {
	return s.Open(0).CallN(id, args...)
}

//关闭RPC服务器
func (s *Server) Close() {
	close(s.ChanCall) //关闭管道调用

	//遍历所有未处理完的消息，返回错误消息(rpc server已关闭)
	for ci := range s.ChanCall {
		s.ret(ci, &RetInfo{
			err: errors.New("chanrpc server closed"),
		})
	}
}

// goroutine safe
// 调用NewClient初始化一个rpc客户端
// 绑定改客户端到rpc服务端实例s
func (s *Server) Open(l int) *Client {
	c := NewClient(l) // 创建一个rpc客户端
	c.Attach(s)       // 保存rpc服务器引用
	return c
}

// 初始化一个rpc客户端
func NewClient(l int) *Client {
	c := new(Client)
	c.chanSyncRet = make(chan *RetInfo, 1) //创建一个管道用于传输同步调用返回信息，同步调用的管道大小一定为1，因为调用以后就需要阻塞读取返回
	c.ChanAsynRet = make(chan *RetInfo, l) //创建一个管道用于传输异步调用返回信息，异步调用的管道大小不一定为1
	return c
}

// 更新rpc客户端实例的rpc服务实例引用
func (c *Client) Attach(s *Server) {
	c.s = s
}

// 发起调用
func (c *Client) call(ci *CallInfo, block bool) (err error) {
	//捕获异常
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	if block { // 阻塞的。当管道满时，阻塞
		c.s.ChanCall <- ci // 将调用消息通过管道传输到rpc服务器
	} else { // 非阻塞的。当管道满时，返回管道已满错误，利用default特性检测chan是否已满
		select {
		case c.s.ChanCall <- ci:
		default:
			err = errors.New("chanrpc channel full")
		}
	}
	return
}

// 通过rpc客户端的rpc服务器实例引用获取该rpc服务器实例fucntions属性中对应id的函数
func (c *Client) f(id interface{}, n int) (f interface{}, err error) {
	if c.s == nil {
		err = errors.New("server not attached")
		return
	}

	f = c.s.functions[id] //根据id取得对应的f
	if f == nil {
		err = fmt.Errorf("function id %v: function not registered", id)
		return
	}

	var ok bool
	//根据n的值判断f类型是否正确
	switch n {
	case 0:
		_, ok = f.(func([]interface{})) //n为0，无返回值
	case 1:
		_, ok = f.(func([]interface{}) interface{}) //n为1，一个返回值
	case 2:
		_, ok = f.(func([]interface{}) []interface{}) //n为2，多个返回值
	default:
		panic("bug")
	}

	if !ok { //类型不匹配
		err = fmt.Errorf("function id %v: return type mismatch", id)
	}
	return
}

// 同步调用0
// 适合参数是切片，值任意。无返回值
// call0 call1 calln 可以将0 1 n记作0个返回值，1个返回值，n个返回值
func (c *Client) Call0(id interface{}, args ...interface{}) error {
	f, err := c.f(id, 0) // 获取函数
	if err != nil {
		return err
	}

	err = c.call(&CallInfo{ // 发起调用
		f: f,
		args: args,
		chanRet: c.chanSyncRet, // 同步函数结果返回管道
	}, true)
	if err != nil {
		return err
	}

	ri := <-c.chanSyncRet //读取结果
	return ri.err         //返回错误字段，代表是否有错
}

// 同步调用1
// 适合参数是切片，值任意。返回值为一个任意值
func (c *Client) Call1(id interface{}, args ...interface{}) (interface{}, error) {
	f, err := c.f(id, 1) //读取f
	if err != nil {
		return nil, err
	}

	err = c.call(&CallInfo{ //发起调用
		f: f,
		args: args,
		chanRet: c.chanSyncRet,
	}, true)
	if err != nil {
		return nil, err
	}

	ri := <-c.chanSyncRet //读取结果
	return ri.ret, ri.err //返回返回值字段和错误字段
}

// 同步调用N
// 适合参数是切片，返回值也是切片，值均为任意
func (c *Client) CallN(id interface{}, args ...interface{}) ([]interface{}, error) {
	f, err := c.f(id, 2) //读取f
	if err != nil {
		return nil, err
	}

	err = c.call(&CallInfo{ //发起调用
		f: f,
		args: args,
		chanRet: c.chanSyncRet,
	}, true)
	if err != nil {
		return nil, err
	}

	ri := <-c.chanSyncRet         //读取结果
	return assert(ri.ret), ri.err //返回返回值字段（先转化类型）和错误字段
}

//发起异步调用(内部的)
func (c *Client) asynCall(id interface{}, args []interface{}, cb interface{}, n int) {
	f, err := c.f(id, n) // 获得函数
	if err != nil {
		c.ChanAsynRet <- &RetInfo{err: err, cb: cb}
		return
	}

	err = c.call(&CallInfo{ // 写入rpc服务的调用管道
		f: f,
		args: args,
		chanRet: c.ChanAsynRet, //异步返回管道
		cb: cb,
	}, false)
	if err != nil {
		c.ChanAsynRet <- &RetInfo{err: err, cb: cb} // 如果异常返回错误和回调函数
		return
	}
}

// 发起异步调用(导出的)
// 异步调用，需要自己写c.Cb(<-c.ChanAsynRet)执行回调
func (c *Client) AsynCall(id interface{}, _args ...interface{}) {
	if len(_args) < 1 { // 检查是否提供了回调函数参数，参数个数必定大于等于1
		panic("callback function not found")
	}

	args := _args[:len(_args)-1] // 取出RPC调用的参数, args 最后一个是回调函数，前面的是RPC调用的参数
	cb := _args[len(_args)-1]    // 取出回调函数

	var n int
	switch cb.(type) { // 判断回调函数的类型
	case func(error): // 只接收一个错误
		n = 0
	case func(interface{}, error): // 接收一个返回值和一个错误
		n = 1
	case func([]interface{}, error): // 接收多个返回值和一个错误
		n = 2
	default:
		panic("definition of callback function is invalid")
	}

	// too many calls
	if c.pendingAsynCall >= cap(c.ChanAsynRet) {
		execCb(&RetInfo{err: errors.New("too many calls"), cb: cb})
		return
	}

	c.asynCall(id, args, cb, n)
	c.pendingAsynCall++ // 增加计数器，待处理的异步调用
}

//执行回调
func execCb(ri *RetInfo) {
	defer func() { //延迟处理异常
		if r := recover(); r != nil {
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				log.Error("%v: %s", r, buf[:l])
			} else {
				log.Error("%v", r)
			}
		}
	}()

	// execute
	switch ri.cb.(type) { // 判断回调类型
	case func(error):               // 无返回值，只接收一个错误
		ri.cb.(func(error))(ri.err) // 执行回调
	case func(interface{}, error):                       // 一个返回值，一个错误
		ri.cb.(func(interface{}, error))(ri.ret, ri.err) // 执行回调
	case func([]interface{}, error):                               // 多个返回值，一个错误
		ri.cb.(func([]interface{}, error))(assert(ri.ret), ri.err) // 执行回调
	default:
		panic("bug")
	}
	return
}

// 执行rpc客户端实例中的回调函数
func (c *Client) Cb(ri *RetInfo) {
	c.pendingAsynCall--
	execCb(ri)
}

// 关闭rpc客户端，执行剩余异步调用
func (c *Client) Close() {
	for c.pendingAsynCall > 0 {
		c.Cb(<-c.ChanAsynRet)
	}
}

// 返回rpc客户端实例异步调用队列是否为空
func (c *Client) Idle() bool {
	return c.pendingAsynCall == 0
}
