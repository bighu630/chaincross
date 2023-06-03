package chainmanger

import (
	conf "chainCross/config"
	"crypto/rand"
	"log"
	"math/big"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type Chain struct {
	name     string
	cli      *client
	peers    []string
	callData callData
}

type client struct {
	// sdk clients
	SDK *fabsdk.FabricSDK
	Rc  *resmgmt.Client
	Cc  *channel.Client
}

// 可能需要优化这里的chain初始化方式，如果换一个chain,目前来看需要修改config.go,最好是不要修改这个文件
func NewChainClient(conf conf.Fabric) *Chain {
	sdk, err := fabsdk.New(config.FromFile(conf.ConfigPath))
	if err != nil {
		log.Panicf("failed to create fabric sdk: %s", err)
	}
	log.Println("Initialized fabric sdk")

	rc, cc := newSdkClient(sdk, conf.ChannelID, conf.OrgName, conf.OrgAdmin, conf.OrgUser)
	cli := client{
		SDK: sdk,
		Rc:  rc,
		Cc:  cc,
	}
	return &Chain{
		name:  conf.Name,
		cli:   &cli,
		peers: conf.Peers,
	}
}

// newSdkClient create resource client and channel client
func newSdkClient(sdk *fabsdk.FabricSDK, channelID, orgName, orgAdmin, OrgUser string) (rc *resmgmt.Client, cc *channel.Client) {
	var err error

	// create rc
	rcp := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))
	rc, err = resmgmt.New(rcp)
	if err != nil {
		log.Panicf("failed to create resource client: %s", err)
	}
	log.Println("Initialized resource client")

	// create cc
	ccp := sdk.ChannelContext(channelID, fabsdk.WithUser(OrgUser))
	cc, err = channel.New(ccp)
	if err != nil {
		log.Printf("failed to create channel client: %s", err)
	}
	log.Println("Initialized channel client")

	return rc, cc
}

// RegisterChaincodeEvent more easy than event client to registering chaincode event.
func (c *Chain) RegisterChaincodeEvent(ccid, eventName string) (fab.Registration, <-chan *fab.CCEvent, error) {
	return c.cli.Cc.RegisterChaincodeEvent(ccid, eventName)
}

func (c Chain) RangPeer() string {
	rang, err := rand.Int(rand.Reader, big.NewInt(1024))
	if err != nil {
		log.Printf("failed to get an rang int %v ", err)
	}

	it := rang.Int64() % int64(len(c.peers))

	return c.peers[it]
}
