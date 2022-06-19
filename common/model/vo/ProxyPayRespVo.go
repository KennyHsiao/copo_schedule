package vo

type ProxyPayRespVO struct {
	Code    string                  `json:"code"`
	Message string                  `json:"message"`
	Data    ChannelAppProxyResponse `json:"data"`
	traceId string                  `json:"traceId"`
}

//渠道app 代付返回物件
type ChannelAppProxyResponse struct {
	ChannelOrderNo string `json:"channelOrderNo"`
	OrderStatus    string `json:"orderStatus"`
}
