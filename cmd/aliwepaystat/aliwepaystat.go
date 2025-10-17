package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/vogo/aliwepaystat"
	"github.com/vogo/aliwepaystat/web"
)

func main() {
    webMode := flag.Bool("web", false, "启动Web服务器模式")
    configFilePath := flag.String("c", "", "config file path")
    transFileDir := flag.String("d", "", "transaction file directory")
    dbFilePath := flag.String("db", "", "sqlite db file path")
    flag.Parse()

    // 确定数据库路径
    dbPath := *dbFilePath
    if dbPath == "" {
        exe, err := os.Executable()
        if err != nil {
            log.Fatal(err)
        }
        baseDir := filepath.Dir(exe)
        dbPath = filepath.Join(baseDir, "aliwepaystat.db")
    }
    log.Println("数据库文件:", dbPath)

    // 打开数据库并确保表结构
    db := aliwepaystat.OpenDB(dbPath)
    aliwepaystat.EnsureSchema(db)

    if *webMode {
        // Web服务器模式
        log.Println("启动Web服务器模式...")
        server := web.NewServer(db)
        if err := server.Start(); err != nil {
            log.Fatal("启动Web服务器失败:", err)
        }
    } else {
        // 传统命令行模式
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
            aliwepaystat.ParseConfig(configPath)
        }

        log.Println("统计输入目录:", baseDir)

        existing := aliwepaystat.LoadExistingIDs(db)
        aliwepaystat.ImportCsvToDB(baseDir, db, existing)

        statDir := filepath.Join(baseDir, "stat")
        if err := os.MkdirAll(statDir, 0770); err != nil && err != os.ErrExist {
            log.Fatal(err)
        }

        aliwepaystat.BuildStatsFromDB(db)
        aliwepaystat.GenHtmlStat(statDir)

        log.Println("统计完成！")
    }
}
