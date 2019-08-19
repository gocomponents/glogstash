package config

import (
	"fmt"
	"testing"
)

func TestGetElasticConfig(t *testing.T) {
	config:=GetElasticConfig()
	fmt.Println(config)
}
