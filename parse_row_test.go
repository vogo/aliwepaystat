package aliwepaystat

import (
	"testing"
)

func TestAlipayParseRow(t *testing.T) {
	fields := []string{
		"2024010100001",       // 0: ID
		"ORDER001",            // 1: OrderID
		"2024-01-01 10:00:00", // 2: CreatedTime
		"2024-01-01 10:00:01", // 3: PaidTime
		"2024-01-01 10:00:02", // 4: ModifiedTime
		"手机",                  // 5: Source
		"即时到账交易",              // 6: Type
		"某商户",                 // 7: Target
		"测试商品",                // 8: Product
		"99.50",               // 9: Amount
		"支出",                  // 10: FinType
		"交易成功",                // 11: Status
		"0.00",                // 12: Charge
		"0.00",                // 13: Refund
		"无",                   // 14: Comment
		"已支出",                 // 15: FundStatus
		"",                    // 16: Other
	}

	trans, err := TransParserAlipay.ParseRow(fields)
	if err != nil {
		t.Fatalf("ParseRow failed: %v", err)
	}

	at, ok := trans.(*AlipayTrans)
	if !ok {
		t.Fatalf("expected *AlipayTrans, got %T", trans)
	}

	if at.ID != "2024010100001" {
		t.Errorf("ID = %q, want %q", at.ID, "2024010100001")
	}
	if at.OrderID != "ORDER001" {
		t.Errorf("OrderID = %q, want %q", at.OrderID, "ORDER001")
	}
	if at.CreatedTime != "2024-01-01 10:00:00" {
		t.Errorf("CreatedTime = %q, want %q", at.CreatedTime, "2024-01-01 10:00:00")
	}
	if at.PaidTime != "2024-01-01 10:00:01" {
		t.Errorf("PaidTime = %q, want %q", at.PaidTime, "2024-01-01 10:00:01")
	}
	if at.ModifiedTime != "2024-01-01 10:00:02" {
		t.Errorf("ModifiedTime = %q, want %q", at.ModifiedTime, "2024-01-01 10:00:02")
	}
	if at.Source != "手机" {
		t.Errorf("Source = %q, want %q", at.Source, "手机")
	}
	if at.Type != "即时到账交易" {
		t.Errorf("Type = %q, want %q", at.Type, "即时到账交易")
	}
	if at.Target != "某商户" {
		t.Errorf("Target = %q, want %q", at.Target, "某商户")
	}
	if at.Product != "测试商品" {
		t.Errorf("Product = %q, want %q", at.Product, "测试商品")
	}
	if at.Amount != 99.50 {
		t.Errorf("Amount = %f, want %f", at.Amount, 99.50)
	}
	if at.FinType != "支出" {
		t.Errorf("FinType = %q, want %q", at.FinType, "支出")
	}
	if at.Status != "交易成功" {
		t.Errorf("Status = %q, want %q", at.Status, "交易成功")
	}
	if at.Charge != 0.0 {
		t.Errorf("Charge = %f, want %f", at.Charge, 0.0)
	}
	if at.Refund != 0.0 {
		t.Errorf("Refund = %f, want %f", at.Refund, 0.0)
	}
	if at.Comment != "无" {
		t.Errorf("Comment = %q, want %q", at.Comment, "无")
	}
	if at.FundStatus != "已支出" {
		t.Errorf("FundStatus = %q, want %q", at.FundStatus, "已支出")
	}
	if at.Other != "" {
		t.Errorf("Other = %q, want %q", at.Other, "")
	}
}

func TestAlipayParseRowInvalidAmount(t *testing.T) {
	fields := []string{
		"2024010100001", "ORDER001", "2024-01-01", "2024-01-01", "2024-01-01",
		"手机", "类型", "商户", "商品",
		"not_a_number", // invalid Amount
		"支出", "交易成功",
		"0.00", "0.00", "备注", "已支出", "",
	}

	_, err := TransParserAlipay.ParseRow(fields)
	if err == nil {
		t.Fatal("expected error for invalid Amount, got nil")
	}
}

