package cronjob

import (
	"context"
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	"github.com/copo888/transaction_service/common/response"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

type CalculateProfit struct {
	logx.Logger
	ctx context.Context
}

func (l *CalculateProfit) Run() {
	var dfOrders []types.OrderX
	var zfOrders []types.OrderX
	var allOrders []types.OrderX
	var profits []*transaction.CalculateProfit

	logx.WithContext(l.ctx).Infof("(補算傭金利潤Schedule)執行開始時間：%s", time.Now().Format("2006-01-02 15:04:05"))
	if err := helper.COPO_DB.Table("tx_orders").
		Where("status != ?", constants.FAIL).
		Where("is_calculate_profit = ?", constants.IS_CALCULATE_PROFIT_NO).
		Where("type = ? ", constants.ORDER_TYPE_DF).
		Find(&dfOrders).Error; err != nil {

		logx.WithContext(l.ctx).Errorf("取得未計算利潤(DF)錯誤:", err.Error())
	}
	if err := helper.COPO_DB.Table("tx_orders").
		Where("status IN (?)", []string{constants.SUCCESS, constants.FROZEN}).
		Where("is_calculate_profit = ?", constants.IS_CALCULATE_PROFIT_NO).
		Where("type = ? ", constants.ORDER_TYPE_ZF).
		Find(&zfOrders).Error; err != nil {

		logx.WithContext(l.ctx).Errorf("取得未計算利潤(ZF)錯誤:", err.Error())
	}
	allOrders = append(allOrders, dfOrders...)
	allOrders = append(allOrders, zfOrders...)

	logx.WithContext(l.ctx).Infof("(補算傭金利潤Schedule)共 %d 筆, 支付 %d 筆, 代付 %d 筆", len(allOrders), len(dfOrders), len(zfOrders))

	if len(allOrders) > 0 {
		for _, txOrder := range allOrders {
			profits = append(profits, &transaction.CalculateProfit{
				MerchantCode:        txOrder.MerchantCode,
				OrderNo:             txOrder.OrderNo,
				Type:                txOrder.Type,
				CurrencyCode:        txOrder.CurrencyCode,
				BalanceType:         txOrder.BalanceType,
				ChannelCode:         txOrder.ChannelCode,
				ChannelPayTypesCode: txOrder.ChannelPayTypesCode,
				OrderAmount:         txOrder.OrderAmount,
			})
		}

		rpc := helper.TransactionRpc
		rpcResp, err := rpc.RecalculateProfitTransaction(context.Background(), &transaction.RecalculateProfitRequest{
			List: profits,
		})

		if err != nil {
			logx.WithContext(l.ctx).Errorf("(補算傭金利潤Schedule)發生錯誤：%s", err.Error())
		} else if rpcResp == nil {
			logx.WithContext(l.ctx).Errorf("(補算傭金利潤Schedule)發生錯誤：rpcResp is nil")
		} else if rpcResp.Code != response.API_SUCCESS {
			logx.WithContext(l.ctx).Errorf("(補算傭金利潤Schedule)發生錯誤：%s", rpcResp.Message)
		} else {
			logx.WithContext(l.ctx).Errorf("(補算傭金利潤Schedule) 完成")
		}
	}
	logx.WithContext(l.ctx).Infof("(補算傭金利潤Schedule)執行結束時間：%s", time.Now().Format("2006-01-02 15:04:05"))
}
