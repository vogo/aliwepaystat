package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/jszwec/csvutil"
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
	if t.FinType == "收入" || strings.Contains(t.Product, "余额宝-自动转入") {
		return true
	}
	return false
}

func (t *AlipayTrans) yearMonth() string {
	return t.ID[:6]
}

type MonthStat struct {
	yearMonth            string
	totalIncome          float32
	totalExpense         float32
	totalTransfer        float32
	huabeiRepayment      float32
	travelExpense        float32
	eatExpense           float32
	waterElectGasExpense float32
	telExpense           float32
	otherExpense         float32
}


func addBufferAmount(buffer *bytes.Buffer, key string, amount float32) {
	buffer.WriteString(fmt.Sprintf("\t%s: %f", key, amount))
}
func (ms *MonthStat) String() string {
	var buffer bytes.Buffer
	b := &buffer

	b.WriteString("月份: " + ms.yearMonth)
	addBufferAmount(b, "收入", ms.totalIncome)
	addBufferAmount(b, "支出", ms.totalExpense)
	addBufferAmount(b, "转账", ms.totalTransfer)
	addBufferAmount(b, "花呗还款", ms.huabeiRepayment)
	addBufferAmount(b, "交通", ms.travelExpense)
	addBufferAmount(b, "餐饮", ms.eatExpense)
	addBufferAmount(b, "水电", ms.waterElectGasExpense)
	addBufferAmount(b, "话费", ms.telExpense)
	addBufferAmount(b, "其他", ms.otherExpense)
	return buffer.String()
}

var (
	eatKeyWords           = []string{"美团", "口碑", "菜", "餐", "食", "饭", "面", "超市", "麦"}
	travelKeyWords        = []string{"出行", "交通", "公交", "车", "打的", "的士", "taxi"}
	waterElectGasKeyWords = []string{"水费", "电费", "燃气"}
	telKeyWords           = []string{"话费", "电信", "移动", "联通","手机充值"}
)

func (ms *MonthStat) add(trans *AlipayTrans) {
	if trans.isIncome() {
		ms.totalIncome += trans.Amount
		return
	}

	// 花呗还款占不统计到消费中
	if strings.Contains(trans.Product, "自动还款-花呗") {
		ms.huabeiRepayment += trans.Amount
		return
	}

	if Contains(trans.Product, "转账") || Contains(trans.Target, "honey kiera") {
		ms.totalTransfer += trans.Amount
		return
	}

	ms.totalExpense += trans.Amount

	if ContainsAny(trans.Product, travelKeyWords...) {
		ms.travelExpense += trans.Amount
	} else if ContainsAny(trans.Product, eatKeyWords...) || ContainsAny(trans.Target, eatKeyWords...) {
		ms.eatExpense += trans.Amount
	} else if ContainsAny(trans.Product, waterElectGasKeyWords...) || ContainsAny(trans.Target, waterElectGasKeyWords...) {
		ms.waterElectGasExpense += trans.Amount
	} else if ContainsAny(trans.Product, telKeyWords...) || ContainsAny(trans.Target, telKeyWords...) {
		ms.telExpense += trans.Amount
	} else {
		ms.otherExpense += trans.Amount
	}
}

var (
	statsMap = make(map[string]*MonthStat)
)

func getMonthStat(yearMonth string) *MonthStat {
	ms, ok := statsMap[yearMonth]
	if !ok {
		ms = &MonthStat{yearMonth: yearMonth}
		statsMap[yearMonth] = ms
	}
	return ms
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

	for _, v := range statsMap {
		fmt.Println(v)
	}
}
