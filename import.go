package aliwepaystat

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/transform"
)

func ImportCsvToDB(baseDir string, db *sql.DB, existingIDs map[string]struct{}) {
	files, err := os.ReadDir(baseDir)
	if err != nil {
		log.Fatalf("读取目录失败! %v", err)
	}

	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".csv") {
			continue
		}
		filePath := filepath.Join(baseDir, fileName)
		if strings.Contains(fileName, "alipay") {
			importFile(filePath, TransParserAlipay, "alipay", db, existingIDs)
		} else if strings.Contains(fileName, "微信") {
			importFile(filePath, TransParserWechat, "wechat", db, existingIDs)
		} else {
			log.Fatalf("未知的账单文件(文件名需包含\"alipay\"或\"微信\"): %s", filePath)
		}
	}
}

func importFile(filePath string, parser TransParser, platform string, db *sql.DB, existingIDs map[string]struct{}) {
	log.Println()
	log.Println("------------------------------------")
	log.Printf("-----> import file: %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("打开文件错误! %v", err)
	}
	defer func() { _ = file.Close() }()

	transformReader := transform.NewReader(file, parser.Enc().NewDecoder())
	reader := bufio.NewReader(transformReader)
	buf := bytes.NewBuffer(nil)
	dataLineStarted := false
	for {
		lineData, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("读取文件错误! %v", err)
		}
		if len(lineData) == 0 {
			continue
		}
		if dataLineStarted {
			lineData = ReplaceCsvLineFieldsSuffixBlank(lineData)
			line := string(lineData)
			if len(strings.Split(line, ",")) != parser.FieldNum() {
				printDataDescLine(line)
				continue
			}
			buf.Write(lineData)
			buf.WriteByte('\n')
			continue
		}
		line := string(lineData)
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

	for {
		fields, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("解析数据失败! %v", err)
			continue
		}
		trans, err := parser.ParseRow(fields)
		if err != nil {
			log.Printf("解析数据失败! %v", err)
			continue
		}

		id := trans.GetID()
		if _, ok := existingIDs[id]; ok {
			continue
		}
		ym := trans.YearMonth()
		if err := InsertTrans(db, trans, platform, ym); err != nil {
			log.Printf("插入数据库失败: %v", err)
			continue
		}
		existingIDs[id] = struct{}{}
	}
}
