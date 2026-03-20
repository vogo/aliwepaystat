package main

import (
	"fmt"
	"log"
	"os"

	"github.com/vogo/aliwepaystat"
)

func main() {
	// Parse global flags manually before subcommand
	configPath := ""
	jsonOutput := false
	var subArgs []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			} else {
				fmt.Fprintln(os.Stderr, "Error: -c requires a path argument")
				os.Exit(2)
			}
		case "--json":
			jsonOutput = true
		default:
			// First non-flag argument is the subcommand
			subArgs = args[i:]
		}
		if len(subArgs) > 0 {
			break
		}
	}

	if len(subArgs) == 0 {
		printUsage()
		os.Exit(2)
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
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", subcommand)
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: aliwepaystat [-c <config-path>] [--json] <command> [args...]

Commands:
  config    Manage global configuration
  import    Import CSV transaction files
  query     Query transaction data and statistics
  web       Start the web UI server
  help      Show this help message

Global Flags:
  -c <path>   Path to config file (default: ~/.aliwepaystat.conf)
  --json      Output in JSON format (for config, import, query commands)`)
}
