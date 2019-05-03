package aliwepaystat

import (
	"flag"
	"github.com/wongoo/aliwepaystat"
	"log"
	"os"
)

func main() {
	configFilePath := flag.String("c", "", "config file path")
	transFileDir := flag.String("d", "", "transaction file directory")
	flag.Parse()

	if *transFileDir == "" {
		log.Fatal("请提供账单存放目录!")
	}

	if *configFilePath != "" {
		aliwepaystat.ParseConfig(*configFilePath)
	}
	baseDir := *transFileDir
	if baseDir[len(baseDir)-1] != os.PathSeparator {
		baseDir += string(os.PathSeparator)
	}
	log.Println("统计输入目录:", baseDir)

	aliwepaystat.ParseCsvTransDir(baseDir)

	aliwepaystat.GenHtmlStat(baseDir)

	log.Println("统计完成！")
}
