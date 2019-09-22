package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/wongoo/aliwepaystat"
)

func main() {
	configFilePath := flag.String("c", "", "config file path")
	transFileDir := flag.String("d", "", "transaction file directory")
	flag.Parse()

	baseDir := *transFileDir
	if baseDir != "" && baseDir[len(baseDir)-1] != os.PathSeparator {
		baseDir += string(os.PathSeparator)
	}
	if baseDir == "" {
		exe, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		baseDir = filepath.Dir(exe)
	}

	configPath := *configFilePath
	if configPath == "" {
		localConfigPath := filepath.Join(baseDir, "config.properties")
		if _, err := os.Stat(localConfigPath); err == nil {
			configPath = localConfigPath
		}
	}
	if configPath != "" {
		log.Println("配置文件:", configPath)
		aliwepaystat.ParseConfig(*configFilePath)
	}

	log.Println("统计输入目录:", baseDir)

	aliwepaystat.ParseCsvTransDir(baseDir)

	statDir := filepath.Join(baseDir, "stat")
	if err := os.MkdirAll(statDir, 0770); err != nil && err != os.ErrExist {
		log.Fatal(err)
	}

	aliwepaystat.GenHtmlStat(statDir)

	log.Println("统计完成！")
}
