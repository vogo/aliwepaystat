package aliwepaystat

import (
	"database/sql"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	EnsureSchema(db)
	return db
}

func TestOpenDBAndEnsureSchema(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Verify tables exist
	tables := []string{"transactions", "config", "month_stats", "category_stats"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestInsertTransAndLoadExistingIDs(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	trans := &AlipayTrans{
		ID:          "2024011234567890",
		OrderID:     "ORD001",
		CreatedTime: "2024-01-12 10:00:00",
		Source:      "手机",
		Type:        "消费",
		Target:      "商家",
		Product:     "商品",
		Amount:      50.0,
		FinType:     "支出",
		Status:      "交易成功",
		Refund:      0,
		Comment:     "",
		FundStatus:  "已支出",
	}

	err := InsertTrans(db, trans, "alipay", "202401")
	if err != nil {
		t.Fatalf("InsertTrans error: %v", err)
	}

	ids := LoadExistingIDs(db)
	if _, ok := ids["2024011234567890"]; !ok {
		t.Error("expected ID to be loaded")
	}
	if len(ids) != 1 {
		t.Errorf("expected 1 ID, got %d", len(ids))
	}

	// Insert same ID again (should be ignored)
	err = InsertTrans(db, trans, "alipay", "202401")
	if err != nil {
		t.Fatalf("InsertTrans duplicate error: %v", err)
	}
	ids2 := LoadExistingIDs(db)
	if len(ids2) != 1 {
		t.Errorf("expected 1 ID after duplicate insert, got %d", len(ids2))
	}
}

func TestGetFundStatusIfAny(t *testing.T) {
	alipay := &AlipayTrans{FundStatus: "已支出"}
	if got := getFundStatusIfAny(alipay); got != "已支出" {
		t.Errorf("getFundStatusIfAny(alipay) = %q, want 已支出", got)
	}
	wechat := &WechatTrans{}
	if got := getFundStatusIfAny(wechat); got != "" {
		t.Errorf("getFundStatusIfAny(wechat) = %q, want empty", got)
	}
}

func TestQueryYearMonths(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Insert transactions for different months
	for _, ym := range []string{"202401", "202402", "202401"} {
		trans := &AlipayTrans{
			ID:      ym + "0001" + ym,
			Amount:  10.0,
			FinType: "支出",
			Status:  "成功",
		}
		if err := InsertTrans(db, trans, "alipay", ym); err != nil {
			t.Fatal(err)
		}
	}

	yms, err := QueryYearMonths(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(yms) != 2 {
		t.Errorf("expected 2 year months, got %d", len(yms))
	}
	if yms[0] != "202401" || yms[1] != "202402" {
		t.Errorf("unexpected year months: %v", yms)
	}
}

func TestQueryTransByYearMonth(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	trans := &AlipayTrans{
		ID:          "2024010001",
		OrderID:     "ORD1",
		CreatedTime: "2024-01-01 10:00:00",
		Source:      "手机",
		Type:        "消费",
		Target:      "商家",
		Product:     "商品",
		Amount:      100.0,
		FinType:     "支出",
		Status:      "成功",
		Comment:     "test",
		FundStatus:  "已支出",
	}
	if err := InsertTrans(db, trans, "alipay", "202401"); err != nil {
		t.Fatal(err)
	}

	list, err := QueryTransByYearMonth(db, "202401")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(list))
	}

	dbTrans := list[0].(*DbTrans)
	if dbTrans.GetID() != "2024010001" {
		t.Errorf("GetID() = %q", dbTrans.GetID())
	}
	if dbTrans.Platform != "alipay" {
		t.Errorf("Platform = %q", dbTrans.Platform)
	}
	if dbTrans.GetAmount() != 100.0 {
		t.Errorf("GetAmount() = %v", dbTrans.GetAmount())
	}

	// Query empty month
	list2, err := QueryTransByYearMonth(db, "202412")
	if err != nil {
		t.Fatal(err)
	}
	if len(list2) != 0 {
		t.Errorf("expected 0 for empty month, got %d", len(list2))
	}
}

func TestDbTransMethods(t *testing.T) {
	trans := &DbTrans{
		ID:           "DB001",
		OrderID:      "DBORD001",
		Platform:     "alipay",
		YearMonthStr: "202401",
		CreatedTime:  "2024-01-15 10:00:00",
		Source:       "手机",
		Type:         "消费",
		Target:       "商家",
		Product:      "商品",
		Amount:       50.0,
		FinType:      "支出",
		Status:       "成功",
		Refund:       0,
		Comment:      "test",
		FundStatus:   "",
	}

	if trans.GetID() != "DB001" {
		t.Errorf("GetID() = %q", trans.GetID())
	}
	if trans.GetOrderID() != "DBORD001" {
		t.Errorf("GetOrderID() = %q", trans.GetOrderID())
	}
	if trans.GetCreatedTime() != "2024-01-15 10:00:00" {
		t.Errorf("GetCreatedTime() = %q", trans.GetCreatedTime())
	}
	if trans.GetSource() != "手机" {
		t.Errorf("GetSource() = %q", trans.GetSource())
	}
	if trans.GetType() != "消费" {
		t.Errorf("GetType() = %q", trans.GetType())
	}
	if trans.GetTarget() != "商家" {
		t.Errorf("GetTarget() = %q", trans.GetTarget())
	}
	if trans.GetProduct() != "商品" {
		t.Errorf("GetProduct() = %q", trans.GetProduct())
	}
	if trans.GetAmount() != 50.0 {
		t.Errorf("GetAmount() = %v", trans.GetAmount())
	}
	if trans.GetFormatAmount() != 50.0 {
		t.Errorf("GetFormatAmount() = %v", trans.GetFormatAmount())
	}
	if trans.GetFinType() != "支出" {
		t.Errorf("GetFinType() = %q", trans.GetFinType())
	}
	if trans.GetStatus() != "成功" {
		t.Errorf("GetStatus() = %q", trans.GetStatus())
	}
	if trans.GetRefund() != 0 {
		t.Errorf("GetRefund() = %v", trans.GetRefund())
	}
	if trans.GetComment() != "test" {
		t.Errorf("GetComment() = %q", trans.GetComment())
	}
	if trans.YearMonth() != "202401" {
		t.Errorf("YearMonth() = %q", trans.YearMonth())
	}
}

func TestDbTransIsIncome(t *testing.T) {
	trans := &DbTrans{FinType: "收入", Product: "x"}
	if !trans.IsIncome() {
		t.Error("expected true for 收入")
	}
	trans2 := &DbTrans{FinType: "支出", Product: "退款"}
	if !trans2.IsIncome() {
		t.Error("expected true for 退款 product")
	}
}

func TestDbTransIsTransfer(t *testing.T) {
	trans := &DbTrans{Type: "转账"}
	if !trans.IsTransfer() {
		t.Error("expected true for 转账 type")
	}
	trans2 := &DbTrans{FundStatus: "资金转移"}
	if !trans2.IsTransfer() {
		t.Error("expected true for 资金转移")
	}
}

func TestDbTransIsInnerTransfer(t *testing.T) {
	trans := &DbTrans{Product: "余额宝-自动转入"}
	if !trans.IsInnerTransfer() {
		t.Error("expected true")
	}
}

func TestDbTransIsClosed(t *testing.T) {
	trans := &DbTrans{Status: "交易关闭"}
	if !trans.IsClosed() {
		t.Error("expected true")
	}
}

func TestDbTransIsShowInList(t *testing.T) {
	trans := &DbTrans{Amount: 100.0}
	if !trans.IsShowInList() {
		t.Error("expected true")
	}
	trans2 := &DbTrans{Amount: 1.0}
	if trans2.IsShowInList() {
		t.Error("expected false")
	}
}

func TestConfigCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Set config
	err := SetConfig(db, "test.key", "test.value", "test description")
	if err != nil {
		t.Fatal(err)
	}

	// Get config
	val, err := GetConfig(db, "test.key")
	if err != nil {
		t.Fatal(err)
	}
	if val != "test.value" {
		t.Errorf("GetConfig() = %q, want test.value", val)
	}

	// Update config
	err = SetConfig(db, "test.key", "new.value", "updated")
	if err != nil {
		t.Fatal(err)
	}
	val2, err := GetConfig(db, "test.key")
	if err != nil {
		t.Fatal(err)
	}
	if val2 != "new.value" {
		t.Errorf("GetConfig() after update = %q, want new.value", val2)
	}

	// Get all config (includes defaults from initDefaultConfig)
	all, err := GetAllConfig(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := all["test.key"]; !ok {
		t.Error("test.key not in GetAllConfig")
	}
	if len(all) < 1 {
		t.Error("expected at least 1 config item")
	}
}

func TestSplitKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{"a，b，c", []string{"a", "b", "c"}},
		{"a,b，c", []string{"a", "b", "c"}},
		{"single", []string{"single"}},
		{"", nil},
		{",a,,b,", []string{"a", "b"}},
	}
	for _, tt := range tests {
		got := splitKeywords(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("splitKeywords(%q) = %v, want %v", tt.input, got, tt.expected)
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("splitKeywords(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
			}
		}
	}
}

