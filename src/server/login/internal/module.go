package internal

import (
	"github.com/name5566/leaf/module"
	"server/base"
	"github.com/name5566/leaf/gate"
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/gate/user"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton
}

func (m *Module) OnDestroy() {

}

// 检查当前连接关联Agent的用户
// 获取Agent, 通过Agent.UserData查找UserID, Token
func GetUserInfo(arg interface{}) (gate.Agent, uint, string, error) {
	a, ok := arg.(gate.Agent)
	if !ok {
		errMsg := fmt.Sprintf("args %v is not right Agent", a)
		log.Error(errMsg)
		return nil, 0, "", fmt.Errorf(errMsg)
	}
	userData1 := a.UserData()
	if userData1 == nil {
		return a, 0, "", fmt.Errorf("unlogin")
	}
	userData, ok := userData1.(user.UserData)
	if !ok {
		errMsg := fmt.Sprintf("user data %v is not valid:  ", userData1)
		log.Error(errMsg)
		return a, 0, "", fmt.Errorf(errMsg)
	}
	return a, userData.UserID, userData.Token, nil
}
