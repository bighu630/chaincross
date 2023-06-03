func (n *NFT_HTCL) ChangOwner(ctx contractapi.TransactionContextInterface, tokenId string, to string) error {
	//检测账本是否初始化
	hasInit, err := checkInitialized(ctx)
	isExists := _nftExists(ctx, tokenId)
	//确认上链资产的所有人
	clientID, err := ctx.GetClientIdentity().GetID()
	sender := clientID
	owner, err := n.OwnerOf(ctx, tokenId)
	// 如果在HTCL锁里面就无法交易
	hkey, err := ctx.GetStub().CreateCompositeKey(HTCLKEY, []string{tokenId})
	respose, err := ctx.GetStub().GetState(hkey)
	// 删除资产的交易锁
	tkey, err := ctx.GetStub().CreateCompositeKey(TKEY, []string{tokenId})
	respose, err = ctx.GetStub().GetState(tkey)
	if respose != nil {
		err = ctx.GetStub().DelState(tkey)
		if err != nil {
			return fmt.Errorf("删除数据失败 :%v ", err)
		}
	}
	// 获取NFT,修改拥有者
	token, err := _readNFT(ctx, tokenId)
	token.Owner = to
	cnftKey, err := ctx.GetStub().CreateCompositeKey(NFTKEY, []string{tokenId})
	byteToken, err := json.Marshal(token)
	err = ctx.GetStub().PutState(cnftKey, byteToken)
	return nil
}


// 锁定交易
func (n *NFT) LockTranscation(ctx contractapi.TransactionContextInterface, transcationInfo string, hash string) (string, error) {
	//检测账本是否初始化
	hasInit, err := checkInitialized(ctx)
	var LockInfo info
	err = json.Unmarshal([]byte(transcationInfo), &LockInfo)
	if exist := _nftExists(ctx, LockInfo.LtokenId); !exist {}
	//确认上链资产的所有人
	clientID, err := ctx.GetClientIdentity().GetID()
	sender := clientID
	owner, err := n.OwnerOf(ctx, LockInfo.LtokenId)
	if owner != sender {
		return "", fmt.Errorf("just token Owner can do it ")
	}
	infoHash := sha256.Sum256([]byte(transcationInfo))
	if string(infoHash[:]) != hash {}
	tkey, err := ctx.GetStub().CreateCompositeKey(TKEY, []string{LockInfo.LtokenId})
	err = ctx.GetStub().PutState(tkey, []byte(hash))
	return hash, nil
}

func (n *NFT) HTCLLock(ctx contractapi.TransactionContextInterface, transcationInfo string, HTCLHash string, time int64) (string, error) {
	//检测账本是否初始化
	hasInit, err := checkInitialized(ctx)
	var Info info
	err = json.Unmarshal([]byte(transcationInfo), &Info)
	if exist := _nftExists(ctx, Info.LtokenId); !exist {}
	tkey, err := ctx.GetStub().CreateCompositeKey(TKEY, []string{Info.LtokenId})
	//获取链上授权信息，如果没有授权将无法执行
	thash, err := ctx.GetStub().GetState(tkey)
	var h [32]byte
	copy(h[:], thash)
	infoHash := sha256.Sum256([]byte(transcationInfo))
	if infoHash != h {}
	lockinfo := LockInfo{
		TokenId: Info.LtokenId,
		To:      Info.Lto,
		Hash:    HTCLHash,
		Time:    time,
	}
	hKey, err := ctx.GetStub().CreateCompositeKey(HTCLKEY, []string{Info.LtokenId})
	byteInfo, err := json.Marshal(lockinfo)
	err = ctx.GetStub().PutState(hKey, byteInfo)
	return "", nil
}



