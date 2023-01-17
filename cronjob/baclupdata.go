package cronjob

import (
	"context"
	"fmt"
	"github.com/gioco-play/gozzle"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
)

type BackupData struct {
	logx.Logger
	ctx context.Context
}

func (l *BackupData) Run() {
	logx.WithContext(l.ctx).Info("Backup Data Start")
	span := trace.SpanFromContext(l.ctx)
	backUpUrl := fmt.Sprintf("%s:8080/api/v1/backup/create", viper.GetString("SERVER"))

	res, errx := gozzle.Post(backUpUrl).Timeout(20).Trace(span).JSON("")
	if res != nil {
		logx.WithContext(l.ctx).Info("Channel Balance Record response Status:", res.Status())
		logx.WithContext(l.ctx).Info("Channel Balance Record response Body:", string(res.Body()))
	}
	if errx != nil {
		logx.WithContext(l.ctx).Errorf("Channel Balance Record ERROR: %s", errx.Error())
	}
	logx.WithContext(l.ctx).Info("Channel Balance Record End ")
}