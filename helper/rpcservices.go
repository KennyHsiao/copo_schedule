package helper

import (
	"fmt"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	_ "github.com/neccoys/go-zero-extension/consul"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/zrpc"
	"sync"
)

var (
	RpcServices sync.Map
	TransactionRpc transactionclient.Transaction
)

func init() {
	transactionClient := zrpc.MustNewClient(zrpc.RpcClientConf{
		Target: fmt.Sprintf("consul://%s/transaction.rpc?wait=14s", viper.GetString("CONSUL_TARGET")),
	})
	TransactionRpc = transactionclient.NewTransaction(transactionClient)
}

//func RpcService(channel string) zrpc.Client {
//
//	rpc, ok := RpcServices.Load(channel)
//	target := fmt.Sprintf("consul://%s/@?wait=14s", viper.GetString("CONSUL_TARGET"))
//	if !ok {
//		ch := strings.Replace(target, "@", channel, 1)
//		client, err := zrpc.NewClientWithTarget(ch)
//
//		if err != nil {
//			log.Panicln("Consul Error:", err)
//		}
//
//		return client
//	}
//
//	return rpc.(zrpc.Client)
//
//}
