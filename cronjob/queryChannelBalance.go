package cronjob

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryChannelBalance struct {
	logx.Logger
	ctx context.Context
}

func NewQueryChannelBalance(ctx context.Context) QueryChannelBalance {
	return QueryChannelBalance{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
	}
}

func (l *QueryChannelBalance) Run() {

	//var channels []types.ChannelData
	//span := trace.SpanFromContext(l.ctx)
	//if err := helper.COPO_DB.Table("ch_channels").Where("`status` = ? ","1").Find(&channels).Error; err != nil {
	//	logx.WithContext(l.ctx).Info("Err", err.Error())
	//}
	//
	//if len(channels) > 0 {
	//	//有余额查询URL才做查询
	//	for _,channel := range channels {
	//
	//		if !strings.EqualFold(channel.ProxyPayQueryUrl, "") {
	//			ProxyKey, errk := utils.MicroServiceEncrypt(viper.GetString("PROXY_KEY"), viper.GetString("PUBLIC_KEY"))
	//			if errk != nil {
	//				logx.Errorf("MicroServiceEncrypt Error: %s",errk.Error())
	//			}
	//			proxyQueryBalanceRespVO := &vo.ProxyQueryBalanceRespVO{}
	//			url := fmt.Sprintf("%s:%s/api/proxy-pay-query-balance-internal", svcCtx.Config.Server, channel.ChannelPort)
	//			if chnResp, chnErr := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationProxyPaykey", ProxyKey).JSON(nil); chnErr != nil {
	//				return chnErr
	//			} else if decErr := chnResp.DecodeJSON(proxyQueryBalanceRespVO); decErr != nil {
	//				return decErr
	//			} else if proxyQueryBalanceRespVO.Code != "0" {
	//				return errorz.New(response.UPDATE_CHANNEL_BALANCE_ERROR, proxyQueryBalanceRespVO.Data.ChannelCodingtring)
	//			}
	//
	//			var proxypayBalance float64 = 0
	//			var errBalance error
	//			if proxypayBalance, errBalance = strconv.ParseFloat(proxyQueryBalanceRespVO.Data.ProxyPayBalance, 64); errBalance != nil {
	//				return errBalance
	//			}
	//
	//			channel.ProxypayBalance = proxypayBalance
	//		}
	//
	//
	//
	//	}
	//
	//
	//}

}
