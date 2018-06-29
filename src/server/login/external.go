package login

import (
	"server/login/internal"
)

var (
	Module  = new(internal.Module)
	HandleLogin = internal.HandleLogin
	GetUserInfo = internal.GetUserInfo
	ChanRPC = internal.ChanRPC
)
