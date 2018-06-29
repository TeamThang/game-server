package network

type Agent interface {
	OnInit(data interface{})
	Run()
	OnClose()
}
