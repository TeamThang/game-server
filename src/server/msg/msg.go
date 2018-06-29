package msg

import (
	"github.com/name5566/leaf/network/json"
	"server/msg/account"
	"reflect"
	"github.com/name5566/leaf/log"
)

// 使用默认的 JSON 消息处理器（默认还提供了 protobuf 消息处理器）
var Processor = json.NewProcessor()

func init() {
	// 注册Processor支持的msg
	Processor.Register(&Response{})
	registLogin()
}

// 标准返回数据
type Response struct {
	Status  int         `json:"status"`
	Message interface{} `json:"message"`
	Data    interface{} `json:"data"`
}

// 构造返回消息
// {msgID: {}}
func MakeResponse(msgID string, msgRes *Response) *map[string]interface{} {
	resMsg := make(map[string]interface{})
	resMsg[msgID] = map[string]interface{}{
		"status":  msgRes.Status,
		"message": msgRes.Message,
		"data":    msgRes.Data,
	}
	return &resMsg
}

// 获取msgID
func GetMsgID(msg interface{}) (string) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.Error("message %v not right struct", msg)
		return ""
	}
	msgID := msgType.Elem().Name()
	return msgID
}

func registLogin() {
	// 用户账户
	Processor.Register(&account.UserCreate{})
	Processor.Register(&account.UserQuery{})
	Processor.Register(&account.UserUpdate{})
	Processor.Register(&account.UserDelete{})
	// 密码修改
	Processor.Register(&account.PwdChange{})
	Processor.Register(&account.EmailRestPwdSend{})
	Processor.Register(&account.EmailRestPwdCheck{})
	// 用户权限
	Processor.Register(&account.RightCreate{})
	Processor.Register(&account.RightQuery{})
	Processor.Register(&account.RightUpdate{})
	Processor.Register(&account.RightDelete{})
	Processor.Register(&account.RightBind{})
	Processor.Register(&account.RightUnBind{})
	Processor.Register(&account.BindRightQuery{})
	// 登陆注销
	Processor.Register(&account.Login{})
	Processor.Register(&account.Logout{})
	Processor.Register(&account.GetUserInfo{})
	// 验证邮寄
	Processor.Register(&account.VerifyEmailSend{})
	Processor.Register(&account.VerifyEmailCheck{})
	// 提醒邮寄
	Processor.Register(&account.NotifyEmailCreate{})
	Processor.Register(&account.NotifyEmailQuery{})
	Processor.Register(&account.NotifyEmailSend{})
	Processor.Register(&account.NotifyEmailDelete{})
	Processor.Register(&account.NotifyEmailSub{})
	Processor.Register(&account.NotifyEmailUnSub{})
	// api key
	Processor.Register(&account.ApiKeyCreate{})
	Processor.Register(&account.ApiKeyDelete{})
	Processor.Register(&account.ApiKeyQuery{})
}
