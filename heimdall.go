package main

import (
	"github.com/rafaelgfirmino/heimdall/configuration"
	"github.com/rafaelgfirmino/heimdall/gateway"
	"github.com/rafaelgfirmino/heimdall/server"
)

func main() {
	configuration.Load()
	gateway.Start()
	server.StartHeimdall()
}
