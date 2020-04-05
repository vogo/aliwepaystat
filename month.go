// Copyright 2019 wongoo. All rights reserved.

package aliwepaystat

type MonthStat struct {
	TransMap             map[string]Trans
	YearMonth            string
	Investment           *TransGroup
	InnerTransfer        *TransGroup
	Income               *TransGroup
	IncomeTransfer       *TransGroup
	Loan                 *TransGroup
	LoanRepayment        *TransGroup
	CreditRepayment      *TransGroup
	ExpenseTotal         float64
	ExpenseTransfer      *TransGroup
	ExpenseTravel        *TransGroup
	ExpenseEat           *TransGroup
	ExpenseWaterElectGas *TransGroup
	ExpenseTel           *TransGroup
	ExpenseOther         *TransGroup
}

func (ms *MonthStat) FormatExpenseTotal() float64 {
	return RoundFloat(ms.ExpenseTotal)
}

func (ms *MonthStat) add(trans Trans) {
	//ignore closed
	if trans.IsClosed() {
		return
	}

	// trans already exists
	if _, ok := ms.TransMap[trans.GetID()]; ok {
		return
	}


	ms.TransMap[trans.GetID()] = trans

	// [1] 贷款放在收入之前判断
	if EitherContainsAny(trans.GetProduct(), trans.GetTarget(), cfg.LoanKeyWords...) {
		ms.Loan.add(trans)
		return
	}

	// [2] 贷款还款
	if EitherContainsAny(trans.GetProduct(), trans.GetTarget(), cfg.LoanRepaymentKeyWords...) {
		ms.LoanRepayment.add(trans)
		return
	}

	if IsInvestment(trans.GetProduct()) {
		ms.Investment.add(trans)
		return
	}

	// [2] 信用还款
	if EitherContainsAny(trans.GetProduct(), trans.GetTarget(), cfg.RepaymentKeyWords...) {
		ms.CreditRepayment.add(trans)
		return
	}

	// [3] 内部转账
	if trans.IsInnerTransfer() {
		ms.InnerTransfer.add(trans)
		return
	}

	// [4] 收入判断 (包括转入)
	if trans.IsIncome() {

		// 转账收入单独统计，不计入普通收入
		if ContainsAny(trans.GetProduct(), cfg.TransferKeyWords...) {
			ms.IncomeTransfer.add(trans)
			return
		}

		ms.Income.add(trans)
		return
	}

	// [5] 转账 (转账支出单独统计，不计入普通支出)
	if trans.IsTransfer() {
		ms.ExpenseTransfer.add(trans)
		return
	}

	// [6] 开始统计支出
	ms.ExpenseTotal += trans.GetAmount()

	switch {
	case EitherContainsAny(trans.GetProduct(), trans.GetTarget(), cfg.TravelKeyWords...):
		ms.ExpenseTravel.add(trans)
	case EitherContainsAny(trans.GetProduct(), trans.GetTarget(), cfg.EatKeyWords...) || IsWechatGroupAAExpense(trans):
		ms.ExpenseEat.add(trans)
	case EitherContainsAny(trans.GetProduct(), trans.GetTarget(), cfg.WaterElectGasKeyWords...):
		ms.ExpenseWaterElectGas.add(trans)
	case EitherContainsAny(trans.GetProduct(), trans.GetTarget(), cfg.TelKeyWords...):
		ms.ExpenseTel.add(trans)
	default:
		ms.ExpenseOther.add(trans)
	}
}

var (
	monthStatsMap = make(map[string]*MonthStat)
	yearMonths    []string
)

func getMonthStat(yearMonth string) *MonthStat {
	ms, ok := monthStatsMap[yearMonth]
	if !ok {
		ms = &MonthStat{
			TransMap:             make(map[string]Trans),
			YearMonth:            yearMonth,
			Investment:           &TransGroup{},
			Loan:                 &TransGroup{},
			ExpenseTransfer:      &TransGroup{},
			InnerTransfer:        &TransGroup{},
			IncomeTransfer:       &TransGroup{},
			Income:               &TransGroup{},
			LoanRepayment:        &TransGroup{},
			CreditRepayment:      &TransGroup{},
			ExpenseTravel:        &TransGroup{},
			ExpenseEat:           &TransGroup{},
			ExpenseWaterElectGas: &TransGroup{},
			ExpenseTel:           &TransGroup{},
			ExpenseOther:         &TransGroup{},
		}
		monthStatsMap[yearMonth] = ms
		yearMonths = append(yearMonths, yearMonth)
	}
	return ms
}
