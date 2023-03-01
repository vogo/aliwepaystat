// Copyright 2019 vogo. All rights reserved.

package aliwepaystat

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	LoanKeyWords          []string
	TransferKeyWords      []string
	InnerTransferKeyWords []string
	IncomeKeyWords        []string
	RepaymentKeyWords     []string
	LoanRepaymentKeyWords []string
	EatKeyWords           []string
	TravelKeyWords        []string
	WaterElectGasKeyWords []string
	TelKeyWords           []string
	FamilyMembers         []string
	ListMinAmount         float64
}

var (
	// default config
	cfg = &Config{
		LoanKeyWords:          []string{"放款"},
		TransferKeyWords:      []string{"转账"},
		InnerTransferKeyWords: []string{"余额宝-自动转入", "网商银行转入", "余额宝-转出到余额", "转出到网商银行"},
		IncomeKeyWords:        []string{"收入", "红包奖励发放", "收益发放", "奖励", "退款", "Collection Bill"},
		RepaymentKeyWords:     []string{"还款"},
		LoanRepaymentKeyWords: []string{"蚂蚁借呗还款"},
		EatKeyWords: []string{"美团", "饿了么", "口碑", "外卖", "菜", "餐饮", "美食", "饭", "超市", "汉堡", "安德鲁森",
			"节奏者", "拉面", "洪濑鸡爪", "肉夹馍", "麦之屋", "沙县小吃", "重庆小面", "咖啡", "85度C"},
		TravelKeyWords:        []string{"出行", "交通", "公交", "车", "打的", "的士", "taxi", "滴滴", "地铁"},
		WaterElectGasKeyWords: []string{"水费", "电费", "燃气"},
		TelKeyWords:           []string{"话费", "电信", "移动", "联通", "手机充值"},
		ListMinAmount:         10.0,
	}
)

func ParseConfig(configFile string) {
	props := readProperties(configFile)

	parseKeyWords(props, &cfg.LoanKeyWords, "loan")
	parseKeyWords(props, &cfg.TransferKeyWords, "transfer")
	parseKeyWords(props, &cfg.InnerTransferKeyWords, "inner-transfer")
	parseKeyWords(props, &cfg.IncomeKeyWords, "income")
	parseKeyWords(props, &cfg.RepaymentKeyWords, "repayment")
	parseKeyWords(props, &cfg.LoanRepaymentKeyWords, "loan-repayment")
	parseKeyWords(props, &cfg.EatKeyWords, "eat")
	parseKeyWords(props, &cfg.TravelKeyWords, "travel")
	parseKeyWords(props, &cfg.WaterElectGasKeyWords, "water-elect-gas")
	parseKeyWords(props, &cfg.TelKeyWords, "tel")
	parseKeyWords(props, &cfg.FamilyMembers, "family")

	parseFloat(props, &cfg.ListMinAmount, "list.min.amount")
}

func parseKeyWords(props ConfigProperties, target *[]string, key string) {
	fullKey := "key.words." + key
	if value, ok := props[fullKey]; ok && value != "" {
		log.Printf("%25s: %s", fullKey, value)
		*target = strings.Split(value, ",")
	}
}

func parseFloat(props ConfigProperties, target *float64, key string) {
	if value, ok := props[key]; ok && value != "" {
		log.Printf("%25s: %s", key, value)
		if f, err := strconv.ParseFloat(value, 64); err != nil {
			*target = f
		}
	}
}

type ConfigProperties map[string]string

func readProperties(filename string) ConfigProperties {
	config := ConfigProperties{}

	if len(filename) == 0 {
		return config
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				config[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return nil
	}

	return config
}
