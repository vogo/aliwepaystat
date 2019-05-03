// Copyright 2019 wongoo. All rights reserved.

package aliwepaystat

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"log"
	"strconv"
	"strings"
)

const (
	WechatCsvHeader   = "交易时间,交易类型,交易对方,商品,收/支,金额(元),支付方式,当前状态,交易单号,商户单号,备注"
	WechatCsvFieldNum = 11
)

//WechatTrans Wechat transaction
type WechatTrans struct {
	CreatedTime string `csv:"created_time"`
	Type        string `csv:"type"`
	Target      string `csv:"target"`
	Product     string `csv:"product"`
	FinType     string `csv:"fin_type"`
	Amount      string `csv:"amount"`
	amt         float64
	Source      string `csv:"source"`
	Status      string `csv:"status"`
	ID          string `csv:"id"`
	OrderID     string `csv:"order_id"`
	Comment     string `csv:"comment"`
	refund      float64
}

func (t *WechatTrans) IsIncome() bool {
	return Contains(t.FinType, "收入") || ContainsAny(t.Product, incomeKeyWords...)
}

func (t *WechatTrans) IsInnerTransfer() bool {
	return ContainsAny(t.Target, familyMembers...) || ContainsAny(t.Product, innerTransferKeyWords...)
}

func (t *WechatTrans) IsTransfer() bool {
	return Contains(t.Type, "转账") || EitherContainsAny(t.Product, t.Target, transferKeyWords...) || ContainsAny(t.Target, familyMembers...)
}

func (t *WechatTrans) IsClosed() bool {
	return Contains(t.Status, "交易关闭")
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
	if t.amt == 0 && t.Amount != "" {
		var err error
		t.amt, err = strconv.ParseFloat(strings.ReplaceAll(t.Amount, "¥", ""), 32)
		if err != nil {
			log.Fatalf("无法解析金额: %v", t.Amount)
		}
	}
	return t.amt
}
func (t *WechatTrans) GetFinType() string { return t.FinType }
func (t *WechatTrans) GetStatus() string  { return t.Status }
func (t *WechatTrans) GetRefund() float64 { return t.refund }
func (t *WechatTrans) GetComment() string { return t.Comment }

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

var TransParserWechat = &wechatTransParser{}

func IsWechatGroupAAExpense(trans Trans) bool {
	if we, ok := trans.(*WechatTrans); ok {
		return we.Type == "群收款" && we.FinType == "支出"
	}
	return false
}
