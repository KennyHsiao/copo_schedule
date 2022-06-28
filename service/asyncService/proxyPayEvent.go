package service

import (
	"fmt"
	_ "fmt"
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/errors"
	"github.com/copo888/copo_schedule/common/model/vo"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	"github.com/copo888/copo_schedule/service/orderService"
	"github.com/copo888/transaction_service/common/errorz"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/neccoys/go-zero-extension/redislock"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/net/context"
	"sync"
)

type ProxyPayEvent struct {
	logx.Logger
	ctx context.Context
}

func AsyncProxyPayEvent(url string, order *types.OrderX, wg *sync.WaitGroup) (respVO *vo.ProxyPayRespVO, err error) {
	redisKey := fmt.Sprintf("%s-%s", order.MerchantCode, order.OrderNo)
	redisLock := redislock.New(helper.REDIS, redisKey, "proxy-call-back:")
	redisLock.SetExpire(5)
	//为避免代付提单在发送过程中，三方渠道突然callback回调，导致余状态异常，故增加一把Redis Lock 原则上没送单之前，应该不会有任何动作产生
	if isOK, _ := redisLock.Acquire(); isOK {
		if respVO, err = internal_AsyncProxyPayEvent(url, order, wg); err != nil {
			return nil, err
		}
		defer redisLock.Release()
	} else {
		//为避免已经有其他逻辑正在处里，故这边不对Redis Lock抛出的Exception做任何处里
		logx.Infof("提单 %s 目前正在处理中(Redis Lock)，无法发送", order.OrderNo)
		return nil, errorz.New(errors.INSERT_REDIS_FAILURE)
	}
	return respVO, nil
}

func internal_AsyncProxyPayEvent(url string, order *types.OrderX, wg *sync.WaitGroup) (*vo.ProxyPayRespVO, error) {
	defer wg.Done()
	logx.Info("异步调代付渠道服务(Restful或Service)====================>开始")
	logx.Infof("发送代付提单 %s 处理请求 To 渠道：%s 网关地址:%s", order.OrderNo, order.ChannelCode, url)
	var context context.Context
	// 1. call 渠道app
	//var chnErr error
	//proxyPayRespVO := &vo.ProxyPayRespVO{} //接渠道返回的物件
	var respOrder = &types.OrderX{} // 返回上層的TxOrder物件
	var queryErr error
	if respOrder, queryErr = orderService.QueryOrderByOrderNo(helper.COPO_DB, order.OrderNo, ""); queryErr != nil {
		logx.Errorf("撈取代付訂單錯誤: %s, respOrder:%#v", queryErr, respOrder)
		return nil, errorz.New(errors.FAIL, queryErr.Error())
	}

	proxyPayRespVO, chnErr := orderService.CallChannel_ProxyOrder(&context, url, order)
	// 2. 渠道返回處理 錯誤:商戶錢包加回
	if chnErr != nil || proxyPayRespVO.Code != "0" { //將渠道回傳的錯誤訊息用proxyPayRespVO回傳
		logx.Errorf("代付提單: %s ，渠道返回錯誤: %s, %#v", order.OrderNo, chnErr, proxyPayRespVO)

		rpc := transactionclient.NewTransaction(helper.RpcService("transaction.rpc"))
		var resRpc *transaction.ProxyPayFailResponse
		var errRpc error
		//transaction: 1 .将商户钱包加回 (merchantCode, orderNO) 2. 更新狀態為失敗單
		if order.BalanceType == "DFB" {
			resRpc, errRpc = rpc.ProxyOrderTransactionFail_DFB(context, &transaction.ProxyPayFailRequest{
				MerchantCode: order.MerchantCode,
				OrderNo:      order.OrderNo,
			})
		} else if order.BalanceType == "XFB" {
			resRpc, errRpc = rpc.ProxyOrderTransactionFail_XFB(context, &transaction.ProxyPayFailRequest{
				MerchantCode: order.MerchantCode,
				OrderNo:      order.OrderNo,
			})
		}

		if errRpc != nil {
			return nil, errorz.New(errors.TRANSACTION_FAILURE)
		}

		//因在transaction_service 已更新過訂單，重新抓取訂單
		if respOrder, queryErr := orderService.QueryOrderByOrderNo(helper.COPO_DB, order.OrderNo, ""); queryErr != nil {
			logx.Errorf("撈取代付訂單錯誤: %s, respOrder:%#v", queryErr, respOrder)
			return nil, errorz.New(errors.DATABASE_FAILURE, queryErr.Error())
		}

		respOrder.ErrorType = "1" //   1.渠道返回错误	2.渠道异常	3.商户参数错误	4.账户为黑名单	5.其他
		respOrder.ErrorNote = "渠道返回错误: " + proxyPayRespVO.Message

		if errRpc != nil {
			logx.Errorf("代付提单 %s 还款失败。 Err: %s", respOrder.OrderNo, errRpc.Error())
			respOrder.RepaymentStatus = constants.REPAYMENT_FAIL
			return nil, errorz.New(errors.FAIL, errRpc.Error())
		} else {
			logx.Infof("代付還款rpc完成，%s 錢包還款完成: %#v", order.BalanceType, resRpc)
			respOrder.RepaymentStatus = constants.REPAYMENT_SUCCESS
		}
	} else {
		respOrder.ChannelOrderNo = proxyPayRespVO.Data.ChannelOrderNo
		//条整订单状态从"待处理" 到 "交易中"
		respOrder.Status = constants.TRANSACTION
	}

	// 更新订单
	if errUpdate := helper.COPO_DB.Table("tx_orders").Updates(&respOrder).Error; errUpdate != nil {
		logx.Error("代付订单更新状态错误: ", errUpdate.Error())
	}

	return proxyPayRespVO, nil

}
