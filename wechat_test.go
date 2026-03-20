package aliwepaystat

import (
	"testing"

	"golang.org/x/text/encoding/unicode"
)

func TestWechatTransGetters(t *testing.T) {
	trans := &WechatTrans{
		CreatedTime: "2024-01-12 10:00:00",
		Type:        "消费",
		Target:      "商家B",
		Product:     "商品Y",
		FinType:     "支出",
		Amount:      "¥88.00",
		Source:      "零钱",
		Status:      "支付成功",
		ID:          "WX001",
		OrderID:     "WXORD001",
		Comment:     "备注",
		Refund:      0.0,
	}

	if trans.GetID() != "WX001" {
		t.Errorf("GetID() = %q", trans.GetID())
	}
	if trans.GetOrderID() != "WXORD001" {
		t.Errorf("GetOrderID() = %q", trans.GetOrderID())
	}
	if trans.GetCreatedTime() != "2024-01-12 10:00:00" {
		t.Errorf("GetCreatedTime() = %q", trans.GetCreatedTime())
	}
	if trans.GetSource() != "零钱" {
		t.Errorf("GetSource() = %q", trans.GetSource())
	}
	if trans.GetType() != "消费" {
		t.Errorf("GetType() = %q", trans.GetType())
	}
	if trans.GetTarget() != "商家B" {
		t.Errorf("GetTarget() = %q", trans.GetTarget())
	}
	if trans.GetProduct() != "商品Y" {
		t.Errorf("GetProduct() = %q", trans.GetProduct())
	}
	if trans.GetFinType() != "支出" {
		t.Errorf("GetFinType() = %q", trans.GetFinType())
	}
	if trans.GetStatus() != "支付成功" {
		t.Errorf("GetStatus() = %q", trans.GetStatus())
	}
	if trans.GetRefund() != 0.0 {
		t.Errorf("GetRefund() = %v", trans.GetRefund())
	}
	if trans.GetComment() != "备注" {
		t.Errorf("GetComment() = %q", trans.GetComment())
	}
}

func TestWechatTransGetAmount(t *testing.T) {
	trans := &WechatTrans{Amount: "¥88.50"}
	if amt := trans.GetAmount(); amt != 88.50 {
		t.Errorf("GetAmount() = %v, want 88.50", amt)
	}
	// Second call uses cached Amt
	if amt := trans.GetAmount(); amt != 88.50 {
		t.Errorf("GetAmount() cached = %v, want 88.50", amt)
	}
}

func TestWechatTransGetAmountNoYen(t *testing.T) {
	trans := &WechatTrans{Amount: "100.00"}
	if amt := trans.GetAmount(); amt != 100.0 {
		t.Errorf("GetAmount() = %v, want 100.0", amt)
	}
}

func TestWechatTransGetAmountPreset(t *testing.T) {
	trans := &WechatTrans{Amt: 50.0, Amount: "¥999.00"}
	// When Amt is already set, Amount string is not parsed
	if amt := trans.GetAmount(); amt != 50.0 {
		t.Errorf("GetAmount() = %v, want 50.0", amt)
	}
}

func TestWechatTransGetFormatAmount(t *testing.T) {
	trans := &WechatTrans{Amt: 99.12345}
	if fa := trans.GetFormatAmount(); fa != 99.1235 {
		t.Errorf("GetFormatAmount() = %v, want 99.1235", fa)
	}
}

func TestWechatTransYearMonth(t *testing.T) {
	trans := &WechatTrans{CreatedTime: "2024-01-15 10:00:00"}
	if ym := trans.YearMonth(); ym != "202401" {
		t.Errorf("YearMonth() = %q, want 202401", ym)
	}
}

func TestWechatTransIsClosed(t *testing.T) {
	trans := &WechatTrans{Status: "交易关闭"}
	if !trans.IsClosed() {
		t.Error("expected IsClosed() = true")
	}
	trans2 := &WechatTrans{Status: "支付成功"}
	if trans2.IsClosed() {
		t.Error("expected IsClosed() = false")
	}
}

func TestWechatTransIsIncome(t *testing.T) {
	trans := &WechatTrans{FinType: "收入"}
	if !trans.IsIncome() {
		t.Error("expected IsIncome() = true")
	}
	trans2 := &WechatTrans{FinType: "支出", Product: "退款"}
	if !trans2.IsIncome() {
		t.Error("expected IsIncome() = true for 退款")
	}
}

func TestWechatTransIsTransfer(t *testing.T) {
	trans := &WechatTrans{Type: "转账"}
	if !trans.IsTransfer() {
		t.Error("expected IsTransfer() = true")
	}
}

func TestWechatTransIsInnerTransfer(t *testing.T) {
	trans := &WechatTrans{Product: "余额宝-自动转入"}
	if !trans.IsInnerTransfer() {
		t.Error("expected IsInnerTransfer() = true")
	}
}

func TestWechatTransIsShowInList(t *testing.T) {
	trans := &WechatTrans{Amt: 100.0}
	if !trans.IsShowInList() {
		t.Error("expected IsShowInList() = true for 100")
	}
	trans2 := &WechatTrans{Amt: 1.0}
	if trans2.IsShowInList() {
		t.Error("expected IsShowInList() = false for 1")
	}
}

func TestWechatTransParserMethods(t *testing.T) {
	p := TransParserWechat
	if p.CsvHeader() != WechatCsvHeader {
		t.Errorf("CsvHeader() mismatch")
	}
	if p.FieldNum() != WechatCsvFieldNum {
		t.Errorf("FieldNum() = %d, want %d", p.FieldNum(), WechatCsvFieldNum)
	}
	if p.Enc() != unicode.UTF8 {
		t.Error("Enc() should be UTF8")
	}
	trans := p.NewTrans()
	if _, ok := trans.(*WechatTrans); !ok {
		t.Error("NewTrans should return *WechatTrans")
	}
}

func TestIsWechatGroupAAExpense(t *testing.T) {
	trans := &WechatTrans{Type: "群收款", FinType: "支出"}
	if !IsWechatGroupAAExpense(trans) {
		t.Error("expected true for 群收款 + 支出")
	}
	trans2 := &WechatTrans{Type: "群收款", FinType: "收入"}
	if IsWechatGroupAAExpense(trans2) {
		t.Error("expected false for 群收款 + 收入")
	}
	trans3 := &WechatTrans{Type: "消费", FinType: "支出"}
	if IsWechatGroupAAExpense(trans3) {
		t.Error("expected false for 消费")
	}
}
