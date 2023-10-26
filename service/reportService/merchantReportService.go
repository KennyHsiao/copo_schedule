package reportService

import (
	"context"
	"github.com/copo888/copo_schedule/common/model"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/copo_schedule/common/utils"
	"github.com/copo888/transaction_service/common/errorz"
	"github.com/copo888/transaction_service/common/response"
	"gorm.io/gorm"
)

func InterMerchantReport(db *gorm.DB, req *types.MerchantReportQueryRequest, ctx context.Context) (resp *types.MerchantReportQueryResponse, err error) {

	resp = &types.MerchantReportQueryResponse{}
	//var terms []string
	//var count int64
	var reportList []types.MerchantReport

	//代理商戶查詢
	if req.IsProxySearch == "1" {
		var merchantCods []string
		var merchant *types.Merchant
		var subAgentMerchants []types.Merchant
		// 取得商戶
		if merchant, err = model.NewMerchant(db).GetMerchantByCode(req.MerchantCode); err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 若有層級編號 取得下級商戶
		if merchant.AgentLayerCode != "" {
			if subAgentMerchants, err = model.NewMerchant(db).GetDescendantAgents(merchant.AgentLayerCode, true); err != nil {
				return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
			}
			for _, agentMerchant := range subAgentMerchants {
				merchantCods = append(merchantCods, agentMerchant.Code)
			}
		} else {
			merchantCods = append(merchantCods, merchant.Code)
		}

		db = db.Where("tx.`merchant_code` IN ? ", merchantCods)
	} else if req.IsProxySearch == "0" && len(req.MerchantCode) > 0 {
		db = db.Where("tx.`merchant_code` like ?", "%"+req.MerchantCode+"%")
	}
	if len(req.ChannelCode) > 0 {
		db = db.Where("(ch.code like ? or chxf.code like ?' )", "%"+req.CurrencyCode+"%", "%"+req.ChannelCode+"%")
	}
	if len(req.ChannelName) > 0 {
		db = db.Where("(ch.`name` like ? or chxf.`name` like ?)", "%"+req.ChannelName+"%", "%"+req.ChannelName+"%")
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where("tx.currency_code = ?", req.CurrencyCode)
	}
	if len(req.TransactionType) > 0 {
		db = db.Where("tx.type = ?", req.TransactionType)
	} else {
		db = db.Where("tx.type in ('NC','ZF','DF','XF') ")
	}

	//endAt := utils.ParseTimeAddOneSecond(req.EndAt)

	db = db.Where("tx.`created_at` >= ?", req.StartAt)
	db = db.Where("tx.`created_at` < ?", req.EndAt)
	db = db.Where("(ch.code is not null or chxf.code is not null )")
	db = db.Where("tx.is_test != '1'")
	selectX := "DATE_ADD(DATE_FORMAT(tx.created_at, '%Y-%m-%d %H:00:00') ,INTERVAL 8 HOUR) AS settlement_time," +
		"tx.`merchant_code` AS merchant_code," +
		"CASE WHEN tx.type = 'XF' THEN chxf.`code` ELSE ch.`code` END AS channel_code, " + // 下發的渠道不同表
		"CASE WHEN tx.type = 'XF' THEN chxf.`name` ELSE ch.`name` END AS channel_name, " + // 下發的渠道不同表
		"tx.`type`            AS transaction_type," +
		"tx.currency_code     AS currency_code," +
		"pt.`name`            AS pay_type_name," +
		"mcr.fee              AS merchant_fee," +
		"CASE WHEN tx.`type` = 'XF' AND (mcr.handling_fee = 0 OR mcr.handling_fee IS NULL) THEN bsr.withdraw_handling_fee ELSE mcr.handling_fee END AS merchant_handling_fee," +
		"cpt.fee              AS channel_fee," +
		"cpt.handling_fee     AS channel_handling_fee," +
		"SUM(tx.order_amount) AS order_amount," + //訂單總額
		"COUNT(*)             AS order_quantity," + //訂單數量
		"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN oc.order_amount ELSE IF(tx.actual_amount != 0 ,tx.actual_amount,tx.order_amount) END ELSE 0 END) AS success_amount," + //成功總額
		"SUM(CASE WHEN tx.`status` = '20' THEN 1 ELSE 0 END) AS success_quantity," + //成功數量
		"floor(SUM(CASE WHEN tx.`status` = '20' THEN 1 ELSE 0 END) / GREATEST(count(*),1)*100) AS success_rate," + //成功率
		"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN oc.handling_fee ELSE ofp.transfer_handling_fee END ELSE 0 END) AS system_cost," + //系統成本 ps.下發單看(tx_order_channels.handling_fee) 其他單看(tx_orders_fee_profit.transfer_handling_fee)
		"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN ofp.profit_amount/cc.channel_count ELSE ofp.profit_amount END ELSE 0 END) AS system_profit " //系統利潤 ps.(下發要除該訂單的渠道數量)

	//selectTotal := "SUM(tx.order_amount) AS total_order_amount," + //訂單總額
	//	"COUNT(*) AS total_order_quantity," + //訂單數量
	//	"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN oc.order_amount ELSE IF(tx.actual_amount != 0 ,tx.actual_amount,tx.order_amount) END ELSE 0 END) AS total_success_amount," + //成功總額
	//	"SUM(CASE WHEN tx.`status` = '20' THEN 1 ELSE 0 END) AS total_success_quantity," + //成功數量
	//	"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN oc.handling_fee ELSE ofp.transfer_handling_fee END ELSE 0 END) AS total_cost," +
	//	"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN ofp.profit_amount/cc.channel_count ELSE ofp.profit_amount END ELSE 0 END) AS total_profit "

	//termWhere := strings.Join(terms, " AND ")

	tx := db.
		Table("tx_orders AS tx ").
		Joins("LEFT JOIN tx_orders_fee_profit ofp ON ofp.order_no = tx.order_no and ofp.merchant_code = '00000000' ").
		Joins("LEFT JOIN mc_merchant_channel_rate mcr ON mcr.merchant_code = tx.merchant_code and mcr.channel_pay_types_code = tx.channel_pay_types_code ").
		Joins("LEFT JOIN ch_channel_pay_types cpt ON cpt.`code` = tx.channel_pay_types_code ").
		Joins("LEFT JOIN ch_pay_types pt ON tx.pay_type_code = pt.`code` ").
		Joins("LEFT JOIN ch_channels ch ON tx.channel_code = ch.`code` ").
		Joins("LEFT JOIN tx_order_channels oc ON tx.order_no = oc.order_no "). // 下發用的渠道
		Joins("LEFT JOIN ch_channels chxf ON oc.channel_code = chxf.`code` "). // 下發用的渠道
		Joins("LEFT JOIN ( SELECT order_no, count(*) AS channel_count FROM tx_order_channels GROUP BY order_no ) cc ON cc.order_no = tx.order_no ").
		Joins("LEFT JOIN bs_system_rate bsr ON tx.currency_code = bsr.currency_code")

	groupX := "tx.type, ch.code, chxf.code, pt.code, tx.merchant_code, settlement_time"
	orderX := "settlement_time ASC,currency_code ASC,merchant_code ASC,transaction_type DESC"

	if req.GroupType == "merchantCode" {
		groupX = "tx.merchant_code"
	} else if req.GroupType == "orderType" {
		groupX = "tx.type, tx.merchant_code"
	}

	//if err = tx.Select(selectTotal).Find(resp).Error; err != nil {
	//	return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	//}

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	//if err = tx.Group(groupX).Count(&count).Error; err != nil {
	//	return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	//}

	if err = tx.Select(selectX).Group(groupX).Order(orderX).
		Find(&reportList).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	totalOrderAmount := 0.0
	totalOrderQuantity := 0.0
	totalSuccessAmount := 0.0
	totalSuccessQuantity := 0.0
	totalCost := 0.0
	totalProfit := 0.0

	for _, report := range reportList {
		totalOrderAmount += report.OrderAmount
		totalOrderQuantity += report.OrderQuantity
		totalSuccessAmount += report.SuccessAmount
		totalSuccessQuantity += report.SuccessQuantity
		totalCost += report.SystemCost
		totalProfit += report.SystemProfit
	}

	//=============================================================================
	resp.List = reportList
	resp.PageNum = req.PageNum
	resp.PageSize = req.PageSize
	resp.TotalOrderAmount = totalOrderAmount
	resp.TotalOrderQuantity = totalOrderQuantity
	resp.TotalSuccessAmount = totalSuccessAmount
	resp.TotalSuccessQuantity = totalSuccessQuantity
	resp.TotalCost = totalCost
	resp.TotalProfit = totalProfit
	return
}

