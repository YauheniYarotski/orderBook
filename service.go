package orderBook

import "orderBook/core"


type Service struct {
	Manager   *core.Manager
}

func (e *Service) InitExchanger(conf core.ManagerConfiguration) {
	var manager = core.NewManager()
	go manager.StartListen(conf)
	e.Manager = manager
}
