package cronjob

import (
	"context"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	"github.com/copo888/copo_schedule/service/reportService"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

type MerhchantReport struct {
	logx.Logger
	ctx context.Context
}

func (l *MerhchantReport) Run() {
	// TODO: 捞取4小时前的单，并将资料写入""， 并新增栏位日结算 "" groupby
	// 捞tx_orders 每日区间 yyyy-MM-dd 16:00:00(+0) ~ yyyy-MM-dd+1 15:59:59(+0)   在这区间，groupby 日其写入 yyyy-MM-dd+1

	// 每小时的五分跑 抓4小时前到3小时前的资料
	//ex: 当下时间yyyy-MM-dd 15:05:00(+0) -> yyyy-MM-dd 11:00:00(+0) ~ yyyy-MM-dd 12:00:00(+0) = yyyy-MM-dd 19:00:00(+8) ~ yyyy-MM-dd 20:00:00(+8)
	//ex2: 当下时间yyyy-MM-dd 19:05:00(+0) -> yyyy-MM-dd 15:00:00(+0) ~ yyyy-MM-dd 16:00:00(+0) = yyyy-MM-dd 23:00:00(+8) ~ yyyy-MM-dd 24:00:00(+8)
	//ex3: 当下时间yyyy-MM-dd 20:05:00(+0) -> yyyy-MM-dd 16:00:00(+0) ~ yyyy-MM-dd 17:00:00(+0) = yyyy-MM-dd+1 00:00:00(+8) ~ yyyy-MM-dd+1 01:00:00(+8)
	//                                               startAt                  endAt               groupByStart                groupByEnd

	startAt := time.Now().UTC().Add(-1 * time.Hour)
	endAt := time.Now().UTC()
	//startAt := time.Now().UTC().Add(-4 * time.Hour)
	//endAt := time.Now().UTC().Add(-3 * time.Hour)
	groupByStart := startAt.Add(8 * time.Hour).Format("2006-01-02 15:04:05")[:13] + ":00:00"
	groupByEnd := endAt.Add(8 * time.Hour).Format("2006-01-02 15:04:05")[:13] + ":00:00"
	logx.WithContext(l.ctx).Infof("商户报表排程开始: %s", time.Now().Format("2006-01-02 15:04:05"))
	logx.WithContext(l.ctx).Infof("商户报表结算开始，开始时间: %s,结速时间: %s", startAt.Format("2006-01-02 15:04:05")[:13]+":00:00", endAt.Format("2006-01-02 15:04:05")[:13]+":00:00")
	println("商户报表结算开始，开始时间: " + startAt.Format("2006-01-02 15:04:05")[:13] + ":00:00" + " 结速时间: " + endAt.Format("2006-01-02 15:04:05")[:13] + ":00:00")
	println("商户报表groupBy，开始时间: " + groupByStart + " 结速时间: " + groupByEnd)
	resp := &types.MerchantReportQueryResponse{}
	var err error
	if resp, err = reportService.InterMerchantReport(helper.COPO_DB, &types.MerchantReportQueryRequest{
		MerchantCode:    "",
		ChannelCode:     "",
		ChannelName:     "",
		StartAt:         startAt.Format("2006-01-02 15:04:05")[:13] + ":00:00",
		EndAt:           endAt.Format("2006-01-02 15:04:05")[:13] + ":00:00",
		TransactionType: "",
		CurrencyCode:    "",
		GroupType:       "",
		IsProxySearch:   "",
	}, l.ctx); err != nil {
		logx.WithContext(l.ctx).Errorf("商户报表结算错误:%s", err.Error())
	}
	if len(resp.List) > 0 {
		var merchantReportCreates []types.MerchantReportCreate

		for _, merReport := range resp.List {
			merReport.SettlementDate = groupByStart[:10]
			merchantReport := types.MerchantReportCreate{
				MerchantReport: merReport,
				CreatedAt:      time.Now().UTC().Format("2006-01-02 15:04:05"),
			}
			merchantReportCreates = append(merchantReportCreates, merchantReport)
		}

		if err := helper.COPO_DB.Table("rp_merchant_report").CreateInBatches(merchantReportCreates, len(merchantReportCreates)).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("商户报表新增结算错误:%s", err.Error())
		}
	}

}