func TestContainsString(t *testing.T) {
	if !containsString("hello world", "world") {
		t.Error("expected true")
	}
	if containsString("hello", "world") {
		t.Error("expected false")
	}
	if !containsString("abc", "") {
		t.Error("empty substr should match")
	}
	if containsString("ab", "abc") {
		t.Error("shorter string should not match longer substr")
	}
}

func TestContainsKeywords(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Set test keywords
	if err := SetConfig(db, "test.keywords", "美团,外卖", "test"); err != nil {
		t.Fatal(err)
	}

	if !containsKeywords(db, "test.keywords", "美团外卖订单", "") {
		t.Error("expected true for 美团")
	}
	if containsKeywords(db, "test.keywords", "其他商品", "其他备注") {
		t.Error("expected false for unmatched")
	}
	if containsKeywords(db, "nonexistent.key", "anything", "") {
		t.Error("expected false for nonexistent key")
	}
}

func TestSaveAndGetMonthStats(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	stats := &MonthStatData{
		YearMonth:    "202401",
		TotalIncome:  1000.0,
		TotalExpense: 500.0,
		LoanTotal:    100.0,
	}

	err := saveMonthStats(db, stats)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetMonthStats(db, "202401")
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalIncome != 1000.0 {
		t.Errorf("TotalIncome = %v, want 1000", got.TotalIncome)
	}
	if got.TotalExpense != 500.0 {
		t.Errorf("TotalExpense = %v, want 500", got.TotalExpense)
	}

	// Get non-existent month
	_, err = GetMonthStats(db, "202412")
	if err == nil {
		t.Error("expected error for non-existent month")
	}
}

