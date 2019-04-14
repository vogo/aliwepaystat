package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jszwec/csvutil"
)

var (
	loanKeyWords          = []string{"放款"}
	transferKeyWords      = []string{"转账"}
	innerTransferKeyWords = []string{"余额宝-自动转入", "网商银行转入", "余额宝-转出到余额", "转出到网商银行"}
	incomeKeyWords        = []string{"收入", "红包奖励发放", "收益发放", "奖励", "退款"}
	repaymentKeyWords     = []string{"还款"}
	loanRepaymentKeyWords = []string{"蚂蚁借呗还款"}
	eatKeyWords           = []string{"美团", "饿了么", "口碑", "外卖", "菜", "餐饮", "美食", "饭", "超市", "汉堡", "安德鲁森", "节奏者", "拉面", "洪濑鸡爪", "肉夹馍", "麦之屋", "沙县小吃", "重庆小面"}
	travelKeyWords        = []string{"出行", "交通", "公交", "车", "打的", "的士", "taxi", "滴滴", "地铁"}
	waterElectGasKeyWords = []string{"水费", "电费", "燃气"}
	telKeyWords           = []string{"话费", "电信", "移动", "联通", "手机充值"}

	familyMembers = []string{"kiera", "望哥", "杨柳", "守望人之初", "龙补琴"}
)

func Contains(s string, f string) bool {
	return strings.Contains(s, f)
}

func ContainsAny(s string, f ...string) bool {
	for _, a := range f {
		if strings.Contains(s, a) {
			return true
		}
	}
	return false
}

func BothContainsAny(s1, s2 string, f ...string) bool {
	for _, a := range f {
		if strings.Contains(s1, a) || strings.Contains(s2, a) {
			return true
		}
	}
	return false
}

//AlipayTrans alipay transaction
type AlipayTrans struct {
	ID           string  `csv:"id"`
	OrderID      string  `csv:"order_id"`
	CreatedTime  string  `csv:"created_time"`
	PaidTime     string  `csv:"paid_time"`
	ModifiedTime string  `csv:"modified_time"`
	Source       string  `csv:"source"`
	Type         string  `csv:"type"`
	Target       string  `csv:"target"`
	Product      string  `csv:"product"`
	Amount       float32 `csv:"amount"`
	FinType      string  `csv:"fin_type"`
	Status       string  `csv:"status"`
	Charge       float32 `csv:"charge"`
	Refund       float32 `csv:"refund"`
	Comment      string  `csv:"comment"`
	FundStatus   string  `csv:"fund_status"`
}

func (t *AlipayTrans) isIncome() bool {
	return Contains(t.FinType, "收入") || ContainsAny(t.Product, incomeKeyWords...)
}

func (t *AlipayTrans) isInnerTransfer() bool {
	return ContainsAny(t.Target, familyMembers...) || ContainsAny(t.Product, innerTransferKeyWords...)
}

func (t *AlipayTrans) isClosed() bool {
	return Contains(t.Status, "交易关闭")
}

func (t *AlipayTrans) yearMonth() string {
	if t.ID[:2] != "20" {
		return "20" + t.ID[:4]
	}
	return t.ID[:6]
}

type TransGroup struct {
	Total     float32
	TransList []*AlipayTrans
}

func (g *TransGroup) add(trans *AlipayTrans) {
	g.Total += trans.Amount
	g.TransList = append(g.TransList, trans)
}

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
	statsMap   = make(map[string]*MonthStat)
	baseDir    string
	yearMonths []string
)

func getMonthStat(yearMonth string) *MonthStat {
	ms, ok := statsMap[yearMonth]
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
		statsMap[yearMonth] = ms
		yearMonths = append(yearMonths, yearMonth)
	}
	return ms
}

func genHtmlReport() {
	for _, yearMonth := range yearMonths {
		f, err := os.Create(baseDir + "/alipay-stat-" + yearMonth + ".html")
		if err != nil {
			panic(err)
		}
		genMonthStatReport(f, statsMap[yearMonth])
		f.Close()
	}
}

func main() {
	if len(os.Args) < 1 {
		log.Fatal("请提供支付宝账当下载文件地址! 登录网页版 https://www.alipay.com/ 去下载！")
	}
	transFile := os.Args[1]
	file, err := os.Open(transFile)
	if err != nil {
		log.Fatalf("打开文件错误! %v", err)
	}
	defer file.Close()

	baseDir = filepath.Dir(transFile)
	log.Println("base dir:", baseDir)

	transHeader, err := csvutil.Header(AlipayTrans{}, "csv")
	if err != nil {
		log.Fatalf("程序错误! %v", err)
	}
	csvReader := csv.NewReader(file)
	dec, err := csvutil.NewDecoder(csvReader, transHeader...)
	if err != nil {
		log.Fatalf("创建解析器失败! %v", err)
	}

	for {
		var t AlipayTrans
		trans := &t
		if err := dec.Decode(trans); err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("解析数据失败! %v", err)
		}

		getMonthStat(trans.yearMonth()).add(trans)
	}

	genHtmlReport()
}
