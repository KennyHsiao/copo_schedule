package cronjob

import (
	"context"
	"github.com/copo888/copo_schedule/helper"
	"github.com/copo888/transaction_service/common/response"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type CalculateProfit struct {

}

func (l *CalculateProfit) Run() {
	rpcRequest := transaction.ConfirmPayOrderRequest{

	}
	// CALL transactionc
	ctx := context.Background()
	rpc := transactionclient.NewTransaction(helper.RpcService("transaction.rpc"))
	rpcResp, err2 := rpc.ConfirmPayOrderTransaction(ctx, &rpcRequest)
	if err2 != nil {
		logx.Error("err")
	} else if rpcResp == nil {
		logx.Error("PayOrderTranaction rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		logx.Error("rpcResp.Code err")
	} else {
		logx.Error("ok")
	}
}
