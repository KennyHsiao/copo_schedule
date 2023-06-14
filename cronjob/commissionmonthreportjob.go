package cronjob

import (
	"context"
	"fmt"
	"github.com/copo888/copo_schedule/helper"
	"github.com/copo888/copo_schedule/service/linenotifyService"
	"github.com/copo888/transaction_service/common/response"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

type CommissionMonthReport struct {
	logx.Logger
	ctx context.Context
}

func (l *CommissionMonthReport) Run() {

	location, _ := time.LoadLocation("Asia/Taipei")
	month := time.Now().In(location).AddDate(0, -1, 0).Format("2006-01")

	logx.WithContext(l.ctx).Infof("(計算月傭金報表 Schedule) %s 執行開始時間：%s", month, time.Now().Format("2006-01-02 15:04:05"))

	rpcRequest := transaction.CalculateCommissionMonthAllRequest{
		Month: month,
	}
	// CALL transactionc
	rpc := helper.TransactionRpc
	rpcResp, err := rpc.CalculateCommissionMonthAllReport(context.Background(), &rpcRequest)

	if err != nil {
		logx.WithContext(l.ctx).Errorf("(計算月傭金報表 Schedule)發生錯誤：%s", err.Error())
		msg := fmt.Sprintf("(計算月傭金報表 Schedule)發生錯誤：'%s'", err.Error())
		linenotifyService.DoCallLineSendURL(l.ctx, msg)
	} else if rpcResp == nil {
		logx.WithContext(l.ctx).Errorf("(計算月傭金報表 Schedule)發生錯誤：rpcResp is nil")
		msg := fmt.Sprintf("(計算月傭金報表 Schedule)發生錯誤：rpcResp is nil")
		linenotifyService.DoCallLineSendURL(l.ctx, msg)
	} else if rpcResp.Code != response.API_SUCCESS {
		logx.WithContext(l.ctx).Errorf("(計算月傭金報表 Schedule)發生錯誤：%s", rpcResp.Message)
		msg := fmt.Sprintf("(計算月傭金報表 Schedule)發生錯誤：'%s'", rpcResp.Message)
		linenotifyService.DoCallLineSendURL(l.ctx, msg)
	} else {
		logx.WithContext(l.ctx).Errorf("(計算月傭金報表 Schedule) 完成")
	}
	logx.WithContext(l.ctx).Infof("(計算月傭金報表 Schedule) %s 執行結束時間：%s", month, time.Now().Format("2006-01-02 15:04:05"))
}
