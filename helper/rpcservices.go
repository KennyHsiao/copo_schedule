package helper

import (
	"fmt"
	_ "github.com/neccoys/go-zero-extension/consul"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/zrpc"
	"log"
	"strings"
	"sync"
)

var (
	RpcServices sync.Map
)

func RpcService(channel string) zrpc.Client {

	rpc, ok := RpcServices.Load(channel)
	target := fmt.Sprintf("consul://%s/@?wait=14s", viper.GetString("CONSUL_TARGET"))
	if !ok {
		ch := strings.Replace(target, "@", channel, 1)
		client, err := zrpc.NewClientWithTarget(ch)

		if err != nil {
			log.Panicln("Consul Error:", err)
		}

		return client
	}

	return rpc.(zrpc.Client)

}
