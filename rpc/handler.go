package rpc

import (
	chainmanger "chainCross/chainManger"
	"chainCross/dao"
	"chainCross/p2p"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type RpcHandler struct {
	chainManger *chainmanger.ChainManager
	p2pServer   *p2p.P2PServer
}

func (h *RpcHandler) CreateHTCLTxByPreInfo(PreInfo *PreHTCL, resout *int64) error {
	Tx := dao.HTLCTx{
		ChainAName:        PreInfo.ChainAName,
		ChainBName:        PreInfo.CHainBName,
		TradeNFTID:        PreInfo.TradeNFTID,
		NFTRecipientAddr:  PreInfo.NFTRecipientAddr,
		CoinNUM:           PreInfo.CoinNUM,
		CoinRecipientAddr: PreInfo.CoinRecipientAddr,
		Hs:                PreInfo.Hs,
		AproveID:          PreInfo.AproveID,
		TimeStart:         time.Now().Unix(),
		TimeEnd:           time.Now().Unix() + int64(PreInfo.TimeInterval*60*1000),
	}

	res, err := dao.CreateHTLCTx(Tx)
	*resout = res
	return err
}

func (h *RpcHandler) GetHTCLTxByID(id int64, resout *dao.HTLCTx) error {
	res, err := dao.GetHTLCTx(id)
	*resout = res
	return err
}

func (h *RpcHandler) StartHTLC(id int64, resout *string) error {
	tx, err := dao.GetHTLCTx(id)
	if err != nil {
		log.Println(err)
		*resout = err.Error()
		return err
	}
	err = h.handlerHTCL(tx)
	if err != nil {
		log.Println(err)
		*resout = err.Error()
		return err
	}
	*resout = "HTCL启动成功"
	return nil
}

type UnlockArgs struct {
	Id int64
	S  string
}

func (h *RpcHandler) UnlockHTCL(UA *UnlockArgs, resout *string) error {
	unlock := UnlockHTCL{
		TransaID: UA.Id,
		S:        UA.S,
	}
	h.handlerHTCLUnlock(unlock)
	*resout = ""
	return nil
}

func (h *RpcHandler) Test(PreInfo *PreHTCL, resout *string) error {
	Tx := dao.HTLCTx{
		TradeNFTID:        PreInfo.TradeNFTID,
		NFTRecipientAddr:  PreInfo.NFTRecipientAddr,
		CoinNUM:           PreInfo.CoinNUM,
		CoinRecipientAddr: PreInfo.CoinRecipientAddr,
		Hs:                PreInfo.Hs,
		TimeStart:         time.Now().Unix(),
		TimeEnd:           time.Now().Unix() + int64(PreInfo.TimeInterval*60*1000),
	}
	data, err := json.Marshal(&Tx)
	*resout = string(data)
	return err
}

func (r *RpcHandler) handlerHTCL(tx dao.HTLCTx) error {
	if r.chainManger.IsUseableChain(tx.ChainAName) {
		chain := r.chainManger.GetChainByName(tx.ChainAName)
		if chain == nil {
			return fmt.Errorf("没有找到名为%s的链", tx.ChainAName)
		}
		r.p2pServer.RemoteLock(tx.ChainBName, tx.AproveID, tx.Hs)
		_, err := chain.StartHTCL(tx)
		return err
	}
	return nil
}

func (r *RpcHandler) handlerHTCLUnlock(unlock UnlockHTCL) error {
	HTLCtx, err := dao.GetHTLCTx(unlock.TransaID)
	if err != nil {
		log.Println(err)
	}
	chain := r.chainManger.GetChainByName(HTLCtx.ChainAName)
	if chain == nil {
		return fmt.Errorf("没有找到名为%s的链", HTLCtx.ChainAName)
	}
	r.p2pServer.RemoteUnlock(HTLCtx.ChainBName, HTLCtx.AproveID, unlock.S)

	_, err = chain.UnlockHTLC(HTLCtx.TradeNFTID, unlock.S)
	return err
}
