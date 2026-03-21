package main

import (
	"fmt"
	"os"

	"github.com/vogo/aliwepaystat"
)

const importUsage = `用法: aliwepaystat import -t <alipay|wechat> <file>

参数:
  -t, --type    平台类型，必须为 alipay 或 wechat
  <file>        CSV 账单文件路径（支持绝对路径或相对路径）

示例:
  aliwepaystat import -t alipay ./alipay_202503.csv
  aliwepaystat import -t wechat /path/to/微信支付账单.csv`

func importError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n\n%s\n", msg, importUsage)
	os.Exit(2)
}

func runImport(ctx *appContext, args []string) {
	// 解析 -t 参数和文件路径
	var platform string
	var filePath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-t", "--type":
			if i+1 < len(args) {
				platform = args[i+1]
				i++
			} else {
				importError("-t 缺少平台类型参数")
			}
		default:
			if filePath == "" {
				filePath = args[i]
			} else {
				importError(fmt.Sprintf("多余的参数: %s", args[i]))
			}
		}
	}

	// 验证平台类型
	if platform == "" {
		importError("缺少 -t 参数，请指定平台类型")
	}

	if platform != "alipay" && platform != "wechat" {
		importError(fmt.Sprintf("无效的平台类型 %q，必须为 alipay 或 wechat", platform))
	}

	// 验证文件路径
	if filePath == "" {
		importError("缺少文件路径")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		importError(fmt.Sprintf("文件不存在或无法访问: %v", err))
	}

	if info.IsDir() {
		importError(fmt.Sprintf("%s 是目录，请指定 CSV 文件路径", filePath))
	}

	existing := aliwepaystat.LoadExistingIDs(ctx.db)

	result, err := aliwepaystat.ImportFileToDBWithResult(filePath, platform, ctx.db, existing)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error importing: %v\n", err)
		os.Exit(1)
	}

	if ctx.jsonOutput {
		if err := printJSON(os.Stdout, result); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("Imported:        %d\n", result.Imported)
	fmt.Printf("Skipped:         %d\n", result.Skipped)
	fmt.Printf("Errors:          %d\n", result.Errors)
}
