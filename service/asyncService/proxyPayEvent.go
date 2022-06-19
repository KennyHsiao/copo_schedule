package service

import (
	"fmt"
	_ "fmt"
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/model/bo"
	"github.com/copo888/copo_schedule/common/model/vo"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/common/utils"
	"github.com/copo888/copo_schedule/helper"
	"github.com/gioco-play/gozzle"
	"github.com/neccoys/go-zero-extension/redislock"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
	"sync"
)

func AsyncProxyPayEvent(url string, order *types.OrderX, wg *sync.WaitGroup) *vo.ProxyPayRespVO {
	var respVO *vo.ProxyPayRespVO
	redisKey := fmt.Sprintf("%s-%s", order.MerchantCode, order.OrderNo)
	redisLock := redislock.New(helper.REDIS, redisKey, "proxy-call-back:")
	redisLock.SetExpire(5)
	//为避免代付提单在发送过程中，三方渠道突然callback回调，导致余状态异常，故增加一把Redis Lock 原则上没送单之前，应该不会有任何动作产生
	if isOK, _ := redisLock.Acquire(); isOK {
		if err := internal_AsyncProxyPayEvent(url, order, wg); err != nil {
			return err
		}
		defer redisLock.Release()
	} else {
		//为避免已经有其他逻辑正在处里，故这边不对Redis Lock抛出的Exception做任何处里
		logx.Infof("提单 %s 目前正在处理中(Redis Lock)，无法发送", order.OrderNo)
		return nil
	}
	return respVO
}

func internal_AsyncProxyPayEvent(url string, order *types.OrderX, wg *sync.WaitGroup) *vo.ProxyPayRespVO {
	defer wg.Done()
	logx.Info("异步调代付渠道服务(Restful或Service)====================>开始")
	logx.Infof("发送代付提单 %s 处理请求 To 渠道：%s 网关地址:%s", order.OrderNo, order.ChannelCode, url)
	var context context.Context
	span := trace.SpanFromContext(context)
	// call 渠道app
	ProxyKey, errk := utils.MicroServiceEncrypt(viper.GetString("PROXY_KEY"), viper.GetString("PUBLIC_KEY"))
	if errk != nil {
		logx.Errorf("微服务加密错误: %s ", errk.Error())
	}

	// 新增请求代付请求app 物件 ProxyPayBO
	ProxyPayBO := bo.ProxyPayBO{
		OrderNo:              order.OrderNo,
		TransactionType:      constants.TRANS_TYPE_PROXY_PAY,
		TransactionAmount:    fmt.Sprintf("%f", order.OrderAmount),
		ReceiptAccountNumber: order.MerchantBankNo,
		ReceiptAccountName:   order.MerchantAccountName,
		ReceiptCardProvince:  order.MerchantBankProvince,
		ReceiptCardCity:      order.MerchantBankCity,
		ReceiptCardArea:      "",
		ReceiptCardBranch:    order.MerchantBankBranch,
		ReceiptCardBankCode:  order.MerchantBankNo,
		ReceiptCardBankName:  order.MerchantBankName,
	}

	chnResp, chnErr := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationProxyPaykey", ProxyKey).JSON(ProxyPayBO)
	//res, err2 := http.Post(url,"application/json",bytes.NewBuffer(body))
	if chnResp != nil {
		logx.Info("response Status:", chnResp.Status())
		logx.Info("response Body:", string(chnResp.Body()))
	}
	proxyPayRespVO := &vo.ProxyPayRespVO{}
	if chnErr != nil {
		proxyPayRespVO.Code = "1"
		proxyPayRespVO.Message = chnErr.Error()
	} else if decodeErr := chnResp.DecodeJSON(proxyPayRespVO); decodeErr != nil {
		proxyPayRespVO.Code = "1"
		proxyPayRespVO.Message = decodeErr.Error()
		logx.Errorf("渠道返回错误: %s， resp: %#v", decodeErr.Error(), decodeErr)
	}

	return proxyPayRespVO

}
