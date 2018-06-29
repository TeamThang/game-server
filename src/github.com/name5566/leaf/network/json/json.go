package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/log"
	"reflect"
)

// 处理器类型定义
type Processor struct {
	msgInfo map[string]*MsgInfo  //消息信息映射
}

// 消息信息类型定义
type MsgInfo struct {
	msgType       reflect.Type  //消息类型
	msgRouter     *chanrpc.Server  //处理消息的RPC服务器
	msgHandler    MsgHandler  //消息处理函数
	msgRawHandler MsgHandler  //原始消息处理函数
	// 处理消息有两种方式，一种的RPC服务器，一种是处理函数，可以同时处理
}

//消息处理函数类型定义
type MsgHandler func([]interface{})

type MsgRaw struct {
	msgID      string
	msgRawData json.RawMessage
}

//创建一个处理器
func NewProcessor() *Processor {
	p := new(Processor)
	p.msgInfo = make(map[string]*MsgInfo)
	return p
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
// 注册消息
func (p *Processor) Register(msg interface{}) string {
	msgType := reflect.TypeOf(msg)  //获取消息的类型
	if msgType == nil || msgType.Kind() != reflect.Ptr {  //判断消息的合法性，不能为空，需要是指针
		log.Fatal("json message pointer required")
	}
	msgID := msgType.Elem().Name()  //获取消息类型本身（不是指针）的名字，作为消息ID
	if msgID == "" {
		log.Fatal("unnamed json message")
	}
	if _, ok := p.msgInfo[msgID]; ok {  //判断消息是否已经注册
		log.Fatal("message %v is already registered", msgID)
	}

	i := new(MsgInfo)  //新建一个消息信息
	i.msgType = msgType  //保存消息类型
	p.msgInfo[msgID] = i  //保存消息信息到映射中
	return msgID
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
//设置路由
func (p *Processor) SetRouter(msg interface{}, msgRouter *chanrpc.Server) {
	msgType := reflect.TypeOf(msg)  //获取消息类型
	if msgType == nil || msgType.Kind() != reflect.Ptr {  //判断消息合法性
		log.Fatal("json message pointer required")
	}
	msgID := msgType.Elem().Name()  //获取消息类型本身的名字，也就是消息ID
	i, ok := p.msgInfo[msgID]  //根据消息ID获得消息信息
	if !ok {  //获取消息信息失败
		log.Fatal("message %v not registered", msgID)
	}

	i.msgRouter = msgRouter
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
//设置消息处理函数
func (p *Processor) SetHandler(msg interface{}, msgHandler MsgHandler) {
	msgType := reflect.TypeOf(msg)  //消息类型
	if msgType == nil || msgType.Kind() != reflect.Ptr {  //判断合法性
		log.Fatal("json message pointer required")
	}
	msgID := msgType.Elem().Name()  //获取消息ID
	i, ok := p.msgInfo[msgID]  //获取消息信息
	if !ok {
		log.Fatal("message %v not registered", msgID)
	}

	i.msgHandler = msgHandler  //保存消息处理函数
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
// 原始消息处理函数
func (p *Processor)SetRawHandler(msgID string, msgRawHandler MsgHandler) {
	i, ok := p.msgInfo[msgID]
	if !ok {
		log.Fatal("message %v not registered", msgID)
	}

	i.msgRawHandler = msgRawHandler
}

// goroutine safe
//路由
func (p *Processor) Route(msg interface{}, userData interface{}) error {
	// raw
	if msgRaw, ok := msg.(MsgRaw); ok {
		i, ok := p.msgInfo[msgRaw.msgID]
		if !ok {
			return fmt.Errorf("message %v not registered", msgRaw.msgID)
		}
		if i.msgRawHandler != nil {
			i.msgRawHandler([]interface{}{msgRaw.msgID, msgRaw.msgRawData, userData})
		}
		return nil
	}

	// json
	msgType := reflect.TypeOf(msg)   //获取消息类型
	if msgType == nil || msgType.Kind() != reflect.Ptr {  //判断合法性
		return errors.New("json message pointer required")
	}
	msgID := msgType.Elem().Name()  //获取消息ID
	i, ok := p.msgInfo[msgID]  //获取消息信息
	if !ok {  //获取失败
		return fmt.Errorf("message %v not registered", msgID)
	}
	if i.msgHandler != nil {  //调用消息处理函数
		i.msgHandler([]interface{}{msg, userData})
	}
	if i.msgRouter != nil {  //调用RPC服务器
		i.msgRouter.Go(msgType, msg, userData)  //rpc服务器自己发起调用
	}
	return nil
}

// goroutine safe
//解码消息
func (p *Processor) Unmarshal(data []byte) (interface{}, error) {
	var m map[string]json.RawMessage  //存储解码数据。RawMessage is a raw encoded JSON object,used to delay JSON decoding
	fmt.Printf("unmarshal1: %s, %v\n", data, m)
	err := json.Unmarshal(data, &m)  //解码
	if err != nil {
		return nil, err
	}
	if len(m) != 1 {  //m的长度必为1，也就是只有一个key value
		return nil, errors.New("invalid json data")
	}

	for msgID, data := range m {  //取出msgID和未解码的data
		fmt.Printf("msgID: %v, data: %s\n" ,msgID, string(data))
		i, ok := p.msgInfo[msgID]  //取出消息信息
		if !ok {
			return nil, fmt.Errorf("message %v not registered", msgID)
		}

		// msg
		if i.msgRawHandler != nil {
			return MsgRaw{msgID, data}, nil
		} else {
			msg := reflect.New(i.msgType.Elem()).Interface()  //存储解码数据，msgType本身为一个Ptr
			fmt.Printf("unmarshal2: %s, %v\n", data, msg)
			return msg, json.Unmarshal(data, msg)  //解码data
		}
	}

	panic("bug")
}

// goroutine safe
// 编码消息
// 增加字典map[string]interface{}直接返回
func (p *Processor) Marshal(msg interface{}) ([][]byte, error) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		return nil, errors.New("json message pointer required")
	}
	msgID := msgType.Elem().Name()
	if msgID == "" {  // 如果msgID是空，判断是否是map[string]interface{}
		if msgDict, ok := msg.(*map[string]interface{}); ok{
			if len(*msgDict) != 1 {
				return nil, fmt.Errorf("message %v not registered", msgID)
			}
			for k,v := range *msgDict {
				msgID = k
				msg = v
				break
			}
		}
	}
	//fmt.Printf("marshal msgID: %s, msgInfo: %v\n", msgID, p.msgInfo)
	if _, ok := p.msgInfo[msgID]; !ok {
		return nil, fmt.Errorf("message %v not registered", msgID)
	}

	// data
	m := map[string]interface{}{msgID: msg}
	data, err := json.Marshal(m)
	return [][]byte{data}, err
}