func TestAlipayParseRowInvalidCharge(t *testing.T) {
	fields := []string{
		"2024010100001", "ORDER001", "2024-01-01", "2024-01-01", "2024-01-01",
		"手机", "类型", "商户", "商品",
		"10.00", "支出", "交易成功",
		"bad", // invalid Charge
		"0.00", "备注", "已支出", "",
	}

	_, err := TransParserAlipay.ParseRow(fields)
	if err == nil {
		t.Fatal("expected error for invalid Charge, got nil")
	}
}

func TestAlipayParseRowInvalidRefund(t *testing.T) {
	fields := []string{
		"2024010100001", "ORDER001", "2024-01-01", "2024-01-01", "2024-01-01",
		"手机", "类型", "商户", "商品",
		"10.00", "支出", "交易成功",
		"0.00",
		"bad", // invalid Refund
		"备注", "已支出", "",
	}

	_, err := TransParserAlipay.ParseRow(fields)
	if err == nil {
		t.Fatal("expected error for invalid Refund, got nil")
	}
}

func TestWechatParseRow(t *testing.T) {
	fields := []string{
		"2024-01-01 10:00:00", // 0: CreatedTime
		"商户消费",                // 1: Type
		"某商户",                 // 2: Target
		"测试商品",                // 3: Product
		"支出",                  // 4: FinType
		"¥99.50",              // 5: Amount
		"零钱",                  // 6: Source
		"支付成功",                // 7: Status
		"4200001234567890",    // 8: ID
		"ORDER001",            // 9: OrderID
		"无",                   // 10: Comment
	}

	trans, err := TransParserWechat.ParseRow(fields)
	if err != nil {
		t.Fatalf("ParseRow failed: %v", err)
	}

	wt, ok := trans.(*WechatTrans)
	if !ok {
		t.Fatalf("expected *WechatTrans, got %T", trans)
	}

	if wt.CreatedTime != "2024-01-01 10:00:00" {
		t.Errorf("CreatedTime = %q, want %q", wt.CreatedTime, "2024-01-01 10:00:00")
	}
	if wt.Type != "商户消费" {
		t.Errorf("Type = %q, want %q", wt.Type, "商户消费")
	}
	if wt.Target != "某商户" {
		t.Errorf("Target = %q, want %q", wt.Target, "某商户")
	}
	if wt.Product != "测试商品" {
		t.Errorf("Product = %q, want %q", wt.Product, "测试商品")
	}
	if wt.FinType != "支出" {
		t.Errorf("FinType = %q, want %q", wt.FinType, "支出")
	}
	if wt.Amount != "¥99.50" {
		t.Errorf("Amount = %q, want %q", wt.Amount, "¥99.50")
	}
	if wt.Source != "零钱" {
		t.Errorf("Source = %q, want %q", wt.Source, "零钱")
	}
	if wt.Status != "支付成功" {
		t.Errorf("Status = %q, want %q", wt.Status, "支付成功")
	}
	if wt.ID != "4200001234567890" {
		t.Errorf("ID = %q, want %q", wt.ID, "4200001234567890")
	}
	if wt.OrderID != "ORDER001" {
		t.Errorf("OrderID = %q, want %q", wt.OrderID, "ORDER001")
	}
	if wt.Comment != "无" {
		t.Errorf("Comment = %q, want %q", wt.Comment, "无")
	}
}

func TestWechatParseRowGetAmount(t *testing.T) {
	fields := []string{
		"2024-01-01 10:00:00", "商户消费", "某商户", "测试商品",
		"支出", "¥99.50", "零钱", "支付成功",
		"4200001234567890", "ORDER001", "无",
	}

	trans, err := TransParserWechat.ParseRow(fields)
	if err != nil {
		t.Fatalf("ParseRow failed: %v", err)
	}

	amount := trans.GetAmount()
	if amount < 99.0 || amount > 100.0 {
		t.Errorf("GetAmount() = %f, want ~99.50", amount)
	}
}
