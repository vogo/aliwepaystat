package main

type MonthStat struct {
	TransMap             map[string]*AlipayTrans
	YearMonth            string
	InnerTransfer        *TransGroup
	Income               *TransGroup
	IncomeTransfer       *TransGroup
	Loan                 *TransGroup
	LoanRepayment        *TransGroup
	CreditRepayment      *TransGroup
	ExpenseTotal         float32
	ExpenseTransfer      *TransGroup
	ExpenseTravel        *TransGroup
	ExpenseEat           *TransGroup
	ExpenseWaterElectGas *TransGroup
	ExpenseTel           *TransGroup
	ExpenseOther         *TransGroup
}

func (ms *MonthStat) add(trans *AlipayTrans) {
	//ignore closed
	if trans.isClosed() {
		return
	}

	// trans already exists
	if _, ok := ms.TransMap[trans.ID]; ok {
		return
	}
	ms.TransMap[trans.ID] = trans

	// [1] 贷款放在收入之前判断
	if BothContainsAny(trans.Product, trans.Target, loanKeyWords...) {
		ms.Loan.add(trans)
		return
	}

	// [2] 贷款还款
	if BothContainsAny(trans.Product, trans.Target, loanRepaymentKeyWords...) {
		ms.LoanRepayment.add(trans)
		return
	}

	// [2] 信用还款
	if BothContainsAny(trans.Product, trans.Target, repaymentKeyWords...) {
		ms.CreditRepayment.add(trans)
		return
	}

	// [3] 内部转账
	if trans.isInnerTransfer() {
		ms.InnerTransfer.add(trans)
		return
	}

	// [4] 收入判断 (包括转入)
	if trans.isIncome() {

		// 转账收入单独统计，不计入普通收入
		if ContainsAny(trans.Product, transferKeyWords...) {
			ms.IncomeTransfer.add(trans)
			return
		}

		ms.Income.add(trans)
		return
	}

	// [5] 转账 (转账支出单独统计，不计入普通支出)
	if Contains(trans.FundStatus, "资金转移") || BothContainsAny(trans.Product, trans.Target, transferKeyWords...) || ContainsAny(trans.Target, familyMembers...) {
		ms.ExpenseTransfer.add(trans)
		return
	}

	// [6] 开始统计支出
	ms.ExpenseTotal += trans.Amount

	if BothContainsAny(trans.Product, trans.Target, travelKeyWords...) {
		ms.ExpenseTravel.add(trans)
	} else if BothContainsAny(trans.Product, trans.Target, eatKeyWords...) {
		ms.ExpenseEat.add(trans)
	} else if BothContainsAny(trans.Product, trans.Target, waterElectGasKeyWords...) {
		ms.ExpenseWaterElectGas.add(trans)
	} else if BothContainsAny(trans.Product, trans.Target, telKeyWords...) {
		ms.ExpenseTel.add(trans)
	} else {
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
			TransMap:             make(map[string]*AlipayTrans),
			YearMonth:            yearMonth,
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
