package telegramNotify

import (
	"context"
	"fmt"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/gioco-play/gozzle"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
)

func CallTelegramNotify(ctx context.Context, msg *types.TelegramNotifyRequest) error {
	url := fmt.Sprintf("%s:20003/telegram/notify", viper.GetString("SERVER"))
	span := trace.SpanFromContext(ctx)
	if _, err := gozzle.Post(url).Timeout(25).Trace(span).JSON(msg); err != nil {
		logx.WithContext(ctx).Errorf("馀额报警通知失敗:%s", err.Error())
	}

	return nil
}
