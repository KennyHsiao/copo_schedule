package cronjob

import (
	"context"
	"fmt"
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	telegramNotify "github.com/copo888/copo_schedule/service"
	"github.com/zeromicro/go-zero/core/logx"
)

type NotifyProxyOrder struct {
	logx.Logger
	ctx context.Context
}

func (l *NotifyProxyOrder) Run() {
	var orders []types.OrderX
	var systemParam types.SystemParams
	if err := helper.COPO_DB.Table("bs_system_params").Where("name = ?", "proxyNotifyMin").Take(&systemParam).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("bs_system_params Err : %s", err.Error())
	}

	//1.取出代付提单的订单状态[3：交易中]的提单 2. created_at - currentTime >5 min
	if err := helper.COPO_DB.Table("tx_orders").
		Where("`type` = ? AND `status` = ?", constants.ORDER_TYPE_DF, constants.TRANSACTION).
		Where("TIMESTAMPADD(DAY, " + fmt.Sprintf("-%s", systemParam.Value) + ",DATE_FORMAT(CURRENT_TIMESTAMP(),'%Y-%m-%d %T')) < created_at").
		Find(&orders).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("Err : %s", err.Error())
	}

	if len(orders) > 0 {
		var msg string
		msg = "代付超過五分鐘未回调的提单: \n"
		for _, order := range orders {
			msg += fmt.Sprintf("%s:%f \n", order.OrderNo, order.OrderAmount)
		}

		telegramNotify.CallTelegramNotify(l.ctx, &types.TelegramNotifyRequest{
			Message: msg,
		})
	}

	logx.WithContext(l.ctx).Infof("启动处理交易中，未回调的提单，笔数：%d 笔", len(orders))
}
