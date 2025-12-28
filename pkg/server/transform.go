package server

import (
	"aegis/pkg/apimodel"
	"aegis/pkg/events"
	"aegis/pkg/frontend"
)


func ExecToFrontend(ev events.ExecEvent) apimodel.ExecEvent {
	return frontend.ExecToFrontend(ev)
}


func FileToFrontend(ev events.FileOpenEvent, filename string) apimodel.FileEvent {
	return frontend.FileToFrontend(ev, filename)
}


func ConnectToFrontend(ev events.ConnectEvent, addr string, processName string) apimodel.ConnectEvent {
	return frontend.ConnectToFrontend(ev, addr, processName)
}