func InterMerchantReport2(db *gorm.DB, req *types.MerchantReportQueryRequest, ctx context.Context) (resp *types.MerchantReportQueryResponse, err error) {

	resp = &types.MerchantReportQueryResponse{}
	//var terms []string
	//var count int64
	var queryParam []interface{}
	var WhereMerchantCodes string
	var WhereTransactionTypes string
	var WhereChannelName string
	var Orders string
	var reportList []types.MerchantReport
	var merchantCods []string
	endAt := utils.ParseTimeAddOneSecond(req.EndAt)
	if ctx != nil {
		db = db.WithContext(ctx)
	}

	//代理商戶查詢
	if req.IsProxySearch == "1" {
		//代理查詢商戶號必填，避免代理底下商戶重複查詢
		if req.MerchantCode == "" {
			return nil, errorz.New(response.DATABASE_FAILURE)
		}

		var merchant *types.Merchant
		var subAgentMerchants []types.Merchant
		// 取得商戶
		if merchant, err = model.NewMerchant(db).GetMerchantByCode(req.MerchantCode); err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 若有層級編號 取得下級商戶
		if merchant.AgentLayerCode != "" {
			if subAgentMerchants, err = model.NewMerchant(db).GetDescendantAgents(merchant.AgentLayerCode, true); err != nil {
				return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
			}
			for _, agentMerchant := range subAgentMerchants {
				merchantCods = append(merchantCods, agentMerchant.Code)
			}
		} else {
			merchantCods = append(merchantCods, merchant.Code)
		}

	} else if req.IsProxySearch == "0" && len(req.MerchantCode) > 0 {
		//單一商戶查詢
		merchantCods = append(merchantCods, req.MerchantCode)
		//db = db.Where("AND tx.`merchant_code` LIKE ? ", req.MerchantCode)
	}

	queryParam = append(queryParam, endAt, req.StartAt, endAt, req.CurrencyCode)
	if len(req.TransactionType) > 0 {
		WhereTransactionTypes = `AND tx.type = ? `
		queryParam = append(queryParam, req.TransactionType)
	} else {
		WhereTransactionTypes = `AND tx.type IN ('NC', 'ZF', 'DF', 'XF') `
	}
	if len(merchantCods) > 0 {
		WhereMerchantCodes = `AND tx.merchant_code IN (?) `
		queryParam = append(queryParam, merchantCods)
	}
	if len(req.ChannelName) > 0 {
		WhereChannelName = `AND (ch.name LIKE ? OR chxf.name LIKE ?) `
		queryParam = append(queryParam, "%"+req.ChannelName+"%", "%"+req.ChannelName+"%")
	}

	//if len(req.Orders) > 0 {
	//	Orders = "ORDER BY "
	//	for _, order := range req.Orders {
	//		Orders += order.Column
	//		if order.Asc {
	//			Orders += " ASC,"
	//		} else {
	//			Orders += " DESC,"
	//		}
	//	}
	//	Orders = Orders[:len(Orders)-1]
	//	//	Orders = `ORDER BY  A.agent_layer_code,
	//	//A.transaction_type DESC,
	//	//interval_time`
	//}

	queryParam = append(queryParam, req.StartAt, endAt, endAt)

	db.Raw(`SELECT
        CASE
            WHEN mmrr.created_at IS NOT NULL THEN mmrr.created_at
            ELSE (
                SELECT
                    MIN(created_at)
                FROM
                    mc_merchant_rate_record
                WHERE
                    created_at > A.order_created_at
                    AND created_at <= ?
                    AND merchant_code = A.merchant_code
                    AND channel_pay_type_code = A.channel_pay_types_code
            )
        END AS interval_time,
        A.merchant_code,
		A.agent_layer_code,
        A.channel_name,
		A.channel_code,
        A.transaction_type,
        A.currency_code,
        A.pay_type_name,
		A.pay_type_code,
        A.merchant_fee,
        A.merchant_handling_fee,
        A.channel_fee,
        A.channel_handling_fee,
        A.channel_pay_types_code,
        SUM(A.order_amount) AS order_amount,
        COUNT(*) AS order_quantity,
        SUM(A.success_amount) AS success_amount,
        SUM(A.success_quantity) AS success_quantity,
        FLOOR(SUM(A.success_quantity) / GREATEST(COUNT(A.order_quantity), 1) * 100) AS success_rate,
        SUM(A.system_cost) AS system_cost,
        SUM(A.system_profit) AS system_profit
    FROM
        (
            SELECT
                tx.merchant_code AS merchant_code,
				mm.agent_layer_code AS agent_layer_code,
                CASE
                    WHEN tx.type = 'XF' THEN chxf.code
                    ELSE ch.code
                END AS channel_code,
                CASE
                    WHEN tx.type = 'XF' THEN chxf.name
                    ELSE ch.name
                END AS channel_name,
                tx.type AS transaction_type,
                tx.pay_type_code AS pay_type_code,
                tx.currency_code,
                pt.name AS pay_type_name,
                IF(ofp2.fee !=0 ,ofp2.fee, mcr.fee)  AS merchant_fee,
                IF(ofp2.handling_fee !=0 ,ofp2.handling_fee, mcr.handling_fee) AS merchant_handling_fee,
                cpt.fee AS channel_fee,
                cpt.handling_fee AS channel_handling_fee,
                tx.channel_pay_types_code,
                tx.order_amount AS order_amount,
                1 AS order_quantity,
                tx.created_at AS order_created_at,
                CASE
                    WHEN tx.status = '20' THEN CASE
                        WHEN tx.type = 'XF' THEN oc.order_amount
                        ELSE IF(tx.actual_amount != 0, tx.actual_amount, tx.order_amount)
                    END
                    ELSE 0
                END AS success_amount,
                CASE
                    WHEN tx.status = '20' THEN 1
                    ELSE 0
                END AS success_quantity,
                CASE
                    WHEN tx.status = '20' THEN CASE
                        WHEN tx.type = 'XF' THEN oc.handling_fee
                        ELSE ofp.transfer_handling_fee
                    END
                    ELSE 0
                END AS system_cost,
                CASE
                    WHEN tx.status = '20' THEN CASE
                        WHEN tx.type = 'XF' THEN ofp.profit_amount / cc.channel_count
                        ELSE ofp.profit_amount
                    END
                    ELSE 0
                END AS system_profit
            FROM
                tx_orders AS tx
                LEFT JOIN tx_orders_fee_profit ofp ON ofp.order_no = tx.order_no
                AND ofp.merchant_code = '00000000'
				LEFT JOIN tx_orders_fee_profit ofp2 ON ofp2.order_no = tx.order_no 
				AND ofp2.merchant_code = tx.merchant_code
                LEFT JOIN mc_merchant_channel_rate mcr ON mcr.merchant_code = tx.merchant_code
                AND mcr.channel_pay_types_code = tx.channel_pay_types_code
				LEFT JOIN mc_merchants mm ON tx.merchant_code = mm.code
                LEFT JOIN ch_channel_pay_types cpt ON cpt.code = tx.channel_pay_types_code
                LEFT JOIN ch_pay_types pt ON tx.pay_type_code = pt.code
                LEFT JOIN ch_channels ch ON tx.channel_code = ch.code
                LEFT JOIN tx_order_channels oc ON tx.order_no = oc.order_no
                LEFT JOIN ch_channels chxf ON oc.channel_code = chxf.code
                LEFT JOIN (
                    SELECT
                        order_no,
                        count(*) AS channel_count
                    FROM
                        tx_order_channels
                    GROUP BY
                        order_no
                ) cc ON cc.order_no = tx.order_no
            WHERE
                tx.created_at >= ?
                AND tx.created_at < ?
                AND (ch.code IS NOT NULL OR chxf.code IS NOT NULL)
                AND tx.is_test != '1'
                AND tx.currency_code = ? `+
		WhereTransactionTypes+
		WhereMerchantCodes+
		WhereChannelName+
		` ) AS A
    LEFT JOIN (
        SELECT
            DISTINCT created_at,
            merchant_code,
            channel_pay_type_code
        FROM
            mc_merchant_rate_record
        WHERE
            created_at >= ?
            AND created_at < ?
    ) AS mmrr ON mmrr.merchant_code = A.merchant_code
    AND A.order_created_at >= mmrr.created_at
    AND A.order_created_at < (
        SELECT
            MIN(created_at)
        FROM
            mc_merchant_rate_record
        WHERE
            created_at > mmrr.created_at
            AND created_at <= ?
            AND merchant_code = mmrr.merchant_code
            AND channel_pay_type_code = mmrr.channel_pay_type_code
    )
	GROUP BY A.transaction_type,
	A.channel_code,
	A.pay_type_code,
    A.merchant_code,
    interval_time `+Orders+`;`, queryParam...).Scan(&reportList)

	if err = db.Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	totalOrderAmount := 0.0
	totalOrderQuantity := 0.0
	totalSuccessAmount := 0.0
	totalSuccessQuantity := 0.0
	totalCost := 0.0
	totalProfit := 0.0

	for _, report := range reportList {
		totalOrderAmount += report.OrderAmount
		totalOrderQuantity += report.OrderQuantity
		totalSuccessAmount += report.SuccessAmount
		totalSuccessQuantity += report.SuccessQuantity
		totalCost += report.SystemCost
		totalProfit += report.SystemProfit
	}

	//=============================================================================
	resp.List = reportList
	resp.PageNum = req.PageNum
	resp.PageSize = req.PageSize
	resp.TotalOrderAmount = totalOrderAmount
	resp.TotalOrderQuantity = totalOrderQuantity
	resp.TotalSuccessAmount = totalSuccessAmount
	resp.TotalSuccessQuantity = totalSuccessQuantity
	resp.TotalCost = totalCost
	resp.TotalProfit = totalProfit
	return

}
