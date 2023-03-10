package cronjob

import (
	"context"
	"fmt"
	"github.com/copo888/copo_schedule/common/constants"
	"github.com/copo888/copo_schedule/common/model/vo"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	service "github.com/copo888/copo_schedule/service/asyncService"
	merchantService "github.com/copo888/copo_schedule/service/merchantService"
	"github.com/copo888/transaction_service/rpc/transaction"
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

	logx.WithContext(l.ctx).Infof("执行时间：%s", time.Now().Format("2006-01-02 15:04:05"))

	p := service.NewProxyPayEvent(l.ctx)
	if err := helper.COPO_DB.Table("tx_orders").Where("`type` = ? AND `status` = ? ", constants.ORDER_TYPE_DF, constants.WAIT_PROCESS).
		Where("TIMEDIFF(CURRENT_TIMESTAMP(), TIMESTAMPADD(MINUTE,480,DATE_FORMAT(created_at,'%Y-%m-%d %T'))) < 300").
		Find(&orders).Error; err != nil {
		logx.WithContext(l.ctx).Info("Err", err.Error())
	}
	logx.WithContext(l.ctx).Infof("待处理-[代付提单]，共 %d 笔", len(orders))
	if len(orders) > 0 {
		logx.WithContext(l.ctx).Infof("已处理-[代付提单updateStatusByScheduleBOs]，共 %d 笔", len(orders))
		//呼叫渠道送出(異部處理)
		wg := &sync.WaitGroup{}
		wg.Add(len(orders))
		for _, order := range orders {
			//没有回调渠道订单号，以及渠道回调时间/交易时间，才会发送到渠道
			if order.ChannelOrderNo == "" && order.ChannelCallBackAt.IsZero() && order.TransAt.Time().IsZero() {
				channel := types.ChannelData{}
				if queryErr := helper.COPO_DB.Table("ch_channels").Where("code = ?", order.ChannelCode).Find(&channel); queryErr != nil {
					logx.WithContext(l.ctx).Error("queryErr: ", queryErr)
				}

				url := fmt.Sprintf("%s:%s/api/proxy-pay-query", viper.Get("CHANNEL_HOST"), channel.ChannelPort)
				logx.WithContext(l.ctx).Infof("發送代付處理請求To渠道: %v。 url: %s", order, url)

				//異步調用-呼叫異步調用服務
				resp := &vo.ProxyQueryRespVO{}
				var err error
				go func() {
					if resp, err = p.AsyncProxyQueryEvent(&l.ctx, url, &order, wg); err != nil {
						logx.WithContext(l.ctx).Errorf("orderNo: %s ,resp: %s", order.OrderNo, err)
					}
					logx.WithContext(l.ctx).Infof("resp:%+v", resp)

					if err != nil || resp.Code != "0" { ////回复失败，都列为失败单，不管是否为网路异常....等
						//1. 处理渠道返回错误讯习
						if resp != nil {
							logx.WithContext(l.ctx).Errorf("代付提单: %s ，渠道交易失败讯息: %s", order.OrderNo, resp.Message)
							order.Status = constants.FAIL
							order.RepaymentStatus = constants.REPAYMENT_WAIT
							order.ErrorType = "1" //1.渠道返回错误	2.渠道异常	3.商户参数错误	4.账户为黑名单	5.其他
							order.ErrorNote = "Code:" + resp.Code + " Message: " + resp.Message
						}

						if err != nil {
							logx.WithContext(l.ctx).Errorf("代付提单: %s ，渠道交易失败讯息Err: %s", order.OrderNo, err.Error())
							order.Status = constants.FAIL
							order.RepaymentStatus = constants.REPAYMENT_WAIT
							order.ErrorType = "1" //1.渠道返回错误	2.渠道异常	3.商户参数错误	4.账户为黑名单	5.其他
							order.ErrorNote += " Err: " + err.Error()
						}

						// 2. 处理还款
						var resRpc *transaction.ProxyPayFailResponse
						var errRpc error
						//呼叫RPCc还款
						balanceType, errBalance := merchantService.GetBalanceType(helper.COPO_DB, order.ChannelCode, order.Type)
						if errBalance != nil {
							logx.Errorf("BalanceType Err: %s", errBalance.Error())
						}
						rpc := helper.TransactionRpc
						if balanceType == "DFB" {
							resRpc, errRpc = rpc.ProxyOrderTransactionFail_DFB(context.Background(), &transaction.ProxyPayFailRequest{
								MerchantCode: order.MerchantCode,
								OrderNo:      order.OrderNo,
							})
						} else if balanceType == "XFB" {
							resRpc, errRpc = rpc.ProxyOrderTransactionFail_XFB(context.Background(), &transaction.ProxyPayFailRequest{
								MerchantCode: order.MerchantCode,
								OrderNo:      order.OrderNo,
							})
						}

						if errRpc != nil {
							logx.WithContext(l.ctx).Errorf("代付提单 %s 还款失败。 Err: %s", order.OrderNo, errRpc.Error())
							order.RepaymentStatus = constants.REPAYMENT_FAIL

							// 更新订单
							if errUpdate := helper.COPO_DB.Table("tx_orders").Updates(order).Error; errUpdate != nil {
								logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
							}

						} else {
							logx.WithContext(l.ctx).Infof("代付還款rpc完成，%s 錢包還款完成: %#v", balanceType, resRpc)
							order.RepaymentStatus = constants.REPAYMENT_SUCCESS
						}

					} else {
						//若成功则更新交易中、無須還款
						order.Status = constants.TRANSACTION
						order.RepaymentStatus = constants.REPAYMENT_NOT
					}

					// 更新订单
					if errUpdate := helper.COPO_DB.Table("tx_orders").Updates(&order).Error; errUpdate != nil {
						logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
					}
				}()
			}
		}
		wg.Wait()
		logx.WithContext(l.ctx).Info("WaitGroup Finished")
	}
}
