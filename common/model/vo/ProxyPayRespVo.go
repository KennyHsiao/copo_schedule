package vo

type ProxyPayRespVO struct {
	Code    string                  `json:"code"`
	Message string                  `json:"message"`
	Data    ChannelAppProxyResponse `json:"data"`
	traceId string                  `json:"traceId"`
}

type ProxyQueryRespVO struct {
	Code    string                       `json:"code"`
	Message string                       `json:"message"`
	Data    ChannelAppProxyQeuryResponse `json:"data"`
	traceId string                       `json:"traceId"`
}

type ChannelAppProxyQeuryResponse struct {
	Status           int    `json:"status"`
	ChannelOrderNo   string `json:"channelOrderNo"`
	OrderStatus      string `json:"orderStatus"`
	CallBackStatus   string `json:"callBackStatus"`
	ChannelReplyDate string `json:"channelReplyDate"`
	ChannelCharge    int    `json:"channelCharge"`
}

//渠道app 代付返回物件
type ChannelAppProxyResponse struct {
	ChannelOrderNo string `json:"channelOrderNo"`
	OrderStatus    string `json:"orderStatus"`
}
