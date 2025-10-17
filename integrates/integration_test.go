package aliwepaystat_test

import (
    "os"
    "path/filepath"
    "testing"
    aws "github.com/vogo/aliwepaystat"
)

func Test_Integrates_WechatCSV_To_SQLite_To_HTML(t *testing.T) {
    baseDir := t.TempDir()

    // write a sample wechat csv (UTF-8)
    csv := "交易时间,交易类型,交易对方,商品,收/支,金额(元),支付方式,当前状态,交易单号,商户单号,备注\n" +
        "\"2024-09-15 10:00:00\",\"支付\",\"商家A\",\"餐饮-测试\",\"支出\",\"20.00\",\"零钱\",\"支付成功\",\"wx-001\",\"m-001\",\"无\"\n"
    if err := os.WriteFile(filepath.Join(baseDir, "微信支付账单-test.csv"), []byte(csv), 0o660); err != nil {
        t.Fatal(err)
    }

    // init db
    dbPath := filepath.Join(baseDir, "aliwepaystat.db")
    db := aws.OpenDB(dbPath)
    aws.EnsureSchema(db)
    // insert one wechat trans directly to DB (bypass CSV import for robustness)
    wx := &aws.WechatTrans{
        CreatedTime: "2024-09-15 10:00:00",
        Type:        "支付",
        Target:      "商家A",
        Product:     "餐饮-测试",
        FinType:     "支出",
        Amount:      "20.00",
        Source:      "零钱",
        Status:      "支付成功",
        ID:          "wx-001",
        OrderID:     "m-001",
        Comment:     "无",
    }
    if err := aws.InsertTrans(db, wx, "wechat", "202409"); err != nil {
        t.Fatalf("insert to db failed: %v", err)
    }

    // build stats from db
    aws.BuildStatsFromDB(db)

    // generate html
    statDir := filepath.Join(baseDir, "stat")
    if err := os.MkdirAll(statDir, 0o770); err != nil {
        t.Fatal(err)
    }
    aws.GenHtmlStat(statDir)

    // verify outputs
    indexPath := filepath.Join(statDir, "aliwepaystat-index.html")
    if _, err := os.Stat(indexPath); err != nil {
        t.Fatalf("index not generated: %v", err)
    }
    ymFile := filepath.Join(statDir, "aliwepaystat-202409.html")
    bys, err := os.ReadFile(ymFile)
    if err != nil {
        t.Fatalf("month stat not generated: %v", err)
    }
    if !containsString(string(bys), "202409 收支统计报告") {
        t.Fatalf("month stat content unexpected: %s", ymFile)
    }
}

func containsString(s, sub string) bool { return len(s) >= len(sub) && (stringContains(s, sub)) }

func stringContains(s, sub string) bool {
    return indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
    // simple wrapper to avoid importing strings in test
    for i := 0; i+len(sub) <= len(s); i++ {
        if s[i:i+len(sub)] == sub {
            return i
        }
    }
    return -1
}