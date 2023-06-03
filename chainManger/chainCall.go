package chainmanger

import (
	"chainCross/dao"
	"encoding/hex"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

// 链调用数据
type callData struct {
	fc   string
	args [][]byte
}

// 设置链调用数据
func (c *Chain) SetCallData(fc string, args [][]byte) {
	c.callData.fc = fc
	c.callData.args = [][]byte{}
	for _, s := range args {
		c.callData.args = append(c.callData.args, s)
	}
}

// 调用链码（写入）
func (c Chain) InvokeCC() (*channel.Response, error) {
	req := channel.Request{
		ChaincodeID: "nft_last",
		Fcn:         c.callData.fc,
		Args:        c.callData.args,
	}
	reqPeers := channel.WithTargetEndpoints(c.peers...)
	resp, err := c.cli.Cc.Execute(req, reqPeers)
	log.Printf("Invoke chaincode response:\n"+
		"id: %v\nvalidate: %v\nchaincode status: %v\n\n",
		resp.TransactionID,
		resp.TxValidationCode,
		resp.ChaincodeStatus)
	if err != nil {
		log.Printf("Invoke chaincode error %v ", err)
		return nil, err
	}
	return &resp, nil
}

func (c *Chain) StartHTCL(tx dao.HTCLTx) (*channel.Response, error) {
	type info struct {
		LtokenId string
		Lto      string
		Cnum     float64
		Cto      string
	}
	in := info{
		LtokenId: tx.TradeNFTID,
		Lto:      tx.NFTRecipientAddr,
		Cnum:     tx.CoinNUM,
		Cto:      tx.CoinRecipientAddr,
	}

	byteIn, err := json.Marshal(&in)
	if err != nil {
		log.Printf("序列化 出错%v", err)
	}
	timestart := strconv.FormatInt(tx.TimeStart, 10)
	c.SetCallData("HTCLLock", [][]byte{byteIn, []byte(tx.Hs), []byte(timestart)})
	resp, err := c.InvokeCC()
	return resp, err
}

func (c *Chain) UnlockHTLC(nftID, s string) (*channel.Response, error) {
	bytes, _ := hex.DecodeString(s)
	c.SetCallData("HTCLUnlock", [][]byte{[]byte(nftID), bytes})
	return c.InvokeCC()
}

// 查询
func (c Chain) QueryCC() {
	req := channel.Request{
		ChaincodeID: "mycc",
		Fcn:         c.callData.fc,
		Args:        c.callData.args,
	}

	//随即选择一个peer查询
	reqPeer := channel.WithTargetEndpoints(c.RangPeer())

	resp, err := c.cli.Cc.Query(req, reqPeer)
	if err != nil {
		log.Printf("Query chaincode error %v ", err)
	}

	log.Printf("Query chaincode tx response:\ntx: %s\nresult: %v\n\n",
		resp.TransactionID,
		string(resp.Payload))
}