services:
  NFT:
    image: hyperledger/fabric-ca:latest 
    labels:
      service: hyperledger-fabric
    environment:
      - FABRIC_CA_HOME=/etc/hyperledger/fabric-ca-server  
      - FABRIC_CA_SERVER_CA_NAME=ca-NFT
      - FABRIC_CA_SERVER_TLS_ENABLED=true
       - FABRIC_CA_SERVER_PORT=7054
       - FABRIC_CA_SERVER_OPERATIONS_LISTENADDRESS=0.0.0.0:17054
     ports:
       - "7054:7054"
       - "17054:17054"
     command: sh -c 'fabric-ca-server start -b admin:adminpw -d'
     volumes:
       - ../ca_conf/NFT:/etc/hyperledger/fabric-ca-server
     container_name: ca_NFT
     networks:
       - test



services:
   peer0.Peer_fish.com:
     container_name: peer0.Peer_fish.com
     image: hyperledger/fabric-peer:2.2
     labels:
       service: hyperledger-fabric
     environment:
       - FABRIC_CFG_PATH=/etc/hyperledger/peercfg
       - CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/fabric/tls/server.crt
       - CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/fabric/tls/server.key
       - CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/fabric/tls/ca.crt
       - CORE_PEER_LOCALMSPID=Org1MSP
       - CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/fabric/msp
       - CORE_OPERATIONS_LISTENADDRESS=peer0.Peer_fish.com:9444
     volumes:
         - ../org_crypto_conf/organizations/peerOrganizations/Peer_fish.com/peers/peer0.Peer_fish.com:/etc/hyperledger/fabric
         - peer0.Peer_fish.com:/var/hyperledger/production
         - ./peercfg:/etc/hyperledger/peercfg
         - /var/run/docker.sock:/host/var/run/docker.sock
     working_dir: /root
     command: peer node start
     ports:
       - 7051:7051
       - 9444:9444
     networks:
       - test


services:

  orderer.fish.com:
    container_name: orderer.fish.com
    image: hyperledger/fabric-orderer:2.2
    labels:
      service: hyperledger-fabric
    environment:
      - FABRIC_LOGGING_SPEC=DEBUG
      - ORDERER_GENERAL_LISTENADDRESS=0.0.0.0
      - ORDERER_GENERAL_LISTENPORT=7050
      - ORDERER_GENERAL_LOCALMSPID=OrdererMSP
      - ORDERER_GENERAL_LOCALMSPDIR=/var/hyperledger/orderer/msp
      - ORDERER_GENERAL_BOOTSTRAPMETHOD=file
      - ORDERER_GENERAL_BOOTSTRAPFILE=/var/hyperledger/orderer/orderer.genesis.block
      - ORDERER_CHANNELPARTICIPATION_ENABLED=true
    working_dir: /root
    command: orderer
    volumes:
        - ../genesis.block:/var/hyperledger/orderer/orderer.genesis.block
        - ../ordererOrganizations/fish.com/orderers/orderer.fish.com/msp:/var/hyperledger/orderer/msp
        - ../ordererOrganizations/fish.com/orderers/orderer.fish.com/tls:/var/hyperledger/orderer/tls
        - orderer.fish.com:/var/hyperledger/production/orderer
    ports:
      - 7050:7050
      - 8053:8053
      - 9443:9443


{
    "config": {
        "latcId": 1,
        "latcGod": "zltc_RTUbadrrZ9tGSnKtP74JeFHwJ8sWWkBik",
        "latcSaints": [
            "zltc_ndTXvhdEiUfJnBF6mVAu5jMajPNjucTCq"
        ],
        "epoch": 30000,
        "tokenless": false,
        "NoEmptyAnchor": false,
        "EmptyAnchorPeriodMul": 3,
        "period": 10000,
        "GM": false,
        "rootPublicKey": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAn/b6cp+8xJlgl81pNXrb\nTyb/vm5BjxVItDOhQWRZE1bMpWlQY3mmgslyt5+24xnIsKjlvsQzXPEE71aIHQ4G\nJWMz58lX78h/u4TUjFftySWP0zKqxLCq/ftR63eiqz/XuojhOIVfSg+jYWShQYIe\nvuVx/atCghTc03bwnyKN4uK1laSiYA5lwAHtDsSqUO1KA0xMPBlLm1IfmMy6wjKw\n82/CfKCJ2We47q/d5938E1WMOH/nkjwNq/7OY0Oh0xK8Wo5aVlkYbAkaSSnN9zTK\nPCGj9eIWdRMYuonTkMGQsDJanrrYl8JUS0kF7GZCttCi8kNFolZauIhZbbfzsZjx\nRwIDAQAB\n-----END PUBLIC KEY-----\n",
        "isContractVote": true,
        "isDictatorship": true,
        "deployRule": 1
    },
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "number": 0,
    "preacher": "zltc_ndTXvhdEiUfJnBF6mVAu5jMajPNjucTCq",
    "godAmount": 1000000000000000000000000000000,
    "timestamp": "0x5e5f1470"
} 


