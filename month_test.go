package aliwepaystat

import "testing"

func TestMonthStatFormatExpenseTotal(t *testing.T) {
	ms := &MonthStat{ExpenseTotal: 123.45678}
	if got := ms.FormatExpenseTotal(); got != 123.4568 {
		t.Errorf("FormatExpenseTotal() = %v, want 123.4568", got)
	}
}

func TestGetMonthStat(t *testing.T) {
	// Reset global state
	monthStatsMap = make(map[string]*MonthStat)
	yearMonths = nil

	ms1 := getMonthStat("202401")
	if ms1 == nil {
		t.Fatal("getMonthStat returned nil")
	}
	if ms1.YearMonth != "202401" {
		t.Errorf("YearMonth = %q, want 202401", ms1.YearMonth)
	}
	if len(yearMonths) != 1 {
		t.Errorf("yearMonths length = %d, want 1", len(yearMonths))
	}

	// Same month returns same instance
	ms2 := getMonthStat("202401")
	if ms1 != ms2 {
		t.Error("expected same instance for same month")
	}
	if len(yearMonths) != 1 {
		t.Errorf("yearMonths length = %d, want 1 (no duplicate)", len(yearMonths))
	}

	// Different month creates new
	ms3 := getMonthStat("202402")
	if ms1 == ms3 {
		t.Error("expected different instance for different month")
	}
	if len(yearMonths) != 2 {
		t.Errorf("yearMonths length = %d, want 2", len(yearMonths))
	}
}

func TestMonthStatAddClosed(t *testing.T) {
	monthStatsMap = make(map[string]*MonthStat)
	yearMonths = nil

	ms := getMonthStat("202401")

	// Closed transaction should be ignored
	closed := &AlipayTrans{
		ID:      "CLOSED001",
		Status:  "交易关闭",
		Amount:  100.0,
		FinType: "支出",
	}
	ms.add(closed)

	if len(ms.TransMap) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(ms.TransMap))
	}
}

func TestMonthStatAddDuplicate(t *testing.T) {
	monthStatsMap = make(map[string]*MonthStat)
	yearMonths = nil

	ms := getMonthStat("202401")

	trans := &AlipayTrans{
		ID:      "DUP001",
		Amount:  50.0,
		FinType: "支出",
		Status:  "成功",
		Product: "其他",
	}
	ms.add(trans)
	ms.add(trans) // duplicate

	if len(ms.TransMap) != 1 {
		t.Errorf("expected 1 transaction after duplicate, got %d", len(ms.TransMap))
	}
}

func TestMonthStatAddIncome(t *testing.T) {
	monthStatsMap = make(map[string]*MonthStat)
	yearMonths = nil

	ms := getMonthStat("202401")

	trans := &AlipayTrans{
		ID:      "INC001",
		Amount:  200.0,
		FinType: "收入",
		Status:  "成功",
		Product: "工资",
	}
	ms.add(trans)

	if ms.Income.Total != 200.0 {
		t.Errorf("Income.Total = %v, want 200", ms.Income.Total)
	}
}

func TestMonthStatAddExpense(t *testing.T) {
	monthStatsMap = make(map[string]*MonthStat)
	yearMonths = nil

	ms := getMonthStat("202401")

	trans := &AlipayTrans{
		ID:      "EXP001",
		Amount:  30.0,
		FinType: "支出",
		Status:  "成功",
		Product: "美团外卖",
		Target:  "美团",
	}
	ms.add(trans)

	if ms.ExpenseTotal != 30.0 {
		t.Errorf("ExpenseTotal = %v, want 30", ms.ExpenseTotal)
	}
	if ms.ExpenseEat.Total != 30.0 {
		t.Errorf("ExpenseEat.Total = %v, want 30", ms.ExpenseEat.Total)
	}
}

func TestMonthStatAddTransfer(t *testing.T) {
	monthStatsMap = make(map[string]*MonthStat)
	yearMonths = nil

	ms := getMonthStat("202401")

	trans := &AlipayTrans{
		ID:         "TRF001",
		Amount:     500.0,
		FinType:    "支出",
		Status:     "成功",
		Product:    "转账",
		FundStatus: "资金转移",
	}
	ms.add(trans)

	if ms.ExpenseTransfer.Total != 500.0 {
		t.Errorf("ExpenseTransfer.Total = %v, want 500", ms.ExpenseTransfer.Total)
	}
	// Transfer should not count as expense
	if ms.ExpenseTotal != 0 {
		t.Errorf("ExpenseTotal should be 0 for transfer, got %v", ms.ExpenseTotal)
	}
}

func TestMonthStatAddLoan(t *testing.T) {
	monthStatsMap = make(map[string]*MonthStat)
	yearMonths = nil

	ms := getMonthStat("202401")

	trans := &AlipayTrans{
		ID:      "LOAN001",
		Amount:  1000.0,
		FinType: "收入",
		Status:  "成功",
		Product: "放款",
	}
	ms.add(trans)

	if ms.Loan.Total != 1000.0 {
		t.Errorf("Loan.Total = %v, want 1000", ms.Loan.Total)
	}
}

func TestMonthStatAddInnerTransfer(t *testing.T) {
	monthStatsMap = make(map[string]*MonthStat)
	yearMonths = nil

	ms := getMonthStat("202401")

	trans := &AlipayTrans{
		ID:      "INT001",
		Amount:  300.0,
		FinType: "支出",
		Status:  "成功",
		Product: "余额宝-自动转入",
	}
	ms.add(trans)

	if ms.InnerTransfer.Total != 300.0 {
		t.Errorf("InnerTransfer.Total = %v, want 300", ms.InnerTransfer.Total)
	}
}

func TestMonthStatAddExpenseTravel(t *testing.T) {
	monthStatsMap = make(map[string]*MonthStat)
	yearMonths = nil

	ms := getMonthStat("202401")

	trans := &AlipayTrans{
		ID:      "TRV001",
		Amount:  15.0,
		FinType: "支出",
		Status:  "成功",
		Product: "滴滴出行",
	}
	ms.add(trans)

	if ms.ExpenseTravel.Total != 15.0 {
		t.Errorf("ExpenseTravel.Total = %v, want 15", ms.ExpenseTravel.Total)
	}
}

func TestMonthStatAddExpenseOther(t *testing.T) {
	monthStatsMap = make(map[string]*MonthStat)
	yearMonths = nil

	ms := getMonthStat("202401")

	trans := &AlipayTrans{
		ID:      "OTH001",
		Amount:  25.0,
		FinType: "支出",
		Status:  "成功",
		Product: "图书",
		Target:  "京东",
	}
	ms.add(trans)

	if ms.ExpenseOther.Total != 25.0 {
		t.Errorf("ExpenseOther.Total = %v, want 25", ms.ExpenseOther.Total)
	}
}
