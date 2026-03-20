// Copyright 2019 vogo. All rights reserved.

package aliwepaystat

import (
	"log"
	"strconv"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

const (
	WechatCsvHeader   = "交易时间,交易类型,交易对方,商品,收/支,金额(元),支付方式,当前状态,交易单号,商户单号,备注"
	WechatCsvFieldNum = 11
)

// WechatTrans Wechat transaction
type WechatTrans struct {
	CreatedTime string  `json:"created_time" comment:"交易时间"`
	Type        string  `json:"type" comment:"交易类型"`
	Target      string  `json:"target" comment:"交易对方"`
	Product     string  `json:"product" comment:"商品"`
	FinType     string  `json:"fin_type" comment:"收/支"`
	Amount      string  `json:"amount" comment:"金额"`
	Amt         float64 `json:"amt" comment:"金额"`
	Source      string  `json:"source" comment:"支付方式"`
	Status      string  `json:"status" comment:"当前状态"`
	ID          string  `json:"id" comment:"交易单号"`
	OrderID     string  `json:"order_id" comment:"商户单号"`
	Comment     string  `json:"comment" comment:"备注"`
	Refund      float64 `json:"refund" comment:"退款金额"`
}

func (t *WechatTrans) IsIncome() bool {
	return Contains(t.FinType, "收入") ||
		ContainsAny(t.Product, cfg.IncomeKeyWords...)
}

func (t *WechatTrans) IsInnerTransfer() bool {
	return ContainsAny(t.Target, cfg.FamilyMembers...) ||
		ContainsAny(t.Product, cfg.InnerTransferKeyWords...)
}

func (t *WechatTrans) IsTransfer() bool {
	return Contains(t.Type, "转账") ||
		EitherContainsAny(t.Product, t.Target, cfg.TransferKeyWords...) ||
		ContainsAny(t.Target, cfg.FamilyMembers...)
}

func (t *WechatTrans) IsClosed() bool {
	return ContainsAny(t.Status, "失败", "交易关闭")
}

func (t *WechatTrans) YearMonth() string {
	return t.CreatedTime[0:4] + t.CreatedTime[5:7]
}

func (t *WechatTrans) GetID() string          { return t.ID }
func (t *WechatTrans) GetOrderID() string     { return t.OrderID }
func (t *WechatTrans) GetCreatedTime() string { return t.CreatedTime }
func (t *WechatTrans) GetSource() string      { return t.Source }
func (t *WechatTrans) GetType() string        { return t.Type }
func (t *WechatTrans) GetTarget() string      { return t.Target }
func (t *WechatTrans) GetProduct() string     { return t.Product }
func (t *WechatTrans) GetAmount() float64 {
	if t.Amt == 0 && t.Amount != "" {
		var err error
		t.Amt, err = strconv.ParseFloat(strings.ReplaceAll(t.Amount, "¥", ""), 64)
		if err != nil {
			log.Fatalf("无法解析金额: %v", t.Amount)
		}
	}
	return t.Amt
}

func (t *WechatTrans) GetFormatAmount() float64 {
	return RoundFloat(t.GetAmount())
}

func (t *WechatTrans) GetFinType() string { return t.FinType }
func (t *WechatTrans) GetStatus() string  { return t.Status }
func (t *WechatTrans) GetRefund() float64 { return t.Refund }
func (t *WechatTrans) GetComment() string { return t.Comment }
func (t *WechatTrans) IsShowInList() bool { return t.GetAmount() > cfg.ListMinAmount }

type wechatTransParser struct {
}

func (p *wechatTransParser) NewTrans() Trans {
	return &WechatTrans{}
}
func (p *wechatTransParser) CsvHeader() string {
	return WechatCsvHeader
}

func (p *wechatTransParser) FieldNum() int {
	return WechatCsvFieldNum
}

func (p *wechatTransParser) Enc() encoding.Encoding {
	return unicode.UTF8
}

func (p *wechatTransParser) ParseRow(fields []string) (Trans, error) {
	return &WechatTrans{
		CreatedTime: fields[0],
		Type:        fields[1],
		Target:      fields[2],
		Product:     fields[3],
		FinType:     fields[4],
		Amount:      fields[5],
		Source:      fields[6],
		Status:      fields[7],
		ID:          fields[8],
		OrderID:     fields[9],
		Comment:     fields[10],
	}, nil
}

var TransParserWechat = &wechatTransParser{}

func IsWechatGroupAAExpense(trans Trans) bool {
	return trans.GetType() == "群收款" && trans.GetFinType() == "支出"
}
