// Copyright 2019 wongoo. All rights reserved.

package aliwepaystat

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"github.com/jszwec/csvutil"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	regexFieldSuffixBlank, _ = regexp.Compile("[ ]+,")
)

func GenHtmlStat(transBaseDir string) {
	log.Println("------------------------------------")
	for _, yearMonth := range yearMonths {
		statFileName := "aliwepaystat-" + yearMonth + ".html"
		f, err := os.Create(transBaseDir + "/" + statFileName)
		if err != nil {
			panic(err)
		}
		genMonthStatReport(f, monthStatsMap[yearMonth])
		f.Close()
		log.Printf("  统计文件: %s", statFileName)
	}
}

func ParseCsvTransDir(baseDir string) {
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		log.Fatalf("读取目录失败! %v", err)
	}

	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".csv") {
			continue
		}
		filePath := baseDir + fileName
		if strings.Contains(fileName, "alipay") {
			ParseCsvTransFile(filePath, TransParserAlipay)
		} else if strings.Contains(fileName, "微信") {
			ParseCsvTransFile(filePath, TransParserWechat)
		} else {
			log.Fatalf("未知的账单文件(文件名需包含\"alipay\"或\"微信\"): %s", filePath)
		}
	}
}

func ParseCsvTransFile(filePath string, parser TransParser) {
	log.Println()
	log.Println("------------------------------------")
	log.Printf("-----> parse file: %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("打开文件错误! %v", err)
	}
	defer file.Close()
	transHeader, err := csvutil.Header(parser.NewTrans(), "csv")
	if err != nil {
		log.Fatalf("程序错误! %v", err)
	}
	transformReader := transform.NewReader(file, parser.Enc().NewDecoder())
	reader := bufio.NewReader(transformReader)

	buf := bytes.NewBuffer(nil)

	dataLineStarted := false
	// loop read the previous content until the csv title
	for {
		bytes, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("读取文件错误! %v", err)
		}
		if len(bytes) == 0 {
			continue
		}
		if dataLineStarted == true {
			bytes = regexFieldSuffixBlank.ReplaceAll(bytes, []byte{','})
			line := string(bytes)
			if len(strings.Split(line, ",")) != parser.FieldNum() {
				printDataDescLine(line)
				continue
			}
			buf.Write(bytes)
			buf.WriteByte('\n')
			continue
		}
		line := string(bytes)
		line = strings.ReplaceAll(line, " ", "")
		if line == parser.CsvHeader() {
			dataLineStarted = true
			continue
		}
		printDataDescLine(line)
	}

	formattedReader := bytes.NewReader(buf.Bytes())
	csvReader := csv.NewReader(formattedReader)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = parser.FieldNum()

	dec, err := csvutil.NewDecoder(csvReader, transHeader...)
	if err != nil {
		log.Fatalf("创建解析器失败! %v", err)
	}
	for {
		trans := parser.NewTrans()
		if err := dec.Decode(trans); err == io.EOF {
			break
		} else if err != nil {
			log.Printf("解析数据失败! %v", err)
			continue
		}

		getMonthStat(trans.YearMonth()).add(trans)
	}
}

func printDataDescLine(line string) {
	if !strings.HasPrefix(line, "----") && !strings.HasPrefix(line, ",,,,") {
		log.Println(line)
	}
}
