package web

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/jszwec/csvutil"
	"golang.org/x/text/transform"
	. "github.com/vogo/aliwepaystat"
)

var (
	regexCsvLineFieldsSuffixBlank, _ = regexp.Compile("[ ]+,")
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
func (p *CSVParser) ParseAndImportCSV(content []byte, parser TransParser, platform string) (int, error) {
	// 获取现有的交易ID以避免重复导入
	existingIDs := LoadExistingIDs(p.db)
	
	// 获取交易结构的CSV头部信息
	transHeader, err := csvutil.Header(parser.NewTrans(), "csv")
	if err != nil {
		return 0, fmt.Errorf("获取CSV头部失败: %v", err)
	}
	
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
			lineData = replaceCsvLineFieldsSuffixBlank(lineData)
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
	
	dec, err := csvutil.NewDecoder(csvReader, transHeader...)
	if err != nil {
		return 0, fmt.Errorf("创建CSV解析器失败: %v", err)
	}
	
	count := 0
	for {
		trans := parser.NewTrans()
		if err := dec.Decode(trans); err == io.EOF {
			break
		} else if err != nil {
			continue // 跳过解析失败的行
		}
		
		// 检查是否已存在
		id := trans.GetID()
		if _, ok := existingIDs[id]; ok {
			continue
		}
		
		// 插入数据库
		ym := trans.YearMonth()
		if err := InsertTrans(p.db, trans, platform, ym); err != nil {
			continue // 跳过插入失败的记录
		}
		
		existingIDs[id] = struct{}{}
		count++
	}
	
	return count, nil
}

// replaceCsvLineFieldsSuffixBlank 替换CSV行字段后缀空白
func replaceCsvLineFieldsSuffixBlank(lineData []byte) []byte {
	return regexCsvLineFieldsSuffixBlank.ReplaceAll(lineData, []byte{','})
}