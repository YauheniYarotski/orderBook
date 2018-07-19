package orderBook

import "orderBook/core"


type Service struct {
	Manager   *core.Manager
}

func (self *Service) InitExchanger(conf core.ManagerConfiguration) {
	var manager = core.NewManager()
	go manager.Start(conf)
	self.Manager = manager
}
