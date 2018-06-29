// 登陆消息结构
package account

import "encoding/json"

// 注册登陆服务消息定义
// 默认为raw json, 直接装发注册服务
type UserCreate json.RawMessage
type UserQuery json.RawMessage
type UserUpdate json.RawMessage
type UserDelete json.RawMessage
type RightCreate json.RawMessage
type RightQuery json.RawMessage
type RightUpdate json.RawMessage
type RightDelete json.RawMessage
type RightBind json.RawMessage
type RightUnBind json.RawMessage
type BindRightQuery json.RawMessage
type Login json.RawMessage
type Logout json.RawMessage
type GetUserInfo json.RawMessage
type VerifyEmailSend json.RawMessage
type VerifyEmailCheck json.RawMessage
type NotifyEmailCreate json.RawMessage
type NotifyEmailQuery json.RawMessage
type NotifyEmailSend json.RawMessage
type NotifyEmailDelete json.RawMessage
type NotifyEmailSub json.RawMessage
type NotifyEmailUnSub json.RawMessage
type PwdChange json.RawMessage
type EmailRestPwdSend json.RawMessage
type EmailRestPwdCheck json.RawMessage

type ApiKeyCreate struct {
}

type ApiKeyDelete struct {
	AccessKey string
}

type ApiKeyQuery struct {
}
