package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

type Args struct {
	A, B int
}

type PreHTCL struct {
	ChainAName        string  `json:"chainAName"`
	CHainBName        string  `json:"chainBName"`
	TradeNFTID        string  `json:"tradeNFTID"`        // 交易NFTID
	NFTRecipientAddr  string  `json:"NFTRecipientAddr"`  // NFT接受者地址
	CoinNUM           float64 `json:"coinNUM"`           // 代币数量
	CoinRecipientAddr string  `json:"coinRecipientAddr"` // 代币接受者地址
	Hs                string  `json:"hs"`                // 哈希时间锁用到的Hash(S)
	TimeInterval      int     `json:"timeInterval"`      // 时间间隔 分钟
}

type HTCLTx struct {
	ChainAName        string
	TradeNFTID        string  // 交易NFTID
	NFTRecipientAddr  string  // NFT接受者地址
	ChainBName        string  // 交易链名称
	CoinNUM           float64 // 代币数量
	CoinRecipientAddr string  // 代币接受者地址
	Hs                string  // 哈希时间锁用到的Hash(S)
	TimeStart         int64   // 开始时间戳
	TimeEnd           int64   // 结束时间戳
}

func main() {
	conn, err := net.Dial("tcp", ":5566")
	if err != nil {
		log.Fatal("dial error:", err)
	}

	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))

	args := &PreHTCL{
		ChainAName:        "fabric_NFT",
		CHainBName:        "my_etc",
		TradeNFTID:        "00001",
		NFTRecipientAddr:  "user2",
		CoinNUM:           12.123,
		CoinRecipientAddr: "chainbUser2",
		Hs:                "test",
		TimeInterval:      5,
	}
	var reply string
	fmt.Println("测试函数 RpcHandler.Test，输出序列化后的结构")
	err = client.Call("RpcHandler.Test", args, &reply)
	if err != nil {
		log.Fatal("Multiply error:", err)
	}
	fmt.Printf("序列化后的结果为: %s\n\n", reply)
	time.Sleep(1 * time.Second)

	var txid int64
	fmt.Printf("工作函数RpcHandler.CreateHTCLTxByPreInfo，生成HTCL交易提案\n")
	err = client.Call("RpcHandler.CreateHTCLTxByPreInfo", args, &txid)
	if err != nil {
		log.Fatal("Multiply error:", err)
	}
	fmt.Printf("生成的交易ID为:%d\n\n", txid)
	time.Sleep(1 * time.Second)

	var tx HTCLTx
	fmt.Printf("工作函数GetHTCLTxByID,根据交易ID获取交易信息\n")
	err = client.Call("RpcHandler.GetHTCLTxByID", txid, &tx)
	if err != nil {
		log.Fatal("Multiply error:", err)
	}
	fmt.Printf("获取到的交易信息为:%v\n\n", tx)

	fmt.Printf("工作函数StartHTCL,根据交易ID执行交易,此处使用不存在的id使之返回错误\n")
	err = client.Call("RpcHandler.StartHTCL", int64(123), &tx)
	if err != nil {
		log.Fatal("Multiply error:", err)
	}
}
