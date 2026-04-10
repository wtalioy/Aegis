package events

type EventHandler interface {
	HandleExec(ev ExecEvent)
	HandleFileOpen(ev FileOpenEvent, filename string)
	HandleConnect(ev ConnectEvent)
}

type HandlerChain struct {
	handlers []EventHandler
}

func NewHandlerChain(handlers ...EventHandler) *HandlerChain {
	return &HandlerChain{handlers: handlers}
}

func (c *HandlerChain) Add(h EventHandler) {
	c.handlers = append(c.handlers, h)
}

func (c *HandlerChain) HandleExec(ev ExecEvent) {
	for _, h := range c.handlers {
		h.HandleExec(ev)
	}
}

func (c *HandlerChain) HandleFileOpen(ev FileOpenEvent, filename string) {
	for _, h := range c.handlers {
		h.HandleFileOpen(ev, filename)
	}
}

func (c *HandlerChain) HandleConnect(ev ConnectEvent) {
	for _, h := range c.handlers {
		h.HandleConnect(ev)
	}
}