func TestGetAllMonthStats(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	for _, ym := range []string{"202401", "202402"} {
		if err := saveMonthStats(db, &MonthStatData{YearMonth: ym, TotalIncome: 100}); err != nil {
			t.Fatal(err)
		}
	}

	all, err := GetAllMonthStats(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 stats, got %d", len(all))
	}
	// Should be ordered DESC
	if all[0].YearMonth != "202402" {
		t.Errorf("expected 202402 first, got %s", all[0].YearMonth)
	}
}

func TestUpdateMonthStats(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Insert some transactions
	transactions := []struct {
		id      string
		amount  float64
		finType string
		product string
	}{
		{"T001", 5000.0, "收入", "工资"},
		{"T002", -30.0, "支出", "美团外卖"},
		{"T003", -15.0, "支出", "滴滴出行"},
		{"T004", -100.0, "支出", "购物"},
	}
	for _, tx := range transactions {
		trans := &AlipayTrans{
			ID:          tx.id,
			Amount:      tx.amount,
			FinType:     tx.finType,
			Status:      "成功",
			Product:     tx.product,
			CreatedTime: "2024-01-15 10:00:00",
		}
		if err := InsertTrans(db, trans, "alipay", "202401"); err != nil {
			t.Fatal(err)
		}
	}

	err := UpdateMonthStats(db, "202401")
	if err != nil {
		t.Fatal(err)
	}

	stats, err := GetMonthStats(db, "202401")
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalIncome != 5000.0 {
		t.Errorf("TotalIncome = %v, want 5000", stats.TotalIncome)
	}
	if stats.TotalExpense != 145.0 {
		t.Errorf("TotalExpense = %v, want 145", stats.TotalExpense)
	}
}

func TestSetConfigIfNotExists(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Set a value
	setConfigIfNotExists(db, "my.key", "value1", "desc1")
	val, _ := GetConfig(db, "my.key")
	if val != "value1" {
		t.Errorf("expected value1, got %q", val)
	}

	// Should not overwrite
	setConfigIfNotExists(db, "my.key", "value2", "desc2")
	val2, _ := GetConfig(db, "my.key")
	if val2 != "value1" {
		t.Errorf("expected value1 (not overwritten), got %q", val2)
	}
}
