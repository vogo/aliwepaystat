// Copyright 2019 wongoo. All rights reserved.

package aliwepaystat

import (
	"bufio"
	"log"
	"os"
	"strings"
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

	familyMembers []string
)

func ParseConfig(configFile string) {
	props := readProperties(configFile)

	parseConfigValues(props, &loanKeyWords, "loan")
	parseConfigValues(props, &transferKeyWords, "transfer")
	parseConfigValues(props, &innerTransferKeyWords, "inner-transfer")
	parseConfigValues(props, &incomeKeyWords, "income")
	parseConfigValues(props, &repaymentKeyWords, "repayment")
	parseConfigValues(props, &loanRepaymentKeyWords, "loan-repayment")
	parseConfigValues(props, &eatKeyWords, "eat")
	parseConfigValues(props, &travelKeyWords, "travel")
	parseConfigValues(props, &waterElectGasKeyWords, "water-elect-gas")
	parseConfigValues(props, &telKeyWords, "tel")
	parseConfigValues(props, &familyMembers, "family")
}

func parseConfigValues(props ConfigProperties, target *[]string, key string) {
	fullKey := "key.words." + key
	if value, ok := props[fullKey]; ok && value != "" {
		log.Printf("%25s: %s", fullKey, value)
		*target = strings.Split(value, ",")
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
