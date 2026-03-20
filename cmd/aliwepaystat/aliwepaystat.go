package main

import (
	"fmt"
	"log"
	"os"

	"github.com/vogo/aliwepaystat"
)

func main() {
	// First pass: extract global flags (--json, -c) from anywhere in args
	configPath := ""
	jsonOutput := false
	var remaining []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			} else {
				mainError("-c 缺少配置文件路径参数")
			}
		case "--json":
			jsonOutput = true
		default:
			remaining = append(remaining, args[i])
		}
	}

	subArgs := remaining

	if len(subArgs) == 0 {
		mainError("缺少命令")
	}

	subcommand := subArgs[0]
	cmdArgs := subArgs[1:]

	if subcommand == "help" || subcommand == "--help" || subcommand == "-h" {
		printUsage()
		return
	}

	// Load app config
	if configPath == "" {
		configPath = defaultConfigPath()
	}

	appCfg, err := ensureAppConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	ctx := &appContext{
		configPath: configPath,
		appConfig:  appCfg,
		jsonOutput: jsonOutput,
	}

	// config subcommand does not need database
	if subcommand == "config" {
		runConfig(ctx, cmdArgs)
		return
	}

	// Open database for all other subcommands
	db := aliwepaystat.OpenDB(appCfg.DBPath)
	aliwepaystat.EnsureSchema(db)
	ctx.db = db
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	switch subcommand {
	case "import":
		runImport(ctx, cmdArgs)
	case "query":
		runQuery(ctx, cmdArgs)
	case "web":
		runWeb(ctx, cmdArgs)
	default:
		mainError(fmt.Sprintf("未知命令: %s", subcommand))
	}
}

const mainUsage = `用法: aliwepaystat [-c <config-path>] <command> [args...] [--json]

命令:
  config    管理应用配置
  import    导入 CSV 账单文件 (import -t <alipay|wechat> <file>)
  query     查询交易数据和统计
  web       启动 Web 界面
  help      显示帮助信息

全局参数:
  -c <path>   配置文件路径 (默认: ~/.aliwepaystat.conf)
  --json      以 JSON 格式输出（可放在任意位置）`

func mainError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n\n%s\n", msg, mainUsage)
	os.Exit(2)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, mainUsage)
}
