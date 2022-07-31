package cronjob

import (
	"fmt"
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	service "github.com/copo888/copo_schedule/service/asyncService"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
	"sync"
)

type HandleRepayment struct {
}

/*
   排程每3分钟取出代付提单的还款状态为[3：还款失败][1:待还款][不等于人工处里]的提单，进行还款处理
	3.1.还款前，前往渠道查询提单的目前状态，并依据下面查询到的规则做处理
 (1).成功提单：指交易成功(已完成代付)，将提单转为成功提单，并执行结单。
 (2).失败提单：指无此提单号或交易失败...等相关交易异常，将提单直接还款并结单。
 (3).待处理及处理中提单：将提单转为已上传及处理中，等待回调。
 (4).无此查询通道或其他查询异常：将提单转为人工处里，由后台管理人员处理提单还款或转成功。
*/
func (l *HandleRepayment) Run() {
	var orders []types.OrderX
	// 抓取订单 [1:待还款][3：还款失败] [不等于人工处里]
	//repayment_status 还款状态：(0：不需还款、1:待还款、2：还款成功、3：还款失败)
	//person_process_status: 人工处理状态：(0:待處理1:處理中2:成功3:失敗 10:不需处理)
	if err := helper.COPO_DB.Table("tx_orders").
		Where("`type` = ? AND (repayment_status IN ('1','3')) AND (person_process_status = '10' OR person_process_status is null)", constants.ORDER_TYPE_DF).
		Find(&orders).Error; err != nil {
		logx.Errorf("Err : %s", err.Error())
	}

	logx.Infof("还款失败及待还款提单V2，待处理共 %d 笔, %#v", len(orders), orders)
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
				if err := service.AsyncProxyPayRepayment(url, &order, wg); err != nil {
					logx.Errorf("排程代付人工處理呼叫失敗: %s，Error:%s", order.OrderNo, err.Error())
				}
			}()
		}
		wg.Wait()
		logx.Info("WaitGroup Finished")
	}

}
