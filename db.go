package aliwepaystat

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func EnsureSchema(db *sql.DB) {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    order_id TEXT,
    platform TEXT,
    year_month TEXT,
    created_time TEXT,
    source TEXT,
    type TEXT,
    target TEXT,
    product TEXT,
    amount REAL,
    fin_type TEXT,
    status TEXT,
    refund REAL,
    comment TEXT,
    fund_status TEXT
);
CREATE INDEX IF NOT EXISTS idx_transactions_year_month ON transactions(year_month);

-- 配置表
CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT,
    description TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 月度统计表
CREATE TABLE IF NOT EXISTS month_stats (
    year_month TEXT PRIMARY KEY,
    total_income REAL DEFAULT 0,
    total_expense REAL DEFAULT 0,
    loan_total REAL DEFAULT 0,
    repayment_total REAL DEFAULT 0,
    investment_total REAL DEFAULT 0,
    inner_transfer_total REAL DEFAULT 0,
    transfer_income_total REAL DEFAULT 0,
    transfer_expense_total REAL DEFAULT 0,
    expense_eat_total REAL DEFAULT 0,
    expense_travel_total REAL DEFAULT 0,
    expense_water_elect_gas_total REAL DEFAULT 0,
    expense_tel_total REAL DEFAULT 0,
    expense_other_total REAL DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 分类统计表
CREATE TABLE IF NOT EXISTS category_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year_month TEXT,
    category TEXT,
    subcategory TEXT,
    total_amount REAL,
    transaction_count INTEGER,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_category_year_month ON category_stats(year_month, category);
`)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化默认配置
	initDefaultConfig(db)
}

func LoadExistingIDs(db *sql.DB) map[string]struct{} {
	ids := make(map[string]struct{})
	rows, err := db.Query("SELECT id FROM transactions")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Fatal(err)
		}
		ids[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return ids
}

func InsertTrans(db *sql.DB, t Trans, platform string, yearMonth string) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO transactions (
            id, order_id, platform, year_month, created_time, source, type, target, product, amount, fin_type, status, refund, comment, fund_status
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.GetID(), t.GetOrderID(), platform, yearMonth, t.GetCreatedTime(), t.GetSource(), t.GetType(), t.GetTarget(), t.GetProduct(), t.GetAmount(), t.GetFinType(), t.GetStatus(), t.GetRefund(), t.GetComment(), getFundStatusIfAny(t),
	)
	return err
}

func getFundStatusIfAny(t Trans) string {
	switch v := t.(type) {
	case *AlipayTrans:
		return v.FundStatus
	default:
		return ""
	}
}

func QueryYearMonths(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT year_month FROM transactions ORDER BY year_month")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var yms []string
	for rows.Next() {
		var ym string
		if err := rows.Scan(&ym); err != nil {
			return nil, err
		}
		yms = append(yms, ym)
	}
	return yms, rows.Err()
}

type DbTrans struct {
	ID           string
	OrderID      string
	Platform     string
	YearMonthStr string
	CreatedTime  string
	Source       string
	Type         string
	Target       string
	Product      string
	Amount       float64
	FinType      string
	Status       string
	Refund       float64
	Comment      string
	FundStatus   string
}

func (t *DbTrans) IsIncome() bool {
	return Contains(t.FinType, "收入") || ContainsAny(t.Product, cfg.IncomeKeyWords...)
}
func (t *DbTrans) IsInnerTransfer() bool {
	return ContainsAny(t.Target, cfg.FamilyMembers...) || ContainsAny(t.Product, cfg.InnerTransferKeyWords...)
}
func (t *DbTrans) IsTransfer() bool {
	return Contains(t.Type, "转账") || Contains(t.FundStatus, "资金转移") || EitherContainsAny(t.Product, t.Target, cfg.TransferKeyWords...) || ContainsAny(t.Target, cfg.FamilyMembers...)
}
func (t *DbTrans) IsClosed() bool    { return ContainsAny(t.Status, "失败", "交易关闭") }
func (t *DbTrans) YearMonth() string { return t.YearMonthStr }

func (t *DbTrans) GetID() string            { return t.ID }
func (t *DbTrans) GetOrderID() string       { return t.OrderID }
func (t *DbTrans) GetCreatedTime() string   { return t.CreatedTime }
func (t *DbTrans) GetSource() string        { return t.Source }
func (t *DbTrans) GetType() string          { return t.Type }
func (t *DbTrans) GetTarget() string        { return t.Target }
func (t *DbTrans) GetProduct() string       { return t.Product }
func (t *DbTrans) GetAmount() float64       { return t.Amount }
func (t *DbTrans) GetFormatAmount() float64 { return RoundFloat(t.Amount) }
func (t *DbTrans) GetFinType() string       { return t.FinType }
func (t *DbTrans) GetStatus() string        { return t.Status }
func (t *DbTrans) GetRefund() float64       { return t.Refund }
func (t *DbTrans) GetComment() string       { return t.Comment }
func (t *DbTrans) IsShowInList() bool       { return t.Amount > cfg.ListMinAmount }

func QueryTransByYearMonth(db *sql.DB, ym string) ([]Trans, error) {
	rows, err := db.Query(`SELECT id, order_id, platform, year_month, created_time, source, type, target, product, amount, fin_type, status, refund, comment, fund_status FROM transactions WHERE year_month = ? ORDER BY created_time`, ym)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var list []Trans
	for rows.Next() {
		var t DbTrans
		if err := rows.Scan(&t.ID, &t.OrderID, &t.Platform, &t.YearMonthStr, &t.CreatedTime, &t.Source, &t.Type, &t.Target, &t.Product, &t.Amount, &t.FinType, &t.Status, &t.Refund, &t.Comment, &t.FundStatus); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

// initDefaultConfig 初始化默认配置项
func initDefaultConfig(db *sql.DB) {
	defaultConfigs := map[string]struct {
		value       string
		description string
	}{
		"key.words.loan":                    {"借款,放款,借给,借出", "贷款关键词"},
		"key.words.repayment":               {"还款,收款", "还款关键词"},
		"key.words.investment":              {"投资,理财,基金,股票", "投资关键词"},
		"key.words.inner.transfer":          {"转账,充值,提现", "内部转账关键词"},
		"key.words.expense.eat":             {"餐饮,美食,外卖,食品", "餐饮支出关键词"},
		"key.words.expense.travel":          {"交通,出行,打车,地铁,公交", "交通出行关键词"},
		"key.words.expense.water.elect.gas": {"水费,电费,燃气,物业", "水电燃气关键词"},
		"key.words.expense.tel":             {"话费,流量,通信", "通信费用关键词"},
		"server.port":                       {"0", "Web服务器端口（0表示随机端口）"},
		"auto.open.browser":                 {"true", "是否自动打开浏览器"},
	}

	for key, config := range defaultConfigs {
		setConfigIfNotExists(db, key, config.value, config.description)
	}
}

// setConfigIfNotExists 仅在配置项不存在时设置
func setConfigIfNotExists(db *sql.DB, key, value, description string) {
	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM config WHERE key = ?", key).Scan(&exists)
	if err != nil {
		log.Printf("Error checking config existence: %v", err)
		return
	}

	if exists == 0 {
		_, err := db.Exec(`
            INSERT INTO config (key, value, description, updated_at) 
            VALUES (?, ?, ?, ?)`,
			key, value, description, time.Now())
		if err != nil {
			log.Printf("Error setting default config %s: %v", key, err)
		}
	}
}

// GetConfig 获取配置项
func GetConfig(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
	return value, err
}

// SetConfig 设置配置项
func SetConfig(db *sql.DB, key, value, description string) error {
	_, err := db.Exec(`
        INSERT OR REPLACE INTO config (key, value, description, updated_at) 
        VALUES (?, ?, ?, ?)`,
		key, value, description, time.Now())
	return err
}

// GetAllConfig 获取所有配置项
func GetAllConfig(db *sql.DB) (map[string]ConfigItem, error) {
	rows, err := db.Query("SELECT key, value, description, updated_at FROM config")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	configs := make(map[string]ConfigItem)
	for rows.Next() {
		var item ConfigItem
		var key string
		err := rows.Scan(&key, &item.Value, &item.Description, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		configs[key] = item
	}

	return configs, nil
}

// ConfigItem 配置项结构
type ConfigItem struct {
	Value       string    `json:"value"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpdateMonthStats 更新月度统计数据
func UpdateMonthStats(db *sql.DB, yearMonth string) error {
	// 获取该月份的所有交易
	transactions, err := QueryTransByYearMonth(db, yearMonth)
	if err != nil {
		return err
	}

	// 计算各项统计数据
	var stats MonthStatData
	stats.YearMonth = yearMonth

	for _, trans := range transactions {
		amount := trans.GetAmount()

		if trans.IsIncome() {
			stats.TotalIncome += amount
		} else if amount < 0 {
			stats.TotalExpense += -amount

			// 根据关键词分类支出
			product := trans.GetProduct()
			comment := trans.GetComment()

			if containsKeywords(db, "key.words.expense.eat", product, comment) {
				stats.ExpenseEatTotal += -amount
			} else if containsKeywords(db, "key.words.expense.travel", product, comment) {
				stats.ExpenseTravelTotal += -amount
			} else if containsKeywords(db, "key.words.expense.water.elect.gas", product, comment) {
				stats.ExpenseWaterElectGasTotal += -amount
			} else if containsKeywords(db, "key.words.expense.tel", product, comment) {
				stats.ExpenseTelTotal += -amount
			} else {
				stats.ExpenseOtherTotal += -amount
			}
		}

		// 特殊分类统计
		if containsKeywords(db, "key.words.loan", trans.GetProduct(), trans.GetComment()) {
			stats.LoanTotal += amount
		} else if containsKeywords(db, "key.words.repayment", trans.GetProduct(), trans.GetComment()) {
			stats.RepaymentTotal += amount
		} else if containsKeywords(db, "key.words.investment", trans.GetProduct(), trans.GetComment()) {
			stats.InvestmentTotal += amount
		} else if trans.IsInnerTransfer() {
			stats.InnerTransferTotal += amount
		} else if trans.IsTransfer() {
			if amount > 0 {
				stats.TransferIncomeTotal += amount
			} else {
				stats.TransferExpenseTotal += -amount
			}
		}
	}

	// 保存到数据库
	return saveMonthStats(db, &stats)
}

// MonthStatData 月度统计数据结构
type MonthStatData struct {
	YearMonth                 string    `json:"year_month"`
	TotalIncome               float64   `json:"total_income"`
	TotalExpense              float64   `json:"total_expense"`
	LoanTotal                 float64   `json:"loan_total"`
	RepaymentTotal            float64   `json:"repayment_total"`
	InvestmentTotal           float64   `json:"investment_total"`
	InnerTransferTotal        float64   `json:"inner_transfer_total"`
	TransferIncomeTotal       float64   `json:"transfer_income_total"`
	TransferExpenseTotal      float64   `json:"transfer_expense_total"`
	ExpenseEatTotal           float64   `json:"expense_eat_total"`
	ExpenseTravelTotal        float64   `json:"expense_travel_total"`
	ExpenseWaterElectGasTotal float64   `json:"expense_water_elect_gas_total"`
	ExpenseTelTotal           float64   `json:"expense_tel_total"`
	ExpenseOtherTotal         float64   `json:"expense_other_total"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

// saveMonthStats 保存月度统计到数据库
func saveMonthStats(db *sql.DB, stats *MonthStatData) error {
	_, err := db.Exec(`
        INSERT OR REPLACE INTO month_stats 
        (year_month, total_income, total_expense, loan_total, repayment_total, 
         investment_total, inner_transfer_total, transfer_income_total, transfer_expense_total,
         expense_eat_total, expense_travel_total, expense_water_elect_gas_total, 
         expense_tel_total, expense_other_total, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		stats.YearMonth, stats.TotalIncome, stats.TotalExpense, stats.LoanTotal, stats.RepaymentTotal,
		stats.InvestmentTotal, stats.InnerTransferTotal, stats.TransferIncomeTotal, stats.TransferExpenseTotal,
		stats.ExpenseEatTotal, stats.ExpenseTravelTotal, stats.ExpenseWaterElectGasTotal,
		stats.ExpenseTelTotal, stats.ExpenseOtherTotal, time.Now())
	return err
}

// GetMonthStats 获取月度统计数据
func GetMonthStats(db *sql.DB, yearMonth string) (*MonthStatData, error) {
	var stats MonthStatData
	err := db.QueryRow(`
        SELECT year_month, total_income, total_expense, loan_total, repayment_total,
               investment_total, inner_transfer_total, transfer_income_total, transfer_expense_total,
               expense_eat_total, expense_travel_total, expense_water_elect_gas_total,
               expense_tel_total, expense_other_total, updated_at
        FROM month_stats WHERE year_month = ?`, yearMonth).Scan(
		&stats.YearMonth, &stats.TotalIncome, &stats.TotalExpense, &stats.LoanTotal, &stats.RepaymentTotal,
		&stats.InvestmentTotal, &stats.InnerTransferTotal, &stats.TransferIncomeTotal, &stats.TransferExpenseTotal,
		&stats.ExpenseEatTotal, &stats.ExpenseTravelTotal, &stats.ExpenseWaterElectGasTotal,
		&stats.ExpenseTelTotal, &stats.ExpenseOtherTotal, &stats.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetAllMonthStats 获取所有月度统计数据
func GetAllMonthStats(db *sql.DB) ([]MonthStatData, error) {
	rows, err := db.Query(`
        SELECT year_month, total_income, total_expense, loan_total, repayment_total,
               investment_total, inner_transfer_total, transfer_income_total, transfer_expense_total,
               expense_eat_total, expense_travel_total, expense_water_elect_gas_total,
               expense_tel_total, expense_other_total, updated_at
        FROM month_stats ORDER BY year_month DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var statsList []MonthStatData
	for rows.Next() {
		var stats MonthStatData
		err := rows.Scan(
			&stats.YearMonth, &stats.TotalIncome, &stats.TotalExpense, &stats.LoanTotal, &stats.RepaymentTotal,
			&stats.InvestmentTotal, &stats.InnerTransferTotal, &stats.TransferIncomeTotal, &stats.TransferExpenseTotal,
			&stats.ExpenseEatTotal, &stats.ExpenseTravelTotal, &stats.ExpenseWaterElectGasTotal,
			&stats.ExpenseTelTotal, &stats.ExpenseOtherTotal, &stats.UpdatedAt)
		if err != nil {
			return nil, err
		}
		statsList = append(statsList, stats)
	}

	return statsList, nil
}

// containsKeywords 检查文本是否包含配置中的关键词
func containsKeywords(db *sql.DB, configKey, product, comment string) bool {
	keywords, err := GetConfig(db, configKey)
	if err != nil {
		return false
	}

	// 分割关键词（以逗号分隔）
	keywordList := splitKeywords(keywords)
	text := product + " " + comment

	for _, keyword := range keywordList {
		if containsString(text, keyword) {
			return true
		}
	}
	return false
}

// splitKeywords 分割关键词字符串
func splitKeywords(keywords string) []string {
	var result []string
	current := ""

	for _, char := range keywords {
		if char == ',' || char == '，' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

// containsString 检查字符串是否包含子串
func containsString(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