func (r *RPCServer) handlerHTCL() {
	for {
		select {
		case tx := <-r.rpcHandler.HTCLch:
			if r.chainManger.IsUsableChain(tx.ChainAName) {
				chain := r.chainManger.GetChainByName(tx.ChainAName)
				if chain == nil {
					r.rpcHandler.Errorch <- fmt.Errorf("没有找到名为%s的链", tx.ChainAName)
					continue
				}
				_, err := chain.StartHTCL(tx)
				r.rpcHandler.Errorch <- err
			}
		case unlock := <-r.rpcHandler.Unlockch:
			chain := r.chainManger.GetChainByName(unlock.ChainName)
			if chain == nil {
				r.rpcHandler.Errorch <- fmt.Errorf("没有找到名为%s的链", unlock.ChainName)
				continue
			}
			_, err := chain.UnlockHTCL(unlock.NFTid, unlock.S)
			r.rpcHandler.Errorch <- err
		}
	}
}

type RpcHandler struct {
	HTCLch   chan (dao.HTCLTx)
	Unlockch chan (UnlockHTCL)
	Errorch  chan (error)
}

func (h *RpcHandler) CreateHTCLTxByPreInfo(PreInfo *preHTCL) (int64, error) {
	Tx := dao.HTCLTx{
		TradeNFTID:        PreInfo.TradeNFTID,
		NFTRecipientAddr:  PreInfo.NFTRecipientAddr,
		CoinNUM:           PreInfo.CoinNUM,
		CoinRecipientAddr: PreInfo.CoinRecipientAddr,
		Hs:                PreInfo.Hs,
		TimeStart:         time.Now().Unix(),
		TimeEnd:           time.Now().Unix() + int64(PreInfo.TimeInterval*60*1000),
	}
	return dao.CreateHTCLTx(Tx)
}

func (h *RpcHandler) GetHTCLTxByID(id int64) (dao.HTCLTx, error) {
	return dao.GetHTCLTx(id)
}

func (h *RpcHandler) StartHTCL(id int64) error {
	tx, err := dao.GetHTCLTx(id)
	if err != nil {
		log.Println(err)
		return err
	}
	err = h.handlerHTCL(tx)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (h *RpcHandler) UnlockHTCL(id int64, s string) error {
	tx, err := dao.GetHTCLTx(id)
	if err != nil {
		log.Println(err)
		return err
	}

	unlock := UnlockHTCL{
		ChainName: tx.ChainAName,
		NFTid:     tx.TradeNFTID,
		S:         s,
	}
	h.Unlockch <- unlock
	return nil
}

func (h *RpcHandler) handlerHTCL(tx dao.HTCLTx) error {
	h.HTCLch <- tx
	return <-h.Errorch
}


func (p *P2PServer) RemoteUnlock(txID int64, chainID string, s string) {
	remoteData, err := json.Marshal(&remoteUnlock{
		ChainID: chainID,
		S:       s,
	})
	if err != nil {

	}
	msg := Msg{
		Type: UNLOCK,
		Data: remoteData,
	}
	byteMsg, err := json.Marshal(&msg)
	SendMsg(byteMsg)
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
		ChaincodeID: "testnft",
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
