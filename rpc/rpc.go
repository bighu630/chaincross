package rpc

import (
	chainmanger "chainCross/chainManger"
	"chainCross/p2p"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type RPCServer struct {
	p2pServer   *p2p.P2PServer
	chainManger *chainmanger.ChainManager
	rpcHandler  *RpcHandler
}

func NewRPCServer(p2p *p2p.P2PServer, chain *chainmanger.ChainManager, port int) {
	server := RPCServer{
		p2pServer:   p2p,
		chainManger: chain,
	}
	go server.startServer(port)
}

func (r *RPCServer) startServer(port int) {
	server := rpc.NewServer()

	p := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", p)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	rpc := RpcHandler{
		p2pServer:   r.p2pServer,
		chainManger: r.chainManger,
	}
	err = server.Register(&rpc)
	if err != nil {
		log.Println(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		jsonrpconn := jsonrpc.NewServerCodec(conn)
		server.ServeCodec(jsonrpconn)
	}
}
