package cronjob

import (
	"context"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/common/utils"
	"github.com/copo888/copo_schedule/helper"
	telegramNotify "github.com/copo888/copo_schedule/service"
	"github.com/zeromicro/go-zero/core/logx"
	"strings"
)

type MerchantBalancesCheck struct {
	logx.Logger
	ctx context.Context
}

func (l *MerchantBalancesCheck) Run() {
	logx.WithContext(l.ctx).Infof("開始檢查商戶子錢包")
	var merchantPtBalances []types.MerchantPtBalance
	var merchantCurrencies []types.MerchantCurrency

	if err := helper.COPO_DB.Table("mc_merchant_currencies").
		Where("is_display_pt_balance = ?", "1").
		Order("merchant_code").
		Find(&merchantCurrencies).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("取得商户币别錯誤:", err.Error())
	}

	if len(merchantCurrencies) > 0 {
		var msg = "子钱包余额功能异常"
		merchantMap := make(map[string]string)

		for _, currency := range merchantCurrencies {
			var merchantDfbBalance types.MerchantBalance
			var merchantXfbBalance types.MerchantBalance
			merchantCode := currency.MerchantCode
			currencyCode := currency.CurrencyCode

			if err := helper.COPO_DB.Table("mc_merchant_balances").
				Where("merchant_code = ?", merchantCode).
				Where("currency_code = ?", currencyCode).
				Where("balance_type = ?", "DFB").
				Find(&merchantDfbBalance).Error;err != nil {
				logx.WithContext(l.ctx).Errorf("取得商户馀额錯誤:", err.Error())
			}

			if err := helper.COPO_DB.Table("mc_merchant_balances").
				Where("merchant_code = ?", merchantCode).
				Where("currency_code = ?", currencyCode).
				Where("balance_type = ?", "XFB").
				Find(&merchantXfbBalance).Error;err != nil {
				logx.WithContext(l.ctx).Errorf("取得商户馀额錯誤:", err.Error())
			}

			if err := helper.COPO_DB.Table("mc_merchant_pt_balances").
				Where("merchant_code = ?", merchantCode).
				Where("currency_code = ?", currencyCode).
				Find(&merchantPtBalances).Error; err != nil {
				logx.WithContext(l.ctx).Errorf("取得商户子钱包錯誤:", err.Error())
			}

			totalBalance := utils.FloatAdd(merchantDfbBalance.Balance, merchantXfbBalance.Balance)
			var totalPtBalance float64
			for _, balance := range merchantPtBalances {
				totalPtBalance = utils.FloatAdd(totalPtBalance, balance.Balance)
			}


			if totalBalance != totalPtBalance {
				if _, ok := merchantMap[merchantCode]; ok {
					msg += "、"+currencyCode
				}else {
					merchantMap[merchantCode] = merchantCode
					msg += "\n\n商户号："+ merchantCode +"\n币别："+ currencyCode
				}
			}
		}
		if strings.Contains(msg, "子钱包余额功能异常\n\n商户号：") {
			logx.WithContext(l.ctx).Infof("通知商戶子錢包有誤，Msg :", msg)
			telegramNotify.CallNoticeUrlForBalance(l.ctx, &types.TelegramNotifyRequest{
				Message: msg,
			})
		}
	}
	logx.WithContext(l.ctx).Infof("檢查商戶子錢包結束")
}