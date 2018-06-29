package internal

import (
	"github.com/name5566/leaf/gate"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
}

// agent 被创建时
func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	_ = a
}

// agent 被关闭时
func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	_ = a
}
