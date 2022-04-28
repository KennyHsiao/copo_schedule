package cronjob

import (
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

type ProxyToChannel struct {
}

func (l *ProxyToChannel) Run() {
	var orders []types.OrderX
	if err := helper.COPO_DB.Table("tx_orders").Where("`type` = ? AND `status` = ? ", constants.ORDER_TYPE_DF, constants.WAIT_PROCESS).Find(&orders).Error; err != nil {
		logx.Info("Err", err.Error())
	}

	updateOrders := orders

	logx.Infof("执行时间：%s，待处理-[代付提单]，共 %d 笔", time.Now().Format("2006-01-02 15:04:05"), len(orders))
	if len(updateOrders) > 0 {
		logx.Infof("已处理-[代付提单updateStatusByScheduleBOs]，共 %d 笔", len(updateOrders))
		var IDs []int64
		helper.COPO_DB.Table("tx_orders").Select("id").Where("`type` = ? AND `status` = ? ", constants.ORDER_TYPE_DF, constants.WAIT_PROCESS).Find(&IDs)

		if errUpdate := helper.COPO_DB.Table("tx_orders").Where("id IN (?)", IDs).Updates(map[string]interface{}{"status": "1", "updated_at": time.Now().UTC()}); errUpdate != nil {
			logx.Info(errUpdate.Error)
			logx.Infof("排程发送前先更新订单状态用，待处理 => 处理中 >> 更新异常，ERR: %v ", errUpdate)
		} else {
			logx.Info("已完成更新")
		}
	}

}
