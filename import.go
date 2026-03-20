package aliwepaystat

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/transform"
)

// ImportResult holds the result of a CSV import operation.
type ImportResult struct {
	FilesProcessed int `json:"files_processed"`
	Imported       int `json:"imported"`
	Skipped        int `json:"skipped"`
	Errors         int `json:"errors"`
}

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

// ImportCsvToDBWithResult imports CSV files from a directory into the database
// and returns an ImportResult with counts. Unlike ImportCsvToDB, it returns errors
// instead of calling log.Fatal.
func ImportCsvToDBWithResult(baseDir string, db *sql.DB, existingIDs map[string]struct{}) (*ImportResult, error) {
	files, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	result := &ImportResult{}

	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".csv") {
			continue
		}
		filePath := filepath.Join(baseDir, fileName)

		var parser TransParser
		var platform string

		if strings.Contains(fileName, "alipay") {
			parser = TransParserAlipay
			platform = "alipay"
		} else if strings.Contains(fileName, "微信") {
			parser = TransParserWechat
			platform = "wechat"
		} else {
			log.Printf("跳过未知的账单文件(文件名需包含\"alipay\"或\"微信\"): %s", filePath)
			result.Errors++
			continue
		}

		fileResult, err := importFileWithResult(filePath, parser, platform, db, existingIDs)
		if err != nil {
			log.Printf("导入文件失败 %s: %v", filePath, err)
			result.Errors++
			continue
		}
		result.FilesProcessed++
		result.Imported += fileResult.Imported
		result.Skipped += fileResult.Skipped
		result.Errors += fileResult.Errors
	}

	return result, nil
}

// ImportFileToDBWithResult imports a single CSV file into the database
// and returns an ImportResult. It detects the platform from the filename.
func ImportFileToDBWithResult(filePath string, db *sql.DB, existingIDs map[string]struct{}) (*ImportResult, error) {
	fileName := filepath.Base(filePath)

	var parser TransParser
	var platform string

	if strings.Contains(fileName, "alipay") {
		parser = TransParserAlipay
		platform = "alipay"
	} else if strings.Contains(fileName, "微信") {
		parser = TransParserWechat
		platform = "wechat"
	} else {
		return nil, fmt.Errorf("未知的账单文件(文件名需包含\"alipay\"或\"微信\"): %s", filePath)
	}

	fileResult, err := importFileWithResult(filePath, parser, platform, db, existingIDs)
	if err != nil {
		return nil, err
	}
	fileResult.FilesProcessed = 1
	return fileResult, nil
}

// importFileWithResult imports a single file and returns counts.
func importFileWithResult(filePath string, parser TransParser, platform string, db *sql.DB, existingIDs map[string]struct{}) (*ImportResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件错误: %w", err)
	}
	defer func() { _ = file.Close() }()

	result := &ImportResult{}

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
			return nil, fmt.Errorf("读取文件错误: %w", err)
		}
		if len(lineData) == 0 {
			continue
		}
		if dataLineStarted {
			lineData = ReplaceCsvLineFieldsSuffixBlank(lineData)
			line := string(lineData)
			if len(strings.Split(line, ",")) != parser.FieldNum() {
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
			result.Errors++
			continue
		}
		trans, err := parser.ParseRow(fields)
		if err != nil {
			result.Errors++
			continue
		}

		id := trans.GetID()
		if _, ok := existingIDs[id]; ok {
			result.Skipped++
			continue
		}
		ym := trans.YearMonth()
		if err := InsertTrans(db, trans, platform, ym); err != nil {
			result.Errors++
			continue
		}
		existingIDs[id] = struct{}{}
		result.Imported++
	}

	return result, nil
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
