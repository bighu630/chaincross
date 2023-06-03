package chainmanger

import (
	conf "chainCross/config"
	"chainCross/dao"
)

type ChainManager struct {
	ChainList map[string]*Chain // chainname - chain
}

func NewChainManger() *ChainManager {
	return &ChainManager{
		ChainList: make(map[string]*Chain),
	}
}

// 判断chain是否在chainList中
func (c ChainManager) IsUsableChain(chainID string) bool {
	for _, it := range c.ChainList {
		if chainID == it.name {
			return true
		}
	}
	return false
}

func (c *ChainManager) GetChainByName(chainID string) *Chain {
	if c.IsUsableChain(chainID) {
		return c.ChainList[chainID]
	}
	return nil
}

// 添加一条链
func (c *ChainManager) AddChain(chain *Chain) {
	if _, ok := c.ChainList[chain.name]; !ok {
		c.ChainList[chain.name] = chain
	}
}

func (c *ChainManager) AddChainByConfig(conf conf.Fabric) {
	if conf.Name == "" {
		return
	}
	c.AddChain(NewChainClient(conf))
	dao.AddChainsName([]string{conf.Name})
}
