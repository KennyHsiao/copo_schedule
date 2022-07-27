package service

import (
	"context"
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/model/vo"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	service "github.com/copo888/copo_schedule/service/orderService"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"sync"
	"time"
)

func AsyncProxyPayRepayment(url string, order *types.OrderX, context *context.Context, wg *sync.WaitGroup) (*vo.ProxyQueryRespVO, error) {
	defer wg.Done()
	logx.Info("异步人工處理(Restful或Service)====================>开始")
	span := trace.SpanFromContext(*context)

	//call 渠道查詢訂單
	proxyQueryRespVO, chnErr := service.CallChannel_ProxyQuery(span, url, order)
	logx.Infof("提单单号: %s，渠道订单查询结果: %s ", order.OrderNo, proxyQueryRespVO.Data.OrderStatus) //(0:待處理 1:處理中 20:成功 30:失敗 31:凍結)
	//查询回传status=0，成功才执行(失败有可能是网路异常或是网关错误...等)
	if chnErr != nil || proxyQueryRespVO.Code != "0" {
		//查询状态回传失败(异常)，改为人工处里状态、单状态修改为失败
		helper.COPO_DB.Table("tx_orders").
			Where("order_no = ?", order.OrderNo).
			Updates(map[string]interface{}{"status": "30", "person_process_status": "0"})
		//TODO 发送人工还款推播信息

	} else if proxyQueryRespVO.Code == "0" {

		//查询交易状态:成功 (需再多增加判断该笔订单是否已经回调商户，如果未回调是否要在细部区分，让渠道回调来变更)
		if proxyQueryRespVO.Data.OrderStatus == "20" { //成功
			order.Status = "20"
			order.RepaymentStatus = "0" //还款状态：(0：不需还款、1:待还款、2：还款成功、3：还款失败)
			order.ErrorNote = "渠道查询-交易成功"
			order.UpdatedAt = time.Now()

			var callBack bool = false
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
				logx.Error("代付订单更新状态错误: ", errUpdate.Error())
			}

			//回调商户
			if order.Source == constants.API && callBack {
				//回調商戶
				//if errPoseMer := service.PostCallbackToMerchant(helper.COPO_DB, &context, order); errPoseMer != nil {
				//	//不拋錯
				//	logx.Error("回調商戶錯誤:", errPoseMer)
				//}
			}

		}

	}

	//进行还款作业

	return nil, nil
}
