package config

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type peerConfig struct {
	Dest string `toml:"dest"` // 连接其中一个网关
	Port int    `toml:"port"`
}

type Config struct {
	Peer   peerConfig `toml:"peer"`
	Rpc    rpcConfig  `toml:"rpc"`
	Chain  Chain      `toml:"chain"`
	Fabric Fabric     `toml:"fabric"`
}

type rpcConfig struct {
	Port int `toml:"port"`
}

type Fabric struct {
	Name       string `toml:"name"`
	ConfigPath string `toml:"configPath"`
	OrgName    string `toml:"orgname"`
	OrgAdmin   string `toml:"orgadmin"`
	OrgUser    string `toml:"orguser"`

	// Same for each peer
	ChannelID string   `toml:"channelid"`
	Peers     []string `toml:"peers"`
}

type Chain struct {
	Name       string `toml:"name"`
	ConfigPath string `toml:"configPath"`
}

func GetConfig() *Config {
	Conf := &Config{}
	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("filed to open config file:%v ", err)
		return nil
	}
	_, err = toml.DecodeFile(pwd+"/config.toml", Conf)

	if err != nil {
		log.Printf("filed to decode config file:%v ", err)
		return nil
	}
	return Conf
}
