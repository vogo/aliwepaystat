package main

import (
	"encoding/csv"
	"flag"
	"github.com/jszwec/csvutil"
	"io"
	"log"
	"os"
	"path/filepath"
)

func genHtmlReport() {
	for _, yearMonth := range yearMonths {
		statFileName := "alipay-stat-" + yearMonth + ".html"
		f, err := os.Create(transBaseDir + "/" + statFileName)
		if err != nil {
			panic(err)
		}
		genMonthStatReport(f, monthStatsMap[yearMonth])
		f.Close()
		log.Printf("  统计文件: %s", statFileName)
	}
}

func parseCsvTransFile(transFilePath string) {
	file, err := os.Open(transFilePath)
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
}

var (
	transBaseDir string
)

func main() {
	configFilePath := flag.String("c", "", "config file path")
	transFilePath := flag.String("f", "", "alipay transaction record file path")
	flag.Parse()

	if *transFilePath == "" {
		log.Fatal("请提供支付宝账当下载文件地址! 登录网页版 https://www.alipay.com/ 去下载！")
	}

	if *configFilePath != "" {
		parseConfig(*configFilePath)
	}
	transBaseDir = filepath.Dir(*transFilePath)
	log.Println("统计输入目录:", transBaseDir)

	parseCsvTransFile(*transFilePath)

	genHtmlReport()

	log.Println("统计完成！")
}
