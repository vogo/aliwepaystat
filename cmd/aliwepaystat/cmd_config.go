package main

import (
	"fmt"
	"os"
)

const configUsage = `用法: aliwepaystat config <action> [args...]

操作:
  list                 列出所有配置项
  get <key>            获取配置项的值
  set <key> <value>    设置配置项的值

支持的配置项: db, dir

示例:
  aliwepaystat config list
  aliwepaystat config get db
  aliwepaystat config set db /path/to/data.db`

func configError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n\n%s\n", msg, configUsage)
	os.Exit(2)
}

func runConfig(ctx *appContext, args []string) {
	if len(args) == 0 {
		configError("缺少操作参数")
	}

	action := args[0]
	actionArgs := args[1:]

	switch action {
	case "list":
		runConfigList(ctx)
	case "get":
		runConfigGet(ctx, actionArgs)
	case "set":
		runConfigSet(ctx, actionArgs)
	default:
		configError(fmt.Sprintf("未知操作: %s", action))
	}
}

func runConfigList(ctx *appContext) {
	if ctx.jsonOutput {
		data := map[string]string{
			"db":  ctx.appConfig.DBPath,
			"dir": ctx.appConfig.Dir,
		}
		if err := printJSON(os.Stdout, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	printKeyValue(os.Stdout, [][2]string{
		{"db", ctx.appConfig.DBPath},
		{"dir", ctx.appConfig.Dir},
	})
}

func runConfigGet(ctx *appContext, args []string) {
	if len(args) == 0 {
		configError("缺少配置项名称")
	}

	key := args[0]
	var value string
	switch key {
	case "db":
		value = ctx.appConfig.DBPath
	case "dir":
		value = ctx.appConfig.Dir
	default:
		configError(fmt.Sprintf("未知配置项: %s", key))
	}

	if ctx.jsonOutput {
		data := map[string]string{"key": key, "value": value}
		if err := printJSON(os.Stdout, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Println(value)
}

func runConfigSet(ctx *appContext, args []string) {
	if len(args) < 2 {
		configError("缺少配置项名称和/或值")
	}

	key := args[0]
	value := args[1]

	switch key {
	case "db":
		ctx.appConfig.DBPath = value
	case "dir":
		ctx.appConfig.Dir = value
	default:
		configError(fmt.Sprintf("未知配置项: %s", key))
	}

	if err := saveAppConfig(ctx.configPath, ctx.appConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	if ctx.jsonOutput {
		data := map[string]string{"key": key, "value": value}
		if err := printJSON(os.Stdout, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("Set %s = %s\n", key, value)
}
