package p2p

import (
	"bufio"
	chainmanger "chainCross/chainManger"
	"chainCross/config"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/multiformats/go-multiaddr"
)

type P2PServer struct {
	host        host.Host
	conf        *config.Config
	addr        string
	chainManger *chainmanger.ChainManager
}

var sendmsg = make(chan []byte, 1024)
var handlemsg = make(chan []byte, 1024)
var stop chan interface{}

// Start 启动服务，并保持运行
func (p *P2PServer) Start(chain *chainmanger.ChainManager) {

	p.chainManger = chain
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 加载配置文件
	r := rand.Reader
	// 生成主机
	var err error
	p.host, err = makeHost(p.conf.Peer.Port, r)
	add := fmt.Sprintln(p.host.Addrs()[0])
	p.addr = fmt.Sprintf("%s/p2p/%s", add[:len(add)-1], fmt.Sprintln(p.host.ID()))

	if err != nil {
		log.Println(err)
		return
	}
	if p.conf.Peer.Dest == "" {
		go startPeer(p, handleStream)
	} else {
		rw := startWithOtherPeer(ctx, p)

		go writeData(rw)
		go readData(rw)
	}
	go p.handler()
	select {}
}

func (p *P2PServer) handler() {
	for {
		data := <-handlemsg
		var msg Msg
		err := json.Unmarshal(data, &msg)
		if err != nil {
			log.Println("无法将数据还原为msg")
		}
		switch msg.Type {
		case MSG:
			fmt.Println(string(msg.Data))
		case UNLOCK:
			var unlock RemoteUnlock
			err := json.Unmarshal(msg.Data, &unlock)
			if err != nil {
				log.Println("无法将数据还原为unlock")
			}
			chain := p.chainManger.GetChainByName(unlock.ChainID)
			if chain == nil {
				log.Println("没有找到对应链" + unlock.ChainID)
				return
			}
			chain.UnlockHTLC(unlock.LockID, unlock.S)
		case LOCK:
			var lock RemoteLock
			err := json.Unmarshal(msg.Data, &lock)
			if err != nil {
				log.Println("无法将数据还原为lock")
			}
			chain := p.chainManger.GetChainByName(lock.ChainID)
			if chain == nil {
				log.Println("没有找到对应链" + lock.ChainID)
				return
			}
			chain.UnlockHTLC(lock.AproveID, lock.Hs)
		}
	}
}

// LoadConfig 加载配置信息
func (p *P2PServer) LoadConfig(conf *config.Config) {
	p.conf = conf
}

// Stop 关闭
func (p *P2PServer) Stop() {
	stop <- new(interface{})
	p.host.Close()
}

// SendMsg 发送信息，对外开放
func SendMsg(data []byte) {
	sendmsg <- data
}

// NodeInfo 返回节点信息 /ip4/127.0.0.1/tcp/port/nodeID
func (p P2PServer) NodeInfo() string {
	return p.addr
}

// 处理p2p数据流
func handleStream(s network.Stream) {

	log.Println("Got a new stream!")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)
}

func (p *P2PServer) RemoteUnlock(chainID, aproveID string, s string) {
	remoteData, err := json.Marshal(&RemoteUnlock{
		ChainID: chainID,
		LockID:  aproveID,
		S:       s,
	})
	if err != nil {
		log.Println("无法格式化remoteLock")
	}
	msg := Msg{
		Type: UNLOCK,
		Data: remoteData,
	}
	byteMsg, err := json.Marshal(&msg)
	SendMsg(byteMsg)
}

func (p *P2PServer) RemoteLock(chainID, aproveID, Hs string) {
	remoteLock := RemoteLock{
		ChainID:  chainID,
		AproveID: aproveID,
		Hs:       Hs,
	}
	data, err := json.Marshal(remoteLock)
	if err != nil {
		log.Println("无法格式化remoteLock")
	}
	msg := Msg{
		Type: LOCK,
		Data: data,
	}
	byteMsg, err := json.Marshal(&msg)
	if err != nil {
		log.Panicln("无法格式化byteMsg")
	}
	SendMsg(byteMsg)
}

func (p *P2PServer) Unlocked(id int64) {

}

func (p *P2PServer) Say(str string) {
	msg := Msg{
		Type: MSG,
		Data: []byte(str),
	}
	byteMsg, err := json.Marshal(&msg)
	if err != nil {
		log.Panicln("无法序列化用户发送的信息")
	}
	SendMsg(byteMsg)
}

// 读取数据
func readData(rw *bufio.ReadWriter) {
	for {
		data, err := rw.ReadBytes('\n')
		if err != nil {
			return
		}
		if data == nil {
			return
		}
		if len(data) != 1 {
			handlemsg <- data
		}
	}
}

// 写入处理
func writeData(rw *bufio.ReadWriter) {
	for {
		select {
		case data := <-sendmsg:
			data = append(data, '\n')
			_, err := rw.Write(data)
			if err != nil {
				fmt.Println(err)
			}
			err = rw.Flush()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

// 构造节点
func makeHost(listenPort int, reader io.Reader) (host.Host, error) {

	privKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, reader)
	if err != nil {
		return nil, fmt.Errorf("faild to creat node privKey %v ", err)
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(privKey),
	}
	host, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("faild to New a host: %v", err)
	}

	return host, nil
}

// 启动节点
func startPeer(p2p *P2PServer, streamHandler network.StreamHandler) {
	// 没有其他节点的情况
	p2p.host.SetStreamHandler("/cross/1.0.0", streamHandler)

	var port string
	for _, la := range p2p.host.Network().ListenAddresses() {
		if po, err := la.ValueForProtocol(multiaddr.P_TCP); err == nil {
			port = po
			break
		}
	}

	if port == "" {
		log.Panicln("was not able to find actual local port")
		return
	}
	log.Println("start success")
}

// 以连接其他节点的方式启动
func startWithOtherPeer(ctx context.Context, p2p *P2PServer) *bufio.ReadWriter {
	log.Println("This node's multiaddresses:")
	for _, la := range p2p.host.Addrs() {
		log.Printf(" - %v\n", la)
	}
	maddr, err := multiaddr.NewMultiaddr(p2p.conf.Peer.Dest)
	if err != nil {
		log.Println(err)
		return nil
	}
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Println(err)
		return nil
	}

	p2p.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	s, err := p2p.host.NewStream(ctx, info.ID, "/cross/1.0.0")
	if err != nil {
		fmt.Println(err)
		log.Println(err)
	}

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	return rw
}
