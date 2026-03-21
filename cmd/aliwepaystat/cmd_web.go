package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/vogo/aliwepaystat"
	"github.com/vogo/aliwepaystat/web"
)

const webUsage = `用法: aliwepaystat web [--port <port>]

参数:
  --port <port>    服务端口号（默认: 使用配置或随机端口）

示例:
  aliwepaystat web
  aliwepaystat web --port 8080`

func webError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n\n%s\n", msg, webUsage)
	os.Exit(2)
}

func runWeb(ctx *appContext, args []string) {
	fs := flag.NewFlagSet("web", flag.ContinueOnError)
	port := fs.Int("port", 0, "")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s\n", webUsage)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if fs.NArg() > 0 {
		webError(fmt.Sprintf("多余的参数: %s", fs.Arg(0)))
	}

	// Override port in SQLite config if specified
	if *port > 0 {
		if err := aliwepaystat.SetConfig(ctx.db, "server.port", strconv.Itoa(*port), "CLI override"); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting port config: %v\n", err)
			os.Exit(1)
		}
	}

	log.Println("Starting web server...")
	server := web.NewServer(ctx.db)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting web server: %v\n", err)
		os.Exit(1)
	}
}
