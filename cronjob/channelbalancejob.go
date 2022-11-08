package cronjob

import (
	"context"
	"fmt"
	"github.com/gioco-play/gozzle"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
)

type ChannelBalance struct {
	logx.Logger
	ctx context.Context
}

func (l *ChannelBalance) Run() {
	span := trace.SpanFromContext(l.ctx)
	notifyUrl := fmt.Sprintf("%s:8080/api/v1/channel_schedule/channelbalance/update", viper.GetString("SERVER"))

	res, errx := gozzle.Post(notifyUrl).Timeout(20).Trace(span).JSON("")
	if res != nil {
		logx.WithContext(l.ctx).Info("response Status:", res.Status())
		logx.WithContext(l.ctx).Info("response Body:", string(res.Body()))
	}
	if errx != nil {
		logx.WithContext(l.ctx).Errorf("call Channel cha: %s", errx.Error())
	} else if res.Status() != 200 {
		logx.WithContext(l.ctx).Errorf("call channelApp httpStatus:%d", res.Status())
	}

}