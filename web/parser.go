package web

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/vogo/aliwepaystat"
	"golang.org/x/text/transform"
)

// CSVParser CSV解析器
type CSVParser struct {
	db *sql.DB
}

// NewCSVParser 创建CSV解析器
func NewCSVParser(db *sql.DB) *CSVParser {
	return &CSVParser{db: db}
}

// ParseAndImportCSV 解析CSV内容并导入数据库
func (p *CSVParser) ParseAndImportCSV(content []byte, parser aliwepaystat.TransParser, platform string) (int, error) {
	// 获取现有的交易ID以避免重复导入
	existingIDs := aliwepaystat.LoadExistingIDs(p.db)

	// 创建编码转换器
	transformReader := transform.NewReader(bytes.NewReader(content), parser.Enc().NewDecoder())
	reader := bufio.NewReader(transformReader)

	buf := bytes.NewBuffer(nil)
	dataLineStarted := false

	// 读取文件内容，找到数据开始行
	for {
		lineData, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, fmt.Errorf("读取文件错误: %v", err)
		}

		if len(lineData) == 0 {
			continue
		}

		if dataLineStarted {
			// 处理数据行
			lineData = aliwepaystat.ReplaceCsvLineFieldsSuffixBlank(lineData)
			line := string(lineData)
			if len(strings.Split(line, ",")) != parser.FieldNum() {
				continue // 跳过格式不正确的行
			}
			buf.Write(lineData)
			buf.WriteByte('\n')
			continue
		}

		// 检查是否到达CSV头部
		line := string(lineData)
		line = strings.ReplaceAll(line, " ", "")
		if line == parser.CsvHeader() {
			dataLineStarted = true
			continue
		}
	}

	// 解析CSV数据
	formattedReader := bytes.NewReader(buf.Bytes())
	csvReader := csv.NewReader(formattedReader)
	csvReader.TrimLeadingSpace = true

	count := 0
	for {
		fields, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // 跳过解析失败的行
		}
		trans, err := parser.ParseRow(fields)
		if err != nil {
			continue // 跳过解析失败的行
		}

		// 检查是否已存在
		id := trans.GetID()
		if _, ok := existingIDs[id]; ok {
			continue
		}

		// 插入数据库
		ym := trans.YearMonth()
		if err := aliwepaystat.InsertTrans(p.db, trans, platform, ym); err != nil {
			continue // 跳过插入失败的记录
		}

		existingIDs[id] = struct{}{}
		count++
	}

	return count, nil
}
