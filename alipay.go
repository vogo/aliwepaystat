// Copyright 2019 vogo. All rights reserved.

package aliwepaystat

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	AlipayCsvHeader   = "交易号,商家订单号,交易创建时间,付款时间,最近修改时间,交易来源地,类型,交易对方,商品名称,金额（元）,收/支,交易状态,服务费（元）,成功退款（元）,备注,资金状态,"
	AlipayCsvFieldNum = 17
)

// AlipayTrans alipay transaction
type AlipayTrans struct {
	ID           string  `json:"id" csv:"id" comment:"交易单号"`
	OrderID      string  `json:"order_id" csv:"order_id" comment:"商户单号"`
	CreatedTime  string  `json:"created_time" csv:"created_time" comment:"交易创建时间"`
	PaidTime     string  `json:"paid_time" csv:"paid_time" comment:"付款时间"`
	ModifiedTime string  `json:"modified_time" csv:"modified_time" comment:"最近修改时间"`
	Source       string  `json:"source" csv:"source" comment:"交易来源地"`
	Type         string  `json:"type" csv:"type" comment:"类型"`
	Target       string  `json:"target" csv:"target" comment:"交易对方"`
	Product      string  `json:"product" csv:"product" comment:"商品名称"`
	Amount       float64 `json:"amount" csv:"amount" comment:"金额"`
	FinType      string  `json:"fin_type" csv:"fin_type" comment:"收/支"`
	Status       string  `json:"status" csv:"status" comment:"交易状态"`
	Charge       float64 `json:"charge" csv:"charge" comment:"服务费（元）"`
	Refund       float64 `json:"refund" csv:"refund" comment:"成功退款（元）"`
	Comment      string  `json:"comment" csv:"comment" comment:"备注"`
	FundStatus   string  `json:"fund_status" csv:"fund_status" comment:"资金状态"`
	Other        string  `json:"other" csv:"other" comment:"其他"`
}

func (t *AlipayTrans) IsIncome() bool {
	return Contains(t.FinType, "收入") || ContainsAny(t.Product, cfg.IncomeKeyWords...)
}

func (t *AlipayTrans) IsInnerTransfer() bool {
	return ContainsAny(t.Target, cfg.FamilyMembers...) ||
		ContainsAny(t.Product, cfg.InnerTransferKeyWords...)
}

func (t *AlipayTrans) IsTransfer() bool {
	return Contains(t.FundStatus, "资金转移") ||
		EitherContainsAny(t.Product, t.Target, cfg.TransferKeyWords...) ||
		ContainsAny(t.Target, cfg.FamilyMembers...)
}

func (t *AlipayTrans) IsClosed() bool {
	return ContainsAny(t.Status, "失败", "交易关闭")
}

func (t *AlipayTrans) YearMonth() string {
	if t.ID[:2] != "20" {
		return "20" + t.ID[:4]
	}
	return t.ID[:6]
}

func (t *AlipayTrans) GetID() string           { return t.ID }
func (t *AlipayTrans) GetOrderID() string      { return t.OrderID }
func (t *AlipayTrans) GetCreatedTime() string  { return t.CreatedTime }
func (t *AlipayTrans) GetPaidTime() string     { return t.PaidTime }
func (t *AlipayTrans) GetModifiedTime() string { return t.ModifiedTime }
func (t *AlipayTrans) GetSource() string       { return t.Source }
func (t *AlipayTrans) GetType() string         { return t.Type }
func (t *AlipayTrans) GetTarget() string       { return t.Target }
func (t *AlipayTrans) GetProduct() string      { return t.Product }
func (t *AlipayTrans) GetAmount() float64      { return t.Amount }

func (t *AlipayTrans) GetFormatAmount() float64 {
	return RoundFloat(t.GetAmount())
}
func (t *AlipayTrans) GetFinType() string    { return t.FinType }
func (t *AlipayTrans) GetStatus() string     { return t.Status }
func (t *AlipayTrans) GetCharge() float64    { return t.Charge }
func (t *AlipayTrans) GetRefund() float64    { return t.Refund }
func (t *AlipayTrans) GetComment() string    { return t.Comment }
func (t *AlipayTrans) GetFundStatus() string { return t.FundStatus }
func (t *AlipayTrans) IsShowInList() bool    { return t.GetAmount() > cfg.ListMinAmount }

func NewAlipayTrans() Trans {
	return &AlipayTrans{}
}

type alipayTransParser struct {
}

func (p *alipayTransParser) NewTrans() Trans {
	return &AlipayTrans{}
}
func (p *alipayTransParser) CsvHeader() string {
	return AlipayCsvHeader
}

func (p *alipayTransParser) FieldNum() int {
	return AlipayCsvFieldNum
}

func (p *alipayTransParser) Enc() encoding.Encoding {
	return simplifiedchinese.GBK
}

var TransParserAlipay = &alipayTransParser{}
