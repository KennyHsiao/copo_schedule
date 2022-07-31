package orderService

import (
	"context"
	"fmt"
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/errors"
	"github.com/copo888/copo_schedule/common/model/bo"
	"github.com/copo888/copo_schedule/common/model/vo"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/common/utils"
	"github.com/copo888/transaction_service/common/errorz"

	//"github.com/copo888/transaction_service/common/errorz"
	"github.com/copo888/transaction_service/common/response"
	"github.com/gioco-play/gozzle"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"strconv"
)

/*  代付提單呼叫渠道
	@param respOrder : 代付儲存成功的訂單
    @param rate		 : 商戶配置的費率

	@return errors    : call 渠道返回錯誤
*/
func CallChannel_ProxyOrder(context *context.Context, url string, respOrder *types.OrderX) (*vo.ProxyPayRespVO, error) {

	span := trace.SpanFromContext(*context)

	precise := utils.GetDecimalPlaces(respOrder.OrderAmount)
	valTrans := strconv.FormatFloat(respOrder.OrderAmount, 'f', precise, 64)

	// 新增请求代付请求app 物件 ProxyPayBO
	ProxyPayBO := bo.ProxyPayBO{
		OrderNo:              respOrder.OrderNo,
		TransactionType:      constants.TRANS_TYPE_PROXY_PAY,
		TransactionAmount:    valTrans,
		ReceiptAccountNumber: respOrder.MerchantBankNo,
		ReceiptAccountName:   respOrder.MerchantAccountName,
		ReceiptCardProvince:  respOrder.MerchantBankProvince,
		ReceiptCardCity:      respOrder.MerchantBankCity,
		ReceiptCardArea:      "",
		ReceiptCardBranch:    respOrder.MerchantBankBranch,
		ReceiptCardBankCode:  respOrder.MerchantBankNo,
		ReceiptCardBankName:  respOrder.MerchantBankName,
	}

	// call 渠道app
	ProxyKey, errk := utils.MicroServiceEncrypt(viper.GetString("PROXY_KEY"), viper.GetString("PUBLIC_KEY"))
	if errk != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, errk.Error())
	}
	logx.Infof("EncryptKey: %s，ProxyKey:%s ，PublicKey:%s ", ProxyKey, viper.GetString("PROXY_KEY"), viper.GetString("PUBLIC_KEY"))
	chnResp, chnErr := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationProxykey", ProxyKey).JSON(ProxyPayBO)
	//res, err2 := http.Post(url,"application/json",bytes.NewBuffer(body))
	if chnResp != nil {
		logx.Info("errors Status:", chnResp.Status())
		logx.Info("errors Body:", string(chnResp.Body()))
	}

	proxyPayRespVO := &vo.ProxyPayRespVO{}

	if chnErr != nil {
		logx.Errorf("渠道返回错误: %s， resp: %#v", chnErr.Error(), chnResp)
		return nil, errorz.New(response.CHANNEL_REPLY_ERROR, chnErr.Error())
	} else if chnResp.Status() != 200 {
		logx.Errorf("渠道返回不正确: %d", chnResp.Status())
		return nil, errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("%d", chnResp.Status()))
	} else if decodeErr := chnResp.DecodeJSON(proxyPayRespVO); decodeErr != nil {
		logx.Errorf("渠道返回错误: %s， resp: %#v", decodeErr.Error(), decodeErr)
		return nil, errorz.New(response.CHANNEL_REPLY_ERROR, decodeErr.Error())
	} else if proxyPayRespVO.Code != "0" {
		return proxyPayRespVO, errorz.New(proxyPayRespVO.Code, proxyPayRespVO.Message)
	} else if proxyPayRespVO.Data.ChannelOrderNo == "" {
		logx.Errorf("渠道未回传渠道订单号,%#v", proxyPayRespVO)
		return proxyPayRespVO, errorz.New(errors.INVALID_CHANNEL_ORDER_NO, "渠道未回传渠道订单号")
	}

	logx.Infof("proxyPayRespVO : %#v", proxyPayRespVO)
	return proxyPayRespVO, nil
}

func CallChannel_ProxyQuery(span trace.Span, url string, order *types.OrderX) (*vo.ProxyQueryRespVO, error) {

	proxyQuery := &bo.ProxyQueryBO{
		OrderNo:        order.OrderNo,
		ChannelOrderNo: order.ChannelOrderNo,
	}

	// call 渠道app
	ProxyKey, errk := utils.MicroServiceEncrypt(viper.GetString("PROXY_KEY"), viper.GetString("PUBLIC_KEY"))
	if errk != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, errk.Error())
	}
	logx.Infof("EncryptKey: %s，ProxyKey:%s ，PublicKey:%s ", ProxyKey, viper.GetString("PROXY_KEY"), viper.GetString("PUBLIC_KEY"))
	chnResp, chnErr := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationProxykey", ProxyKey).JSON(proxyQuery)
	//res, err2 := http.Post(url,"application/json",bytes.NewBuffer(body))
	if chnResp != nil {
		logx.Info("errors Status:", chnResp.Status())
		logx.Info("errors Body:", string(chnResp.Body()))
	}

	proxyPayRespVO := &vo.ProxyQueryRespVO{}

	if chnErr != nil {
		logx.Errorf("渠道返回错误: %s， resp: %#v", chnErr.Error(), chnResp)
		return nil, errorz.New(response.CHANNEL_REPLY_ERROR, chnErr.Error())
	} else if chnResp.Status() != 200 {
		logx.Errorf("渠道返回不正确: %d", chnResp.Status())
		return nil, errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("%d", chnResp.Status()))
	} else if decodeErr := chnResp.DecodeJSON(proxyPayRespVO); decodeErr != nil {
		logx.Errorf("渠道返回错误: %s， resp: %#v", decodeErr.Error(), decodeErr)
		return nil, errorz.New(response.CHANNEL_REPLY_ERROR, decodeErr.Error())
	} else if proxyPayRespVO.Code != "0" {
		return proxyPayRespVO, errorz.New(proxyPayRespVO.Code, proxyPayRespVO.Message)
	}
	//渠道為回傳訂單號，之後看狀況加入
	//else if proxyPayRespVO.Data.ChannelOrderNo == "" {
	//	return proxyPayRespVO, errorz.New(errors.INVALID_CHANNEL_ORDER_NO, "渠道未回传渠道订单号")
	//}

	logx.Infof("proxyPayRespVO : %#v", proxyPayRespVO)
	return proxyPayRespVO, nil

	return nil, nil
}

/*
	@param orderNo    : copo訂單號
    @param merOrderNo : 商戶訂單號
*/
func QueryOrderByOrderNo(db *gorm.DB, orderNo string, merOrderNo string) (*types.OrderX, error) {
	txOrder := &types.OrderX{}
	if orderNo != "" || len(orderNo) > 0 {
		if err := db.Table("tx_orders").Where("order_no = ?", orderNo).Find(&txOrder).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	} else if merOrderNo != "" || len(merOrderNo) > 0 {
		if err := db.Table("tx_orders").Where("merchant_order_no = ? OR order_no = ? ", merOrderNo, orderNo).Find(&txOrder).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	} else {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	return txOrder, nil
}
