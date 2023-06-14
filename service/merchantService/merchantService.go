package service

import (
	"context"
	"fmt"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/common/utils"
	"github.com/gioco-play/gozzle"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"net/url"
)

/*
	回調-商戶代付結果(須注意MerchantService Project 也有一組再用，修改時要注意)
	@return
*/
func PostCallbackToMerchant(db *gorm.DB, context *context.Context, orderX *types.OrderX) (err error) {
	span := trace.SpanFromContext(*context)
	merchant := &types.Merchant{}
	if err = db.Table("mc_merchants").Where("code = ?", orderX.MerchantCode).Find(merchant).Error; err != nil {
		return
	}

	//precise := utils2.GetDecimalPlaces(orderX.OrderAmount)
	//valTrans := strconv.FormatFloat(orderX.OrderAmount, 'f', precise, 64)
	//
	//precise2 := utils2.GetDecimalPlaces(orderX.Fee)
	//valTrans2 := strconv.FormatFloat(orderX.Fee, 'f', precise2, 64)

	status := changeOrderStatusToMerchant(orderX.Status)

	ProxyPayCallBackMerRespVO := url.Values{}
	ProxyPayCallBackMerRespVO.Set("merchantId", orderX.MerchantCode)
	ProxyPayCallBackMerRespVO.Set("orderNo", orderX.MerchantOrderNo)
	ProxyPayCallBackMerRespVO.Set("payOrderNo", orderX.OrderNo)
	ProxyPayCallBackMerRespVO.Set("orderStatus", status)
	ProxyPayCallBackMerRespVO.Set("orderAmount", fmt.Sprintf("%.2f", orderX.OrderAmount))
	ProxyPayCallBackMerRespVO.Set("fee", fmt.Sprintf("%.2f", orderX.Fee))
	ProxyPayCallBackMerRespVO.Set("payOrderTime", orderX.TransAt.Time().Format("200601021504"))
	if orderX.CurrencyCode == "PHP" {
		ProxyPayCallBackMerRespVO.Set("errorNote", orderX.ErrorNote)
	}

	if err != nil {
		logx.WithContext(*context).Error(err.Error())
	}
	sign := utils.SortAndSignFromUrlValues(ProxyPayCallBackMerRespVO, merchant.ScrectKey)
	ProxyPayCallBackMerRespVO.Set("Sign", sign)
	logx.WithContext(*context).Infof("代付提单 %s ，回调商户URL= %s，回调资讯= %#v", orderX.OrderNo, orderX.NotifyUrl, ProxyPayCallBackMerRespVO)

	//TODO retry post for 10 times and 2s between each reqeuest
	//merResp, merCallBackErr := gozzle.Post("http://172.16.204.115:8083/dior/merchant-api/merchant-call-back").Timeout(10).Trace(span).Form(ProxyPayCallBackMerRespVO)
	merResp, merCallBackErr := gozzle.Post(orderX.NotifyUrl).Timeout(10).Trace(span).Form(ProxyPayCallBackMerRespVO)
	if merCallBackErr != nil || merResp.Status() != 200 {
		if merCallBackErr != nil {
			logx.WithContext(*context).Errorf("代付提单%s 回调商户异常，錯誤: %#v", ProxyPayCallBackMerRespVO.Get("OrderNo"), merCallBackErr)
		} else if merResp.Status() != 200 {
			logx.WithContext(*context).Errorf("响应状态 %d 错误", merResp.Status())
		}
	}
	logx.WithContext(*context).Infof("代付提单 %s ，回调商户請求參數 %#v，商戶返回: %#v", ProxyPayCallBackMerRespVO.Get("OrderNo"), ProxyPayCallBackMerRespVO, string(merResp.Body()))
	return
}

func changeOrderStatusToMerchant(status string) string {
	var changeStatus string

	if status == "0" {
		changeStatus = "0"
	} else if status == "1" { //(0:待處理 1:處理中 2:交易中 20:成功 30:失敗 31:凍結)
		changeStatus = "4"
	} else if status == "2" {
		changeStatus = "3"
	} else if status == "20" {
		changeStatus = "1"
	} else if status == "30" || status == "31" {
		changeStatus = "2"
	}

	return changeStatus
}
