package service

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryChannelBalanceEvent struct {
	logx.Logger
	ctx context.Context
}

func NewQueryChannelBalance(ctx context.Context) QueryChannelBalanceEvent {
	return QueryChannelBalanceEvent{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
	}
}
