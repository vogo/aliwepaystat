package aliwepaystat

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/text/transform"
)

// ImportResult holds the result of a CSV import operation.
type ImportResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
	Errors   int `json:"errors"`
}

// ParserForPlatform returns the TransParser for the given platform string.
// Returns an error if the platform is not recognized.
func ParserForPlatform(platform string) (TransParser, error) {
	switch platform {
	case "alipay":
		return TransParserAlipay, nil
	case "wechat":
		return TransParserWechat, nil
	default:
		return nil, fmt.Errorf("未知的平台类型: %q (支持 \"alipay\" 或 \"wechat\")", platform)
	}
}

// ImportFileToDBWithResult imports a single CSV file into the database
// and returns an ImportResult. The platform must be specified explicitly.
func ImportFileToDBWithResult(filePath string, platform string, db *sql.DB, existingIDs map[string]struct{}) (*ImportResult, error) {
	parser, err := ParserForPlatform(platform)
	if err != nil {
		return nil, err
	}

	return importFileWithResult(filePath, parser, platform, db, existingIDs)
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
