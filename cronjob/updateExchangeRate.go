package cronjob

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/helper"
	telegramNotify "github.com/copo888/copo_schedule/service"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"log"
	"net/http"
)

type UpdateExchangeRate struct {
	logx.Logger
	ctx context.Context
}

func (l UpdateExchangeRate) Run() {

	resp, err := http.Get(viper.GetString("OKX_USDT_CNY_URL"))
	if err != nil {
		log.Fatalf("GET failed: %v", err)
		telegramNotify.CallTelegramNotify(l.ctx, &types.TelegramNotifyRequest{
			ChatID:  viper.GetInt("TELEGRAM_NOTIFY_CHAT_ID_FOR_ERROR"),
			Message: err.Error(),
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Read body failed: %v", err)
	}

	var data Response
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatalf("JSON unmarshal failed: %v", err)
	}

	if data.Code != 0 {
		log.Fatalf("API error: code=%d, msg=%s", data.Code, data.Msg)
	}

	price := data.Data.Price
	if err := helper.COPO_DB.Table("bs_system_rate").
		Where("currency_code = 'CNY'").
		Update("u_rate", price).Error; err != nil {
		log.Fatalf("DB update failed: %v", err)
	}

	fmt.Printf("Updated u_exchange_rate to %s for CNY\n", price)
}

type Response struct {
	Code int `json:"code"`
	Data struct {
		Price string `json:"price"`
	} `json:"data"`
	DetailMsg    string `json:"detailMsg"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Msg          string `json:"msg"`
	RequestId    string `json:"requestId"`
}
