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
	ID           string  `csv:"id"`
	OrderID      string  `csv:"order_id"`
	CreatedTime  string  `csv:"created_time"`
	PaidTime     string  `csv:"paid_time"`
	ModifiedTime string  `csv:"modified_time"`
	Source       string  `csv:"source"`
	Type         string  `csv:"type"`
	Target       string  `csv:"target"`
	Product      string  `csv:"product"`
	Amount       float64 `csv:"amount"`
	FinType      string  `csv:"fin_type"`
	Status       string  `csv:"status"`
	Charge       float64 `csv:"charge"`
	Refund       float64 `csv:"refund"`
	Comment      string  `csv:"comment"`
	FundStatus   string  `csv:"fund_status"`
	Other        string  `csv:"other"`
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
