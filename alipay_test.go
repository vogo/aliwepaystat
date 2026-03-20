package aliwepaystat

import (
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestAlipayTransGetters(t *testing.T) {
	trans := &AlipayTrans{
		ID:           "2024011234567890",
		OrderID:      "ORD001",
		CreatedTime:  "2024-01-12 10:00:00",
		PaidTime:     "2024-01-12 10:01:00",
		ModifiedTime: "2024-01-12 10:02:00",
		Source:       "手机",
		Type:         "消费",
		Target:       "商家A",
		Product:      "商品X",
		Amount:       99.5,
		FinType:      "支出",
		Status:       "交易成功",
		Charge:       0.5,
		Refund:       0.0,
		Comment:      "test",
		FundStatus:   "已支出",
	}

	if trans.GetID() != "2024011234567890" {
		t.Errorf("GetID() = %q", trans.GetID())
	}
	if trans.GetOrderID() != "ORD001" {
		t.Errorf("GetOrderID() = %q", trans.GetOrderID())
	}
	if trans.GetCreatedTime() != "2024-01-12 10:00:00" {
		t.Errorf("GetCreatedTime() = %q", trans.GetCreatedTime())
	}
	if trans.GetPaidTime() != "2024-01-12 10:01:00" {
		t.Errorf("GetPaidTime() = %q", trans.GetPaidTime())
	}
	if trans.GetModifiedTime() != "2024-01-12 10:02:00" {
		t.Errorf("GetModifiedTime() = %q", trans.GetModifiedTime())
	}
	if trans.GetSource() != "手机" {
		t.Errorf("GetSource() = %q", trans.GetSource())
	}
	if trans.GetType() != "消费" {
		t.Errorf("GetType() = %q", trans.GetType())
	}
	if trans.GetTarget() != "商家A" {
		t.Errorf("GetTarget() = %q", trans.GetTarget())
	}
	if trans.GetProduct() != "商品X" {
		t.Errorf("GetProduct() = %q", trans.GetProduct())
	}
	if trans.GetAmount() != 99.5 {
		t.Errorf("GetAmount() = %v", trans.GetAmount())
	}
	if trans.GetFormatAmount() != 99.5 {
		t.Errorf("GetFormatAmount() = %v", trans.GetFormatAmount())
	}
	if trans.GetFinType() != "支出" {
		t.Errorf("GetFinType() = %q", trans.GetFinType())
	}
	if trans.GetStatus() != "交易成功" {
		t.Errorf("GetStatus() = %q", trans.GetStatus())
	}
	if trans.GetCharge() != 0.5 {
		t.Errorf("GetCharge() = %v", trans.GetCharge())
	}
	if trans.GetRefund() != 0.0 {
		t.Errorf("GetRefund() = %v", trans.GetRefund())
	}
	if trans.GetComment() != "test" {
		t.Errorf("GetComment() = %q", trans.GetComment())
	}
	if trans.GetFundStatus() != "已支出" {
		t.Errorf("GetFundStatus() = %q", trans.GetFundStatus())
	}
}

func TestAlipayTransYearMonth(t *testing.T) {
	// ID starts with "20"
	trans1 := &AlipayTrans{ID: "202401123456"}
	if ym := trans1.YearMonth(); ym != "202401" {
		t.Errorf("YearMonth() = %q, want 202401", ym)
	}

	// ID does not start with "20"
	trans2 := &AlipayTrans{ID: "2401123456789"}
	if ym := trans2.YearMonth(); ym != "202401" {
		t.Errorf("YearMonth() = %q, want 202401", ym)
	}
}

func TestAlipayTransIsClosed(t *testing.T) {
	trans := &AlipayTrans{Status: "交易关闭"}
	if !trans.IsClosed() {
		t.Error("expected IsClosed() = true for 交易关闭")
	}
	trans2 := &AlipayTrans{Status: "失败"}
	if !trans2.IsClosed() {
		t.Error("expected IsClosed() = true for 失败")
	}
	trans3 := &AlipayTrans{Status: "交易成功"}
	if trans3.IsClosed() {
		t.Error("expected IsClosed() = false for 交易成功")
	}
}

func TestAlipayTransIsIncome(t *testing.T) {
	trans := &AlipayTrans{FinType: "收入"}
	if !trans.IsIncome() {
		t.Error("expected IsIncome() = true for 收入")
	}
	trans2 := &AlipayTrans{FinType: "支出", Product: "退款"}
	if !trans2.IsIncome() {
		t.Error("expected IsIncome() = true for 退款 product")
	}
	trans3 := &AlipayTrans{FinType: "支出", Product: "购物"}
	if trans3.IsIncome() {
		t.Error("expected IsIncome() = false")
	}
}

func TestAlipayTransIsTransfer(t *testing.T) {
	trans := &AlipayTrans{FundStatus: "资金转移"}
	if !trans.IsTransfer() {
		t.Error("expected IsTransfer() = true for 资金转移")
	}
	trans2 := &AlipayTrans{Product: "转账"}
	if !trans2.IsTransfer() {
		t.Error("expected IsTransfer() = true for 转账 product")
	}
}

func TestAlipayTransIsInnerTransfer(t *testing.T) {
	trans := &AlipayTrans{Product: "余额宝-自动转入"}
	if !trans.IsInnerTransfer() {
		t.Error("expected IsInnerTransfer() = true")
	}
	trans2 := &AlipayTrans{Product: "购物"}
	if trans2.IsInnerTransfer() {
		t.Error("expected IsInnerTransfer() = false")
	}
}

func TestAlipayTransIsShowInList(t *testing.T) {
	trans := &AlipayTrans{Amount: 100.0}
	if !trans.IsShowInList() {
		t.Error("expected IsShowInList() = true for Amount 100")
	}
	trans2 := &AlipayTrans{Amount: 1.0}
	if trans2.IsShowInList() {
		t.Error("expected IsShowInList() = false for Amount 1")
	}
}

func TestNewAlipayTrans(t *testing.T) {
	trans := NewAlipayTrans()
	if _, ok := trans.(*AlipayTrans); !ok {
		t.Error("NewAlipayTrans should return *AlipayTrans")
	}
}

func TestAlipayTransParserMethods(t *testing.T) {
	p := TransParserAlipay
	if p.CsvHeader() != AlipayCsvHeader {
		t.Errorf("CsvHeader() mismatch")
	}
	if p.FieldNum() != AlipayCsvFieldNum {
		t.Errorf("FieldNum() = %d, want %d", p.FieldNum(), AlipayCsvFieldNum)
	}
	if p.Enc() != simplifiedchinese.GBK {
		t.Error("Enc() should be GBK")
	}
	trans := p.NewTrans()
	if _, ok := trans.(*AlipayTrans); !ok {
		t.Error("NewTrans should return *AlipayTrans")
	}
}
