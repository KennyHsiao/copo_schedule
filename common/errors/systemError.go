package errors

var (
	SUCCESS = "0"     //"操作成功"
	FAIL    = "EX001" //"Fail"

	/**
	 * 通用系统讯息码
	 */
	//SUCCESS = "1030001" // "操作成功")//

	AGENT_LAYER_NO_GET_ERROR = "1022091" //"代理層級編碼取得異常，請恰管理員"

	/**
	 * 前端操作讯息码
	 */
	SERVICE_ERROR                                          = "1032001"  // "抱歉，您的操作出了问题，请联系客服，谢谢")//
	BUSINESS_ERROR                                         = "1032002"  // "服务异常，请联系客服，谢谢")//
	BOSS_TOKEN_TIMEOUT                                     = "1032003"  // "登录逾时，请重新登入")//
	APP_TOKEN_TIMEOUT                                      = "1032004"  // "登入失败，请重新登入")//
	INVALID_PARAMETER                                      = "1032005"  // "无效的参数")//
	PARAMETER_TYPE_ERROE                                   = "1032006"  // "参数类型错误")//
	UN_AUTHORIZE                                           = "1032009"  // "无授权")//
	MISSING_PARAMETER                                      = "1032013"  // "缺少必要参数")//
	FEIGN_CLIENT_ERROR                                     = "1032014"  // "feign client 連接錯誤")//
	REQUEST_FORMAT_ERROR                                   = "1032016"  // "无效传递格式")//
	DATE_RANGE_ERROR                                       = "1032017"  // "无效的时间区间")//
	NOT_SETTING_CHANNEL                                    = "1032018"  // "商户尚未配置代付渠道")//
	PROXY_PAY_NO_MAPPING_DATA                              = "1032019"  // "无此代付提单资料")//
	PROXY_PAY_CHANNEL_NO_MAPPING_DATA                      = "1032020"  // "无此代付提单对应渠道资料")//
	PROXY_PAY_REPAYMENT_FAIL                               = "1032021"  // "代付提单还款失败")//
	PROXY_PAY_IS_CLOSE                                     = "1032022"  // "此提单目前已为结单状态")//
	PROXY_PAY_CALLBACK_FAIL                                = "1032023"  // "回调失败")//
	PROXY_PAY_IS_NOT_REPAYMENT_FAIL                        = "1032024"  // "非还款失败或待还款提单")//
	PROXY_PAY_AMOUNT_MININUM_FAIL                          = "1032025"  // "单笔小于最低代付金额")//
	PROXY_PAY_AMOUNT_MAXINUM_FAIL                          = "1032026"  // "单笔大于最高代付金额")//
	PROXY_PAY_PERSON_PROCESS_FAIL                          = "1032027"  // "人工处里失败")//
	TRANSCATION_DATE_CAN_NOT_GT_TODAY                      = "1032028"  // "交易日期时间不可大于现在时间")//
	TRANSCATION_DATE_FROMAT_FAIL                           = "1032029"  // "交易日期时间格式错误")//
	IS_SUCCESS_ORDER                                       = "1032030"  // "已为交易成功提单或不需人工处理的提单，如有问题请恰系统人员")//
	NOT_PERSON_PROCESS_ORDER                               = "1032031"  // "非人工处理的提单，如有问题请恰系统人员")//
	PROXY_PAY_REPEAT_REPAYMENT                             = "1032032"  // "此提单已有还款记录，如有问题请恰系统人员")//
	PERSON_PROCESS_PROXY_PAY_ERROR                         = "1032033"  // "人工处里异常代付提单异常")//
	CALLBACK_FOR_PROXY_PAY_ERROR                           = "1032034"  // "渠道=代付)回调异常")//
	PROXY_PAY_REPAYMENT_ERROR                              = "1032035"  // "内部排程补还款异常")//
	TRANSACTION_DATE_CAN_NOT_GT_ORDERAT                    = "1032037"  // "交易日期时间不可小于提单申请时间")//
	TRANSACTION_IP_NOT_EQUAL_TO_LOGIN_IP                   = "1032038"  // "出款IP不等于登录IP，无法进行出款，请重新登录，谢谢")//
	AGENT_PROXY_PAY_PROFIT_SUMMARY_ERROR                   = "1032039"  // "代理[代付佣金]结算失败")//
	CURRENCY_NOT_TRANSACTION                               = "10320380" // "不可交易币别，请确认")//
	BATCH_TRANSACTION_NOT_TWO_CURRENCY                     = "1032040"  // "同一笔批量交易，不可存在2种交易币别，请确认")//
	CALLBACK_FOR_PROXY_PAY_SOURCE_ERROR                    = "1032041"  // "回调失败，非API提单或该笔提单位提供回调地址")//
	ORDER_STATUS_CANT_TO_CHANGE_TO_PERSON_PROCESS          = "1032042"  // "提单目前不可变更为人工处里")//
	CHANNEL_CHARGE_ERROR                                   = "1032043"  // "输入的渠道成本手续费不得大于商户及代理手续费")//
	ORDER_STATUS_IS_REPAYMENT_NOT_CHANGE_TO_PERSON_PROCESS = "1032044"  // "提单目前已在等待还款阶段，不可变更")//
	PROXY_PAY_IS_PERSON_PROCESS                            = "1032045"  // "提单目前为人工处里阶段，不可回调变更")//

	/**
	 *  使用者操作错误
	 */
	USER_NOT_FOUND = "1023001" // "用户不存在")//

	/**
	 *  资料库层级错误
	 */
	DATA_NOT_FOUND_IN_DATABASE = "1038001" // "找不到资料")//
	INSERT_REDIS_FAILURE       = "1038003" // "redis-数据新增失败")//
	UPDATE_REAL_NAME_FAILURE   = "1038005" // "真實性名已設置，不可再修改")//
	NO_DATA_TO_UPDATE          = "1038008" // "无数据可更新或提单已是目前状态")//

	/*
	   EEEE: 9000 ~ 9999// 网络层级错误
	*/
	SERVICE_PROVIDER_NOT_FOUND = "1039001" // "找不到服务提供者")//
	GET_TOKEN_EXCEPTION        = "1039002" // "取得token发生异常")//
	SIGN_KEY_FAIL              = "1039003" // "加签错误，请确认加签规则")

	/**
	 * 系统类型讯息码
	 */
	ILLEGAL_REQUEST         = "1062101" //"非法请求"
	ILLEGAL_PARAMETER       = "1062102" //"非法参数"
	UN_ROLE_AUTHORIZE       = "1062103" //"无此应用服务使用授权"
	DATA_NOT_FOUND          = "1062104" //"资料无法取得"
	GENERAL_ERROR           = "1062105" //"通用错误"
	CONNECT_SERVICE_FAILURE = "1062106" //"服务连线失败"
	UPDATE_DATABASE_FAILURE = "1062107" //"数据更新失败"
	UPDATE_DATABASE_REPEAT  = "1062108" //"数据重复更新"
	INSERT_DATABASE_FAILURE = "1062109" //"数据新增失败"
	DELETE_DATABASE_FAILURE = "1062110" //"数据删除失败"
	DATABASE_FAILURE        = "1062111" //"数据库错误"

	CHANNEL_CLOSED_OR_DEFEND                      = "1062112" //"渠道关闭或维护中"
	RATE_NOT_CONFIGURED                           = "1062113" //"未配置商户渠道费率"
	REPLY_MESSAGE_MALFORMED                       = "1062114" //"返回资讯格式错误"
	PROXY_BAL_MIN_LIMIT_NOT_CONFIGURED            = "1062115" //"渠道代付馀额下限值未设定"
	CHN_BALANCE_NOT_ENOUGH                        = "1062116" //"渠道余额不足扣款金额"
	SINGLE_LIMIT_SETTING_ERROR                    = "1062117" //"单笔限额设定错误，最小值必需小于最大值"
	INVALID_USDT_CHANNEL_CODING                   = "1062118" //"无效的USDT渠道编码"
	USDT_CHANNEL_NAME_DIFFERENT                   = "1062119" //"与对应的USDT渠道名称不一致"
	USDT_CHANNEL_REPEAT_DESIGNATION               = "1062120" //"对应的USDT渠道已被指定"
	RESET_DESIGNATION_MER_RATE_ERROR              = "1062121" //"不得重置已指定的商户费率"
	RESETRADIS_FROM_DB_FAILURE                    = "1062122" //"更新Redis資料錯誤"
	INVALID_CHANNEL_INFO                          = "1062123" //"無對應相關渠道資料"
	RATE_SETTING_ERROR                            = "1062124" //"费率设定异常"
	RATE_NOT_CONFIGURED_OR_CHANNEL_NOT_CONFIGURED = "1062125" //"未配置商户渠道费率或渠道配置错误"
	SYSTEM_RATE_NOT_SET                           = "1062126" //"未配置系統費率"

	//api 相關
	API_SUCCESS                 = "000" // "操作成功"
	NO_RECORD_DATA              = "001" // "无记录"
	API_GENERAL_ERROR           = "002" // "系统忙碌中,请稍后再试"
	GENERAL_EXCEPTION           = "003" // "系统錯誤,请稍后再试"
	API_INVALID_PARAMETER       = "004" // "无效参数"
	SERVICE_RESPONSE_ERROR      = "005" // "服务回傳失败"
	SERVICE_RESPONSE_DATA_ERROR = "006" // "服务回傳資料錯誤"
	API_IP_DENIED               = "007" // "此IP非法登錄，請設定白名單"
	ContentType_ERROR           = "008" // "内容类型错误，请使用 application/js5on"
	API_PARAMETER_TYPE_ERROE    = "009" // "JSON格式或参数类型错误"
	WAIT_LOCK_EXCEPTION         = "010" // "此交易目前正执行中，请稍后再试"
	CACHED_DATA_NOT_FOUND       = "011" // "此交易已执行完毕或交易参数错误"
	/**
	 * 参数错误讯息码
	 */
	INVALID_TIMESTAMP                  = "101" // "无效时间戳"
	INVALID_SIGN                       = "102" // "无效验签"
	INVALID_CURRENCY_CODE              = "103" // "无效货币编码"
	INVALID_ORDER_NO                   = "104" // "无效订单编号"
	REPEAT_ORDER_NO                    = "105" // "重复订单号"
	INVALID_START_DATE                 = "106" // "无效开始日期时间"
	INVALID_END_DATE                   = "107" // "无效结束日期时间"
	INVALID_DATE_RANGE                 = "108" // "无效日期范围"
	INVALID_DATE_TYPE                  = "109" // "无效日期筛选类型"
	INVALID_MERCHANT_CODE              = "110" // "无效商户号"
	MERCHANT_IS_DISABLE                = "111" // "商户已被禁用或结清"
	INVALID_AMOUNT                     = "112" // "无效金额"
	INVALID_LANGUAGE_CODE              = "113" // "无效语言编码"
	INVALID_BANK_ID                    = "114" // "无效开户行号"
	INVALID_BANK_NAME                  = "115" // "无效开户行名"
	INVALID_BANK_NO                    = "116" // "无效银行卡号"
	INVALID_DEFRAY_NAME                = "117" // "无效开户人姓名"
	INVALID_ACCESS_TYPE                = "118" // "无效接入类型"
	INVALID_NOTIFY_URL                 = "119" // "无效URL格式"
	SIGN_ERROR                         = "120" // "签名出错"
	NO_AVAILABLE_CHANNEL_ERROR         = "121" // "没有可用通道"
	CHANNEL_NOT_OPEN_OR_NOT_MEET_RULES = "122" // "指定通道没有开启或不符合指定通道规则"
	NO_CHANNEL_SET                     = "123" // "未指定通道或不符合指定通道规则"
	INVALID_MERCHANT_ID                = "124" // "无效商户号"
	INVALID_MERCHANT_AGENT             = "125" // "无效代理商户"
	INVALID_MERCHANT_ACCOUNT           = "126" // "无效商户帐号"
	ERROR_CHANGE_PASSWORD              = "127" // "商户密码变更错误"
	ERROR_CHANGE_MERCHANT_KEY          = "128" // "商户密钥变更错误"
	INVALID_USER_NAME                  = "129" // "开户人姓名无效或未输入"
	INVALID_USDT_TYPE                  = "130" // "无效协议"
	INVALID_USDT_WALLET_ADDRESS        = "131" // "无效钱包地址"
	INVALID_PAY_TYPE_SUB_NO            = "132" // "多指定模式，PayTypeSubNo為必填"

	// for channel test only
	INVALID_MERCHANT_OR_CHANNEL_PAYTYPE = "160" // "資料庫無此商户号或商户未配置渠道、支付方式等設定错误或关闭或维护"
	INVALID_CHANNEL_PAYTYPE_CURRENCY    = "161" // "商户配置之渠道支付方式與幣別有誤"
	/**
	 *  交易错误讯息码
	 */
	TRANSACTION_FAILURE             = "201" // "交易失败"
	INSUFFICIENT_IN_AMOUNT          = "202" // "余额不足"
	CURRENCY_INCONSISTENT           = "203" // "商户货币別不一致"
	IS_LESS_MINIMUN_AMOUNT          = "204" // "单笔小于最低交易金额"
	IS_GREATER_MXNIMUN_AMOUNT       = "205" // "单笔大于最高交易金额"
	MERCHANT_IS_NOT_SETTING_CHANNEL = "206" // "尚未配置渠道"
	BANK_CODE_EMPTY                 = "207" // "银行代码不得为空值"
	BANK_CODE_INVALID               = "208" // "银行代码错误"
	PAY_TYPE_INVALID                = "209" // "支付类型代码错误"
	CHANNEL_REPLY_ERROR             = "210" // "渠道返回错误"
	INVALID_STATUS_CODE             = "211" // "Http状态码错误"
	INVALID_CHANNEL_ORDER_NO        = "212" // "渠道未回传渠道订单号"
	TRANSACTION_PROCESSING          = "213" // "訂單處理中，請稍後"

	/**
	 * 内部错误
	 */
	INTERNAL_SIGN_ERROR = "301" // "系统验签错误"

	/**
	 * 系统层级错误
	 */
	SYSTEM_ERROR  = "400" // "系统错误"
	NETWORK_ERROR = "401" // "网路异常"

	/**
	 * 應用层级错误： 支付 500~599
	 */
	ORDER_NUMBER_EXIST                            = "500" // "商户订单号重复"
	ORDER_NUMBER_NOT_EXIST                        = "501" // "商户订单号不存在"
	MERCHANT_PAY_TYPE_INVALID_OR_CHANNEL_NOT_SET  = "502" // "商户代码[%s]或支付类型代码[%s]或幣別[%s]错误或指定渠道设定错误或关闭或维护"
	ORDER_AMOUNT_INVALID                          = "503" // "商户下单金额错误"
	ORDER_AMOUNT_LIMIT_MIN                        = "504" // "商户下单金额太小"
	ORDER_AMOUNT_LIMIT_MAX                        = "505" // "商户下单金额太大"
	WALLET_NOT_SET                                = "506" // "商户渠道錢包未设定"
	API_MERCHANT_CHANNEL_NOT_SET                  = "507" // "商户渠道未建立"
	MERCHANT_PAY_TYPE_INVALID_OR_CHANNEL_NOT_SET2 = "508" // "商户代码[%s]或支付类型代码[%s][%s]错误或指定渠道设定错误或关闭或维护"
	WALLET_UPDATE_ERROR                           = "509" // "商户錢包資料错误"
	ORDER_AMOUNT_ERROR                            = "510" // "商户下单金额和回調金額不符"
	ORDER_BANK_NO_LEN_ERROR                       = "511" // "银联行账(卡)号，长度必须13~22位"

	/*代付相关错误*/
	//PROXY_PAY_IS_CLOSE              = "600" // "此提单目前已为结单状态"
	//PROXY_PAY_CALLBACK_FAIL         = "601" // "回调失败"
	//PROXY_PAY_IS_NOT_REPAYMENT_FAIL = "602" // "非还款失败或待还款提单"
	//PROXY_PAY_AMOUNT_MININUM_FAIL   = "603" // "单笔小于最低代付金额"
	//PROXY_PAY_AMOUNT_MAXINUM_FAIL   = "604" // "单笔大于最高代付金额"
	//PROXY_PAY_PERSON_PROCESS_FAIL   = "605" // "人工处里失败"
	//PROXY_PAY_IS_PERSON_PROCESS     = "606" // "提单目前为人工处里阶段，不可回调变更"
	PROXY_PAY_AMOUNT_INVALID = "607" // "代付回调金额与订单金额不符"
)
