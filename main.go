package main

import (
	chainmanger "chainCross/chainManger"
	"chainCross/config"
	"chainCross/console"
	"chainCross/p2p"
	"chainCross/rpc"
	"fmt"
)

func main() {
	start()
}

func start() {
	var p2pServer p2p.P2PServer
	var console console.Console
	chainManger := chainmanger.NewChainManger()

	conf := config.GetConfig()
	chainManger.AddChainByConfig(conf.Fabric)
	p2pServer.LoadConfig(conf)
	rpc.NewRPCServer(&p2pServer, chainManger, conf.Rpc.Port)

	go p2pServer.Start(chainManger)
	fmt.Println("p2pstoped")

	console.SetP2PServer(&p2pServer)
	console.Start()
}
