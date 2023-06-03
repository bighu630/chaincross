package console

import (
	"bufio"
	"chainCross/p2p"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Console struct {
	p2pServer *p2p.P2PServer
	close     chan interface{}
}

var ConsoleKey []string = []string{"node"}

// SwithKeyAndExit 选择执行方案
func (c *Console) SwithKeyAndExit(keys []string) bool {
	key := keys[0]
	switch key {
	case "node":
		getNodeInfo(*c.p2pServer)
	case "exit":
		return true
	case "test":
		c.p2pServer.RemoteUnlock(keys[1], keys[2], keys[3])
	case "lock":
		rand.Seed(time.Now().Unix())
		num := strconv.Itoa(rand.Int())
		pri := sha256.Sum256([]byte(num))
		hash := sha256.Sum256(pri[:])
		fmt.Println("原像是:", hex.EncodeToString(pri[:]))
		fmt.Println("哈希是:", hex.EncodeToString(hash[:]))

	default:
		c.p2pServer.Say(key)
	}
	return false
}

func (c *Console) SetP2PServer(p *p2p.P2PServer) {
	c.p2pServer = p
}

func getNodeInfo(p p2p.P2PServer) {
	nodeAddr := p.NodeInfo()
	fmt.Println(nodeAddr[:len(nodeAddr)-1])
}

// 启动
func (c Console) Start() {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-c.close:
			return
		default:
			fmt.Print("> ")
			strWhthEnter, err := stdReader.ReadString('\n')
			str := strings.TrimRight(strWhthEnter, "\n")
			keys := strings.Split(str, " ")
			if err != nil {
				log.Println(err)
				return
			}
			if c.SwithKeyAndExit(keys) {
				return
			}
		}
	}
}
