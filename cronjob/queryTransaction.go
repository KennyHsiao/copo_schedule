package cronjob

import (
	"context"
	"fmt"
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	"github.com/copo888/copo_schedule/service/orderService"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"sync"
	"time"
)

type QueryTransaction struct {
}

func (l *QueryTransaction) Run() {
	var context context.Context
	span := trace.SpanFromContext(context)
	var orders []types.OrderX
	//1.取出代付提单的订单状态[3：交易中]的提单
	if err := helper.COPO_DB.Table("tx_orders").
		Where("`type` = ? AND `status` = ?", constants.ORDER_TYPE_DF, constants.TRANSACTION).
		Find(&orders).Error; err != nil {
		logx.Errorf("Err : %s", err.Error())
	}
	logx.Infof("启动处理交易中，未回调的提单，笔数：%d 笔", len(orders))

	//前往渠道查单(异步处理)
	wg := &sync.WaitGroup{}
	wg.Add(len(orders))
	if len(orders) > 0 {
		for _, order := range orders {
			channel := types.ChannelData{}

			if queryErr := helper.COPO_DB.Table("ch_channels").Where("code = ?", order.ChannelCode).Find(&channel); queryErr != nil {
				logx.Errorf("queryErr: %#v", queryErr.Error)
			}

			url := fmt.Sprintf("%s:%s/api/proxy-pay-query", viper.Get("CHANNEL_HOST"), channel.ChannelPort)
			logx.Infof("發送代付查询請求To渠道: %s。 url: %s", order.OrderNo, url)

			//異步調用-呼叫異步調用服務
			go func() {
				if proxyQueryRespVO, chnErr := orderService.CallChannel_ProxyQuery(span, url, &order); chnErr != nil || proxyQueryRespVO.Code != "0" {
					logx.Errorf("排程代付人工處理呼叫失敗: %s，Error: %s ，進行人工處理，提單狀態改失敗", order.OrderNo, chnErr.Error())
					logx.Errorf("交易中代付提单：%s，状态查询异常 :%s ", order.OrderNo, chnErr.Error())
					order.Status = "30"
					order.PersonProcessStatus = "0" // 人工处理状态：(0:待處理1:處理中2:成功3:失敗 10:不需处理)
					order.ErrorNote = "查询状态回传失败(异常)"

					//TODO 發送人工還款推撥訊息
				} else if proxyQueryRespVO.Code == "0" {
					logx.Infof("查询代付提单URL：%s，提单编号：%s，提单回传结果: checkResult=%#v", url, order.OrderNo, proxyQueryRespVO)
					order.ChannelOrderNo = proxyQueryRespVO.Data.ChannelOrderNo
					order.UpdatedAt = time.Now().UTC()
					t, _ := time.ParseInLocation("2006-01-02 15:04:05", proxyQueryRespVO.Data.ChannelReplyDate, time.Local)
					order.ChannelCallBackAt = t

					if order.Status == "2" { //交易中才执行处理
						if proxyQueryRespVO.Data.OrderStatus == "20" {

						}
					}

				}

				// 更新订单
				if errUpdate := helper.COPO_DB.Table("tx_orders").Updates(order).Error; errUpdate != nil {
					logx.Errorf("代付订单更新状态错误: %s", errUpdate.Error())
				}
			}()
		}
		wg.Wait()
		logx.Info("WaitGroup Finished")
	}
}
