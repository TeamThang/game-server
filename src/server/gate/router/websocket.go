package router

import (
	"server/msg"
	"server/login"
	"server/msg/account"
)

func init() {
	// 这里指定消息路由到对应的模块
	// 模块间使用 ChanRPC 通讯，消息路由也不例外
	initLogin()
}

func initLogin() {
	msg.Processor.SetRawHandler("UserCreate", login.HandleLogin)
	msg.Processor.SetRawHandler("UserQuery", login.HandleLogin)
	msg.Processor.SetRawHandler("UserUpdate", login.HandleLogin)
	msg.Processor.SetRawHandler("UserDelete", login.HandleLogin)
	msg.Processor.SetRawHandler("RightCreate", login.HandleLogin)
	msg.Processor.SetRawHandler("RightQuery", login.HandleLogin)
	msg.Processor.SetRawHandler("RightUpdate", login.HandleLogin)
	msg.Processor.SetRawHandler("RightDelete", login.HandleLogin)
	msg.Processor.SetRawHandler("RightBind", login.HandleLogin)
	msg.Processor.SetRawHandler("RightUnBind", login.HandleLogin)
	msg.Processor.SetRawHandler("BindRightQuery", login.HandleLogin)
	msg.Processor.SetRawHandler("Login", login.HandleLogin)
	msg.Processor.SetRawHandler("Logout", login.HandleLogin)
	msg.Processor.SetRawHandler("GetUserInfo", login.HandleLogin)
	msg.Processor.SetRawHandler("VerifyEmailSend", login.HandleLogin)
	msg.Processor.SetRawHandler("VerifyEmailCheck", login.HandleLogin)
	msg.Processor.SetRawHandler("NotifyEmailCreate", login.HandleLogin)
	msg.Processor.SetRawHandler("NotifyEmailQuery", login.HandleLogin)
	msg.Processor.SetRawHandler("NotifyEmailSend", login.HandleLogin)
	msg.Processor.SetRawHandler("NotifyEmailDelete", login.HandleLogin)
	msg.Processor.SetRawHandler("NotifyEmailSub", login.HandleLogin)
	msg.Processor.SetRawHandler("NotifyEmailUnSub", login.HandleLogin)
	msg.Processor.SetRawHandler("PwdChange", login.HandleLogin)
	msg.Processor.SetRawHandler("EmailRestPwdSend", login.HandleLogin)
	msg.Processor.SetRawHandler("EmailRestPwdCheck", login.HandleLogin)
	msg.Processor.SetRouter(&account.ApiKeyCreate{}, login.ChanRPC)
	msg.Processor.SetRouter(&account.ApiKeyQuery{}, login.ChanRPC)
	msg.Processor.SetRouter(&account.ApiKeyDelete{}, login.ChanRPC)

}