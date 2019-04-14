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
	innerTransferKeyWords = []string{"余额宝-自动转入"}
	incomeKeyWords        = []string{"收入", "红包奖励发放", "收益发放"}

	eatKeyWords           = []string{"美团", "口碑", "菜", "餐", "食", "饭", "面", "超市", "汉堡", "安德鲁森", "节奏者"}
	travelKeyWords        = []string{"出行", "交通", "公交", "车", "打的", "的士", "taxi", "滴滴"}
	waterElectGasKeyWords = []string{"水费", "电费", "燃气"}
	telKeyWords           = []string{"话费", "电信", "移动", "联通", "手机充值"}
	transferKeyWords      = []string{"转账"}

	familyMembers = []string{"honey kiera", "望哥", "守望人之初"}
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
	return t.FinType == "收入" || ContainsAny(t.Product, incomeKeyWords...)
}

func (t *AlipayTrans) isInnerTransfer() bool {
	return ContainsAny(t.Product, innerTransferKeyWords...)
}

func (t *AlipayTrans) isClosed() bool {
	return Contains(t.Status, "交易关闭")
}

func (t *AlipayTrans) yearMonth() string {
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
	YearMonth            string
	TotalIncome          float32
	TotalExpense         float32
	Transfer             *TransGroup
	HuabeiRepayment      *TransGroup
	TravelExpense        *TransGroup
	EatExpense           *TransGroup
	WaterElectGasExpense *TransGroup
	TelExpense           *TransGroup
	OtherExpense         *TransGroup
}

func (ms *MonthStat) add(trans *AlipayTrans) {
	//ignore
	// 1. inner transfer
	// 2. closed
	if trans.isInnerTransfer() || trans.isClosed() {
		return
	}

	if trans.isIncome() {
		ms.TotalIncome += trans.Amount
		return
	}

	// 花呗还款占不统计到消费中
	if strings.Contains(trans.Product, "自动还款-花呗") {
		ms.HuabeiRepayment.add(trans)
		return
	}

	if ContainsAny(trans.Product, transferKeyWords...) || ContainsAny(trans.Target, familyMembers...) {
		ms.Transfer.add(trans)
		return
	}

	ms.TotalExpense += trans.Amount

	if ContainsAny(trans.Product, travelKeyWords...) {
		ms.TravelExpense.add(trans)
	} else if ContainsAny(trans.Product, eatKeyWords...) || ContainsAny(trans.Target, eatKeyWords...) {
		ms.EatExpense.add(trans)
	} else if ContainsAny(trans.Product, waterElectGasKeyWords...) || ContainsAny(trans.Target, waterElectGasKeyWords...) {
		ms.WaterElectGasExpense.add(trans)
	} else if ContainsAny(trans.Product, telKeyWords...) || ContainsAny(trans.Target, telKeyWords...) {
		ms.TelExpense.add(trans)
	} else {
		ms.OtherExpense.add(trans)
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
		ms = &MonthStat{YearMonth: yearMonth,
			Transfer:             &TransGroup{},
			HuabeiRepayment:      &TransGroup{},
			TravelExpense:        &TransGroup{},
			EatExpense:           &TransGroup{},
			WaterElectGasExpense: &TransGroup{},
			TelExpense:           &TransGroup{},
			OtherExpense:         &TransGroup{},
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
		log.Fatal("请提供支付宝账当下载文件地址!")
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
