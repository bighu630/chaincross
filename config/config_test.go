package config

import (
	"fmt"
	"testing"
)

func TestGetConfig(t *testing.T) {
	conf := GetConfig()
	fmt.Println(conf)
}
