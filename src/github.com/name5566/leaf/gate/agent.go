package gate

import (
	"net"
)

type Agent interface {
	WriteMsg(msg interface{})     //发送消息
	LocalAddr() net.Addr          //本地地址
	RemoteAddr() net.Addr         //远端地址
	Close()                       //关闭代理
	Destroy()                     //销毁方法
	UserData() interface{}        //获取用户数据
	SetUserData(data interface{}) //设置用户数据
}
