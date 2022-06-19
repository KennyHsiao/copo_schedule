package dto

/*	schedule呼叫渠道返回用的參數*/
type ProxyPayChannelResultDTO struct {
	Code           string `json:"status"`         //成功狀態  0或1
	Message        string `json:"message"`        /** 渠道訊息 */
	ChannelOrderNo string `json:"channelOrderNo"` /** 渠道订单号 */
}
