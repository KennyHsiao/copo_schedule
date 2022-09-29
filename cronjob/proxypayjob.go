package cronjob

import (
	"context"
	"fmt"
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/model/vo"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	service "github.com/copo888/copo_schedule/service/asyncService"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
	"sync"
	"time"
)

type ProxyToChannel struct {
	logx.Logger
	ctx context.Context
}

func NewProxyToChannel(ctx context.Context) ProxyToChannel {
	return ProxyToChannel{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
	}
}

func (l *ProxyToChannel) Run() {
	var orders []types.OrderX
	var updateOrders []types.OrderX

	p := service.NewProxyPayEvent(l.ctx)

	if err := helper.COPO_DB.Table("tx_orders").Where("`type` = ? AND `status` = ? ", constants.ORDER_TYPE_DF, constants.WAIT_PROCESS).Find(&orders).Error; err != nil {
		logx.WithContext(l.ctx).Info("Err", err.Error())
	}
	copier.Copy(&updateOrders, &orders)

	logx.WithContext(l.ctx).Infof("执行时间：%s，待处理-[代付提单]，共 %d 笔", time.Now().Format("2006-01-02 15:04:05"), len(orders))
	if len(updateOrders) > 0 {
		logx.WithContext(l.ctx).Infof("已处理-[代付提单updateStatusByScheduleBOs]，共 %d 笔", len(updateOrders))
		var IDs []int64
		helper.COPO_DB.Table("tx_orders").Select("id").Where("`type` = ? AND `status` = ? ", constants.ORDER_TYPE_DF, constants.WAIT_PROCESS).Find(&IDs)

		if errUpdate := helper.COPO_DB.Table("tx_orders").Where("id IN (?)", IDs).Updates(map[string]interface{}{"status": "1"}).Error; errUpdate != nil {
			logx.WithContext(l.ctx).Info(errUpdate.Error)
			logx.WithContext(l.ctx).Infof("排程发送前先更新订单状态用，待处理 => 处理中 >> 更新异常，ERR: %v ", errUpdate)
			return
		}

		//呼叫渠道送出(異部處理)
		wg := &sync.WaitGroup{}
		wg.Add(len(updateOrders))
		for _, order := range updateOrders {
			//没有回调渠道订单号，以及渠道回调时间/交易时间，才会发送到渠道
			if order.ChannelOrderNo == "" && order.ChannelCallBackAt.IsZero() && order.TransAt.Time().IsZero() {
				channel := types.ChannelData{}
				if queryErr := helper.COPO_DB.Table("ch_channels").Where("code = ?", order.ChannelCode).Find(&channel); queryErr != nil {
					logx.WithContext(l.ctx).Error("queryErr: ", queryErr)
				}

				url := fmt.Sprintf("%s:%s/api/proxy-pay", viper.Get("CHANNEL_HOST"), channel.ChannelPort)
				logx.WithContext(l.ctx).Infof("發送代付處理請求To渠道: %v。 url: %s", order, url)

				//異步調用-呼叫異步調用服務
				resp := &vo.ProxyPayRespVO{}
				updateOrder := &types.OrderX{}
				var err error
				go func() {
					resp, err = p.AsyncProxyPayEvent(url, &order, wg)
					if err != nil{
						logx.WithContext(l.ctx).Infof("resp:%s", err)
					}
					logx.WithContext(l.ctx).Infof("resp:%#v", resp)

					if resp != nil {
						proxyPayErrorNote := "渠道返还-" + resp.Message
						if resp.Code == "1" { ////回复失败，都列为失败单，不管是否为网路异常....等
							logx.WithContext(l.ctx).Errorf("代付提单: %s ，渠道交易失败讯息: %s", order.OrderNo, proxyPayErrorNote)
							updateOrder.Status = constants.FAIL
							updateOrder.RepaymentStatus = constants.REPAYMENT_WAIT

							//TODO 通知訊息-异步调用-呼叫代付渠道发送代付提单服务异常
							//restService.sendLineNotifyMessage("异步调用-呼叫代付渠道发送代付提单服务异常，channelCoding: "+orderVO.getChannelCoding()+"，网关网址:" +channelGateway +"，单号:"+orderVO.getProxyPayOrderNo()+ "，渠道返回:" +proxyPayErrorNote);
							logx.WithContext(l.ctx).Errorf("异步调用-呼叫代付渠道发送代付提单服务异常，channelCoding: %s，网关网址:%s，单号:%s，渠道返回:%s", updateOrder.ChannelCode, url, updateOrder.OrderNo, proxyPayErrorNote)
						}

						//若成功则更新交易中、無須還款
						copier.Copy(updateOrder, order)
						updateOrder.Status = constants.TRANSACTION
						updateOrder.RepaymentStatus = constants.REPAYMENT_NOT
					}
				}()
			}
		}
		wg.Wait()
		logx.WithContext(l.ctx).Info("WaitGroup Finished")
	}
}
