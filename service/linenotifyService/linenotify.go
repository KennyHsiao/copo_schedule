package linenotifyService

import (
	"context"
	"fmt"
	"github.com/copo888/copo_schedule/common/utils"
	"github.com/copo888/transaction_service/common/errorz"
	"github.com/copo888/transaction_service/common/response"
	"github.com/gioco-play/gozzle"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
)

func DoCallLineSendURL (ctx context.Context, message string) error {
	span := trace.SpanFromContext(ctx)
	notifyUrl := fmt.Sprintf("%s:%d/line/send", viper.GetString("LINE_HOST"), viper.GetInt("LINE_PORT"))
	data := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}

	lineKey, errk := utils.MicroServiceEncrypt(viper.GetString("LINE_KEY"), viper.GetString("PUBLIC_KEY"))
	if errk != nil {
		logx.WithContext(ctx).Errorf("MicroServiceEncrypt: %s", errk.Error())
		return errorz.New(response.GENERAL_EXCEPTION, errk.Error())
	}

	res, errx := gozzle.Post(notifyUrl).Timeout(20).Trace(span).Header("authenticationLineKey", lineKey).JSON(data)
	if res != nil {
		logx.WithContext(ctx).Info("response Status:", res.Status())
		logx.WithContext(ctx).Info("response Body:", string(res.Body()))
	}
	if errx != nil {
		logx.WithContext(ctx).Errorf("call Channel cha: %s", errx.Error())
		return errorz.New(response.GENERAL_EXCEPTION, errx.Error())
	} else if res.Status() != 200 {
		return errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("call channelApp httpStatus:%d", res.Status()))
	}

	return nil
}