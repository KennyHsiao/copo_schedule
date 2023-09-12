package model

import (
	"fmt"
	"github.com/copo888/copo_schedule/common/errors"
	_ "github.com/copo888/copo_schedule/common/errors"
	"github.com/copo888/copo_schedule/common/types"
	"github.com/copo888/transaction_service/common/errorz"
	"gorm.io/gorm"
	"regexp"
	"strconv"
)

type Merchant struct {
	MyDB  *gorm.DB
	Table string
}

func NewMerchant(mydb *gorm.DB, t ...string) *Merchant {
	table := "mc_merchants"
	if len(t) > 0 {
		table = t[0]
	}
	return &Merchant{
		MyDB:  mydb,
		Table: table,
	}
}

func (m *Merchant) GetMerchant(id int64) (merchant *types.Merchant, err error) {
	err = m.MyDB.Table(m.Table).
		Preload("Users").
		Preload("MerchantBalances").
		Preload("MerchantCurrencies").
		Take(&merchant, id).Error
	return
}
func (m *Merchant) GetMerchantByCode(code string) (merchant *types.Merchant, err error) {
	err = m.MyDB.Table(m.Table).
		Preload("Users").
		Preload("Users.Currencies").
		Preload("MerchantBalances").
		Preload("MerchantCurrencies").
		Where("code = ? ", code).
		Take(&merchant).Error

	return
}

func (m *Merchant) GetDescendantAgentsByCode(merchantCode string, isIncludeItself bool) (merchants []types.Merchant, err error) {
	var merchant *types.Merchant
	if merchant, err = m.GetMerchantByCode(merchantCode); err != nil {
		return
	}
	if merchant.AgentLayerCode == "" {
		return merchants, errorz.New(errors.AGENT_LAYER_NO_GET_ERROR)
	}
	return m.GetDescendantAgents(merchant.AgentLayerCode, isIncludeItself)
}

// GetDescendantAgents 取得所有子孫商戶 (可選擇是否包含自己)
func (m *Merchant) GetDescendantAgents(agentLayerCode string, isIncludeItself bool) (merchants []types.Merchant, err error) {

	db := m.MyDB.Table(m.Table)

	if !isIncludeItself {
		db = db.Where("agent_layer_code != ?", agentLayerCode)
	}

	err = db.Where("agent_layer_code LIKE ?", agentLayerCode+"%").
		Preload("Users").
		Preload("Users.Currencies").
		Order("agent_layer_code").
		Find(&merchants).Error
	return
}

// GetSubAgents 取得所有子商戶
func (m *Merchant) GetSubAgents(code string) (merchants []types.Merchant, err error) {
	err = m.MyDB.Table(m.Table).Where("agent_parent_code = ?", code).
		Order("agent_layer_code").
		Find(&merchants).Error
	return
}

//func (m *Merchant) QueryMerchants(req types.MerchantQueryAllRequestX) (merchants []types.Merchant, count int64, err error) {
//	//var terms []string
//
//	db := m.MyDB.Table(m.Table)
//
//	if len(req.Code) > 0 {
//		//terms = append(terms, fmt.Sprintf("code like '%%%s%%'", req.Code))
//		db = db.Where("code like ?", "%"+req.Code+"%")
//	}
//	if len(req.AgentLayerCode) > 0 {
//		//terms = append(terms, fmt.Sprintf("`agent_layer_code` like '%%%s%%'", req.AgentLayerCode))
//		db = db.Where("`agent_layer_code` like ?", "%"+req.AgentLayerCode+"%")
//	}
//	if len(req.AccountName) > 0 {
//		//terms = append(terms, fmt.Sprintf("`account_name` like '%%%s%%'", req.AccountName))
//		db = db.Where("`account_name` like ?", "%"+req.AccountName+"%")
//	}
//
//	//term := strings.Join(terms, " AND ")
//
//	if err = db.Count(&count).Error; err != nil {
//		return nil, 0, errorz.New(response.DATABASE_FAILURE, err.Error())
//	}
//
//	if err = m.MyDB.Table(m.Table).
//		Preload("MerchantBalances").
//		Preload("Users").
//		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).Find(&merchants).Error; err != nil {
//		return nil, 0, errorz.New(response.DATABASE_FAILURE, err.Error())
//	}
//
//	return
//}

// GetNextMerchantCode 取得最新商戶編號
func (m *Merchant) GetNextMerchantCode() string {
	var code string
	m.MyDB.Table(m.Table).Select("max(code)").Row().Scan(&code)
	reg, _ := regexp.Compile("[^0-9]+")

	if code == "" {
		return "ME00001"
	}
	codeNum, _ := strconv.Atoi(reg.ReplaceAllString(code, ""))
	return "ME" + fmt.Sprintf("%05d", codeNum+1)
}

// GetNextAgentLayerCode 依上級代理 取得最新下級代理層級編號
//func (m *Merchant) GetNextAgentLayerCode(parentMerchant types.Merchant) string {
//	var code string
//
//	m.MyDB.Table(m.Table).Select("max(agent_layer_code)").Where("agent_parent_code = ?", parentMerchant.Code).Row().Scan(&code)
//	reg, _ := regexp.Compile("[^0-9]+")
//
//	if code == "" {
//		return parentMerchant.AgentLayerCode + "001"
//	}
//
//	codeNum, _ := strconv.Atoi(reg.ReplaceAllString(code, ""))
//	return "A" + fmt.Sprintf("%03d", codeNum+1)
//}

func (m *Merchant) GetNextGeneralAgentCode() string {
	var code string
	m.MyDB.Table(m.Table).Select("max(agent_layer_code)").Where("agent_parent_code is null or agent_parent_code = '' ").Row().Scan(&code)
	reg, _ := regexp.Compile("[^0-9]+")
	if code == "" {
		return "A1001"
	}
	codeNum, _ := strconv.Atoi(reg.ReplaceAllString(code, ""))
	return "A" + fmt.Sprintf("%04d", codeNum+1)
}

func (m *Merchant) GetTotalGeneralAgentCode() (merchants []types.Merchant, err error) {
	db := m.MyDB.Table(m.Table)

	err = db.Where("LENGTH(agent_layer_code) = 5").
		Preload("Users").
		Preload("Users.Currencies").
		Order("agent_layer_code").
		Find(&merchants).Error
	return
}

func (m *Merchant) GetPureMerchant() (merchants []types.Merchant, err error) {
	db := m.MyDB.Table(m.Table)

	err = db.Where("agent_layer_code = ''").
		Preload("Users").
		Preload("Users.Currencies").
		Order("agent_layer_code").
		Find(&merchants).Error
	return
}

// GetLowerLevelAgents 取得下层商戶 (可選擇是否包含自己)
func (m *Merchant) GetLowerLevelAgents(agentLayerCode string, isIncludeItself bool) (merchants []types.Merchant, err error) {

	db := m.MyDB.Table(m.Table)

	if !isIncludeItself {
		db = db.Where("agent_layer_code != ?", agentLayerCode)
	}

	err = db.Where("agent_layer_code LIKE ?", agentLayerCode+"%").
		Where("LENGTH(agent_layer_code) BETWEEN 5 AND 8").
		Preload("Users").
		Preload("Users.Currencies").
		Order("agent_layer_code").
		Find(&merchants).Error
	return
}
