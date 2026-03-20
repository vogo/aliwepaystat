package aliwepaystat

import "testing"

func TestBuildStatsFromDB(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Insert transactions
	trans1 := &AlipayTrans{
		ID:          "BS001",
		Amount:      100.0,
		FinType:     "收入",
		Status:      "成功",
		Product:     "工资",
		CreatedTime: "2024-01-15 10:00:00",
	}
	trans2 := &AlipayTrans{
		ID:          "BS002",
		Amount:      -20.0,
		FinType:     "支出",
		Status:      "成功",
		Product:     "午餐",
		Target:      "商家",
		CreatedTime: "2024-02-15 10:00:00",
	}
	if err := InsertTrans(db, trans1, "alipay", "202401"); err != nil {
		t.Fatal(err)
	}
	if err := InsertTrans(db, trans2, "alipay", "202402"); err != nil {
		t.Fatal(err)
	}

	BuildStatsFromDB(db)

	if len(monthStatsMap) != 2 {
		t.Errorf("expected 2 months in stats, got %d", len(monthStatsMap))
	}
	if len(yearMonths) != 2 {
		t.Errorf("expected 2 yearMonths, got %d", len(yearMonths))
	}

	ms1, ok := monthStatsMap["202401"]
	if !ok {
		t.Fatal("202401 not found in stats")
	}
	if ms1.Income.Total != 100.0 {
		t.Errorf("202401 Income.Total = %v, want 100", ms1.Income.Total)
	}
}
