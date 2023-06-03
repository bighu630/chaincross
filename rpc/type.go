package rpc

type PreHTCL struct {
	ChainAName        string  `json:"chainAName"`
	CHainBName        string  `json:"chainBName"`
	TradeNFTID        string  `json:"tradeNFTID"`        //交易NFTID
	NFTRecipientAddr  string  `json:"NFTRecipientAddr"`  // NFT接受者地址
	CoinNUM           float64 `json:"coinNUM"`           //代币数量
	CoinRecipientAddr string  `json:"coinRecipientAddr"` // 代币接受者地址
	Hs                string  `json:"hs"`                //哈希时间锁用到的Hash(S)
	TimeInterval      int     `json:"timeInterval"`      // 时间间隔 分钟
	AproveID          string  `json:"aproveid"`
}

type UnlockHTCL struct {
	TransaID int64
	S        string
}
