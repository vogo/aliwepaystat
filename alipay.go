// Copyright 2019 vogo. All rights reserved.

package aliwepaystat

import (
	"fmt"
	"strconv"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	AlipayCsvHeader   = "交易号,商家订单号,交易创建时间,付款时间,最近修改时间,交易来源地,类型,交易对方,商品名称,金额（元）,收/支,交易状态,服务费（元）,成功退款（元）,备注,资金状态,"
	AlipayCsvFieldNum = 17
)

// AlipayTrans alipay transaction
type AlipayTrans struct {
	ID           string  `json:"id" comment:"交易单号"`
	OrderID      string  `json:"order_id" comment:"商户单号"`
	CreatedTime  string  `json:"created_time" comment:"交易创建时间"`
	PaidTime     string  `json:"paid_time" comment:"付款时间"`
	ModifiedTime string  `json:"modified_time" comment:"最近修改时间"`
	Source       string  `json:"source" comment:"交易来源地"`
	Type         string  `json:"type" comment:"类型"`
	Target       string  `json:"target" comment:"交易对方"`
	Product      string  `json:"product" comment:"商品名称"`
	Amount       float64 `json:"amount" comment:"金额"`
	FinType      string  `json:"fin_type" comment:"收/支"`
	Status       string  `json:"status" comment:"交易状态"`
	Charge       float64 `json:"charge" comment:"服务费（元）"`
	Refund       float64 `json:"refund" comment:"成功退款（元）"`
	Comment      string  `json:"comment" comment:"备注"`
	FundStatus   string  `json:"fund_status" comment:"资金状态"`
	Other        string  `json:"other" comment:"其他"`
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

func (p *alipayTransParser) ParseRow(fields []string) (Trans, error) {
	t := &AlipayTrans{
		ID:           fields[0],
		OrderID:      fields[1],
		CreatedTime:  fields[2],
		PaidTime:     fields[3],
		ModifiedTime: fields[4],
		Source:       fields[5],
		Type:         fields[6],
		Target:       fields[7],
		Product:      fields[8],
		FinType:      fields[10],
		Status:       fields[11],
		Comment:      fields[14],
		FundStatus:   fields[15],
		Other:        fields[16],
	}
	var err error
	t.Amount, err = strconv.ParseFloat(fields[9], 64)
	if err != nil {
		return nil, fmt.Errorf("parse Amount: %w", err)
	}
	t.Charge, err = strconv.ParseFloat(fields[12], 64)
	if err != nil {
		return nil, fmt.Errorf("parse Charge: %w", err)
	}
	t.Refund, err = strconv.ParseFloat(fields[13], 64)
	if err != nil {
		return nil, fmt.Errorf("parse Refund: %w", err)
	}
	return t, nil
}

var TransParserAlipay = &alipayTransParser{}
