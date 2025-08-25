// Copyright 2019 vogo. All rights reserved.

package aliwepaystat

type MonthStat struct {
	TransMap             map[string]Trans `json:"trans_map" csv:"trans_map" comment:"交易记录"`
	YearMonth            string           `json:"year_month" csv:"year_month" comment:"年月"`
	Investment           *TransGroup      `json:"investment" csv:"investment" comment:"投资"`
	InnerTransfer        *TransGroup      `json:"inner_transfer" csv:"inner_transfer" comment:"内部转账"`
	Income               *TransGroup      `json:"income" csv:"income" comment:"收入"`
	IncomeTransfer       *TransGroup      `json:"income_transfer" csv:"income_transfer" comment:"收入转账"`
	Loan                 *TransGroup      `json:"loan" csv:"loan" comment:"贷款"`
	LoanRepayment        *TransGroup      `json:"loan_repayment" csv:"loan_repayment" comment:"贷款还款"`
	CreditRepayment      *TransGroup      `json:"credit_repayment" csv:"credit_repayment" comment:"信用还款"`
	ExpenseTotal         float64          `json:"expense_total" csv:"expense_total" comment:"支出"`
	ExpenseTransfer      *TransGroup      `json:"expense_transfer" csv:"expense_transfer" comment:"支出转账"`
	ExpenseTravel        *TransGroup      `json:"expense_travel" csv:"expense_travel" comment:"支出旅行"`
	ExpenseEat           *TransGroup      `json:"expense_eat" csv:"expense_eat" comment:"支出就餐"`
	ExpenseWaterElectGas *TransGroup      `json:"expense_water_elect_gas" csv:"expense_water_elect_gas" comment:"支出水电 gas"`
	ExpenseTel           *TransGroup      `json:"expense_tel" csv:"expense_tel" comment:"支出电话"`
	ExpenseOther         *TransGroup      `json:"expense_other" csv:"expense_other" comment:"支出其他"`
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

		// 信用还款也作为支出的一部分
		ms.ExpenseTotal += trans.GetAmount()

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
