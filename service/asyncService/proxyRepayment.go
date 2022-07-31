package service

import (
	"context"
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	service "github.com/copo888/copo_schedule/service/merchantService"
	"github.com/copo888/copo_schedule/service/orderService"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"sync"
	"time"
)

func AsyncProxyPayRepayment(url string, order *types.OrderX, wg *sync.WaitGroup) error {
	defer wg.Done()
	logx.Info("异步人工處理(Restful或Service)====================>开始")
	context := context.Background()
	span := trace.SpanFromContext(context)

	var repayment_flag bool = false //使否执行还款
	var callBack bool = false       //是否回调商户
	//call 渠道查詢訂單
	proxyQueryRespVO, chnErr := orderService.CallChannel_ProxyQuery(span, url, order)
	logx.Infof("提单单号: %s，渠道订单查询结果: %s ", order.OrderNo, proxyQueryRespVO.Data.OrderStatus) //(0:待處理 1:處理中 20:成功 30:失敗 31:凍結)
	//查询回传status=0，成功才执行(失败有可能是网路异常或是网关错误...等)
	if chnErr != nil || proxyQueryRespVO.Code != "0" {
		//查询状态回传失败(异常)，改为人工处里状态、单状态修改为失败
		logx.Errorf("查询状态回传失败(异常): %s", chnErr.Error())
		helper.COPO_DB.Table("tx_orders").
			Where("order_no = ?", order.OrderNo).
			Updates(map[string]interface{}{"status": "30", "person_process_status": "0", "repayment_status": "1", "error_note": "查询状态回传失败(异常)"})

		// 新單新增訂單歷程 (不抱錯) TODO: 異步??
		if err4 := helper.COPO_DB.Table("tx_order_actions").Create(&types.OrderActionX{
			OrderAction: types.OrderAction{
				OrderNo:     order.OrderNo,
				Action:      "PERSON_PROCESSING",
				UserAccount: order.MerchantCode,
				Comment:     "",
			},
		}).Error; err4 != nil {
			logx.Error("紀錄訂單歷程出錯:%s", err4.Error())
		}

		//TODO 发送人工还款推播信息

	} else if proxyQueryRespVO.Code == "0" {
		//查询交易状态:成功 (需再多增加判断该笔订单是否已经回调商户，如果未回调是否要在细部区分，让渠道回调来变更)
		if proxyQueryRespVO.Data.OrderStatus == "20" { //成功
			order.Status = "20"
			order.RepaymentStatus = "0" //还款状态：(0：不需还款、1:待还款、2：还款成功、3：还款失败)
			order.ErrorNote = "渠道查询-交易成功"
			order.UpdatedAt = time.Now()

			if order.IsMerchantCallback == constants.MERCHANT_CALL_BACK_NO {
				callBack = true
			}

			if callBack { //是否已经回调商户(0：否、1:是、2:不需回调)(透过API需提供的资讯)
				order.IsMerchantCallback = constants.MERCHANT_CALL_BACK_YES
				order.MerchantCallBackAt = time.Now().UTC()
			}
			//更新提单information状态
			logx.Infof("排程查询提单状态成功单，更新代付主表资讯：%#v", order)
			// 更新订单
			if errUpdate := helper.COPO_DB.Table("tx_orders").Updates(order).Error; errUpdate != nil {
				logx.Errorf("代付订单更新状态错误: %s", errUpdate.Error())
			}

			//回调商户
			if order.Source == constants.API && callBack {
				if errPoseMer := service.PostCallbackToMerchant(helper.COPO_DB, &context, order); errPoseMer != nil {
					//不拋錯
					logx.Error("回調商戶錯誤:", errPoseMer)
				}
			}

		} else if proxyQueryRespVO.Data.OrderStatus == "30" { //失敗
			//查寻结果：交易失败，执行还款作业
			logx.Infof("提单 %s 查寻结果：交易失败，执行还款作业", order.OrderNo)
			//修正重新查询回来后，确认为失败时，复写错误原因
			repayment_flag = true
			order.ErrorNote = "渠道查询-交易失敗"      //复写错误原因
			order.UpdatedAt = time.Now().UTC() //更新日期
			// 交易日期(渠道回复的日期)为空值，补写入日期

			if order.IsMerchantCallback == constants.MERCHANT_CALL_BACK_NO {
				callBack = true
			}

			if callBack { //是否已经回调商户(0：否、1:是、2:不需回调)(透过API需提供的资讯)
				order.IsMerchantCallback = constants.MERCHANT_CALL_BACK_YES
				order.MerchantCallBackAt = time.Now().UTC()
			}

		} else if proxyQueryRespVO.Data.OrderStatus == "0" || proxyQueryRespVO.Data.OrderStatus == "1" { //0:待處理 1:處理中
			order.Status = constants.TRANSACTION
			order.ErrorNote = "渠道查询-交易中"
			order.RepaymentStatus = constants.REPAYMENT_NOT
			order.UpdatedAt = time.Now().UTC()
			order.ChannelOrderNo = proxyQueryRespVO.Data.ChannelOrderNo
		}

		//更新提单information状态
		logx.Infof("排程查询提单状态失败单，更新代付资讯：%#v", order)
		// 更新订单
		if errUpdate := helper.COPO_DB.Table("tx_orders").Updates(order).Error; errUpdate != nil {
			logx.Errorf("代付订单更新状态错误: %s", errUpdate.Error())
			return errUpdate
		}

		var action string
		if order.Status == constants.SUCCESS {
			action = "SUCCESS"
		} else if order.Status == constants.FAIL {
			action = "FAILURE"
		} else if order.Status == constants.TRANSACTION {
			action = "TRANSACTION"
		}

		// 新單新增訂單歷程 (不抱錯) TODO: 異步??
		if err4 := helper.COPO_DB.Table("tx_order_actions").Create(&types.OrderActionX{
			OrderAction: types.OrderAction{
				OrderNo:     order.OrderNo,
				Action:      action,
				UserAccount: order.MerchantCode,
				Comment:     "",
			},
		}).Error; err4 != nil {
			logx.Error("紀錄訂單歷程出錯:%s", err4.Error())
		}

	}

	//進行人工還款
	if repayment_flag {
		logx.Infof("执行还款，還款單號:%s", order.OrderNo)
		var errRpc error
		balanceType, errBalance := service.GetBalanceType(helper.COPO_DB, order.ChannelCode, constants.ORDER_TYPE_DF)
		if errBalance != nil {
			return errBalance
		}

		//呼叫RPC
		rpc := transactionclient.NewTransaction(helper.RpcService("transaction.rpc"))
		rpcRequest := transaction.ProxyPayFailRequest{
			MerchantCode: order.MerchantCode,
			OrderNo:      order.OrderNo,
		}
		//當訂單還款狀態為待还款
		if order.RepaymentStatus == constants.REPAYMENT_WAIT {
			//将商户钱包加回 (merchantCode, orderNO)，更新狀態為失敗單
			var resRpc *transaction.ProxyPayFailResponse
			if balanceType == "DFB" {
				resRpc, errRpc = rpc.ProxyOrderTransactionFail_DFB(context, &rpcRequest)
			} else if balanceType == "XFB" {
				resRpc, errRpc = rpc.ProxyOrderTransactionFail_XFB(context, &rpcRequest)
			}

			if errRpc != nil {
				logx.Errorf("代付提单回调 %s 还款失败。 Err: %s", order.OrderNo, errRpc.Error())
				order.RepaymentStatus = constants.REPAYMENT_FAIL
				return errRpc
			} else {
				logx.Infof("代付還款rpc完成，%s 錢包還款完成: %#v", balanceType, resRpc)
				order.RepaymentStatus = constants.REPAYMENT_SUCCESS
				//TODO 收支紀錄
			}

			// 更新订单
			if errUpdate := helper.COPO_DB.Table("tx_orders").Updates(order).Error; errUpdate != nil {
				logx.Error("代付订单更新状态错误: ", errUpdate.Error())
			}
		}

	}

	//回调商户
	if order.Source == constants.API && callBack {
		logx.Infof("代付订单回调状态码: %s，增加主动回调API订单：%s=======================================>", order.Status, order.OrderNo)
		if errPoseMer := service.PostCallbackToMerchant(helper.COPO_DB, &context, order); errPoseMer != nil {
			//不拋錯
			logx.Error("回調商戶錯誤:", errPoseMer)
		}
	}
	return nil
}
