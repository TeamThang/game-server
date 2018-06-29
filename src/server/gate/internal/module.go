package internal

import (
	"github.com/name5566/leaf/gate"
	"server/conf"
	"server/game"
	"server/msg"
	self "server/gate/http"
)

type Module struct {
	*gate.Gate
}

func (m *Module) OnInit() {
	m.Gate = &gate.Gate{
		MaxConnNum:      conf.Server.MaxConnNum,
		PendingWriteNum: conf.PendingWriteNum,
		MaxMsgLen:       conf.MaxMsgLen,
		WSAddr:          conf.Server.WSAddr,
		HTTPTimeout:     conf.HTTPTimeout,
		CertFile:        conf.Server.CertFile,
		KeyFile:         conf.Server.KeyFile,
		TCPAddr:         conf.Server.TCPAddr,
		LenMsgLen:       conf.LenMsgLen,
		LittleEndian:    conf.LittleEndian,
		Processor:       msg.Processor,
		AgentChanRPC:    game.ChanRPC,
		HTTPAddr:        conf.Server.HTTPAddr,
		HTTPCertFile:    conf.Server.HTTPCertFile,
		HTTPKeyFile:     conf.Server.HTTPKeyFile,
		ServeMux:        *self.HttpServeMux,
	}
}
