package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vogo/aliwepaystat"
)

// TransOutput is a JSON-friendly representation of a transaction.
type TransOutput struct {
	ID          string  `json:"id"`
	OrderID     string  `json:"order_id"`
	Platform    string  `json:"platform"`
	YearMonth   string  `json:"year_month"`
	CreatedTime string  `json:"created_time"`
	Source      string  `json:"source"`
	Type        string  `json:"type"`
	Target      string  `json:"target"`
	Product     string  `json:"product"`
	Amount      float64 `json:"amount"`
	FinType     string  `json:"fin_type"`
	Status      string  `json:"status"`
	Comment     string  `json:"comment"`
}

func transToOutput(t aliwepaystat.Trans) TransOutput {
	out := TransOutput{
		ID:          t.GetID(),
		OrderID:     t.GetOrderID(),
		CreatedTime: t.GetCreatedTime(),
		Source:      t.GetSource(),
		Type:        t.GetType(),
		Target:      t.GetTarget(),
		Product:     t.GetProduct(),
		Amount:      t.GetAmount(),
		FinType:     t.GetFinType(),
		Status:      t.GetStatus(),
		Comment:     t.GetComment(),
		YearMonth:   t.YearMonth(),
	}
	if dt, ok := t.(*aliwepaystat.DbTrans); ok {
		out.Platform = dt.Platform
	}
	return out
}

const queryUsage = `用法: aliwepaystat query <action> [args...]

操作:
  months                          列出已导入的月份
  stats [<year-month>]            查看统计（所有月份或指定月份）
  transactions --month <ym>       查看指定月份交易明细

交易查询参数:
  --month <year-month>    月份，格式如 202503（必填）
  --category <keyword>    按分类关键词筛选
  --limit <n>             最大返回条数（默认: 50）

示例:
  aliwepaystat query months
  aliwepaystat query stats
  aliwepaystat query stats 202503
  aliwepaystat query transactions --month 202503
  aliwepaystat query transactions --month 202503 --category 美团`

func queryError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n\n%s\n", msg, queryUsage)
	os.Exit(2)
}

func runQuery(ctx *appContext, args []string) {
	if len(args) == 0 {
		queryError("缺少查询操作")
	}

	action := args[0]
	actionArgs := args[1:]

	switch action {
	case "months":
		runQueryMonths(ctx)
	case "stats":
		runQueryStats(ctx, actionArgs)
	case "transactions":
		runQueryTransactions(ctx, actionArgs)
	default:
		queryError(fmt.Sprintf("未知查询操作: %s", action))
	}
}

func runQueryMonths(ctx *appContext) {
	yms, err := aliwepaystat.QueryYearMonths(ctx.db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if ctx.jsonOutput {
		if yms == nil {
			yms = []string{}
		}
		if err := printJSON(os.Stdout, yms); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	for _, ym := range yms {
		fmt.Println(ym)
	}
}

func runQueryStats(ctx *appContext, args []string) {
	if len(args) > 0 {
		// Single month stats
		ym := args[0]
		if err := aliwepaystat.UpdateMonthStats(ctx.db, ym); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating stats: %v\n", err)
			os.Exit(1)
		}
		stats, err := aliwepaystat.GetMonthStats(ctx.db, ym)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if ctx.jsonOutput {
			if err := printJSON(os.Stdout, stats); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		printStatsTable(os.Stdout, []aliwepaystat.MonthStatData{*stats})
		return
	}

	// All months
	statsList, err := aliwepaystat.GetAllMonthStats(ctx.db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if ctx.jsonOutput {
		if statsList == nil {
			statsList = []aliwepaystat.MonthStatData{}
		}
		if err := printJSON(os.Stdout, statsList); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	printStatsTable(os.Stdout, statsList)
}

func printStatsTable(w io.Writer, statsList []aliwepaystat.MonthStatData) {
	headers := []string{"YearMonth", "Income", "Expense", "Eat", "Travel", "WEG", "Tel", "Other"}
	var rows [][]string
	for _, s := range statsList {
		rows = append(rows, []string{
			s.YearMonth,
			fmt.Sprintf("%.2f", s.TotalIncome),
			fmt.Sprintf("%.2f", s.TotalExpense),
			fmt.Sprintf("%.2f", s.ExpenseEatTotal),
			fmt.Sprintf("%.2f", s.ExpenseTravelTotal),
			fmt.Sprintf("%.2f", s.ExpenseWaterElectGasTotal),
			fmt.Sprintf("%.2f", s.ExpenseTelTotal),
			fmt.Sprintf("%.2f", s.ExpenseOtherTotal),
		})
	}
	printTable(w, headers, rows)
}

const transactionsUsage = `用法: aliwepaystat query transactions --month <year-month> [--category <keyword>] [--limit <n>]

参数:
  --month <year-month>    月份，格式如 202503（必填）
  --category <keyword>    按分类关键词筛选
  --limit <n>             最大返回条数（默认: 50）

示例:
  aliwepaystat query transactions --month 202503
  aliwepaystat query transactions --month 202503 --category 美团
  aliwepaystat query transactions --month 202503 --limit 100`

func transactionsError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n\n%s\n", msg, transactionsUsage)
	os.Exit(2)
}

func runQueryTransactions(ctx *appContext, args []string) {
	fs := flag.NewFlagSet("query transactions", flag.ContinueOnError)
	month := fs.String("month", "", "")
	category := fs.String("category", "", "")
	limit := fs.Int("limit", 50, "")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s\n", transactionsUsage)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if *month == "" {
		transactionsError("缺少 --month 参数，请指定查询月份")
	}

	transList, err := aliwepaystat.QueryTransByYearMonth(ctx.db, *month)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Filter by category keyword if specified
	if *category != "" {
		var filtered []aliwepaystat.Trans
		for _, t := range transList {
			text := t.GetProduct() + " " + t.GetTarget() + " " + t.GetComment()
			if containsIgnoreCase(text, *category) {
				filtered = append(filtered, t)
			}
		}
		transList = filtered
	}

	// Apply limit
	if *limit > 0 && len(transList) > *limit {
		transList = transList[:*limit]
	}

	if ctx.jsonOutput {
		outputs := make([]TransOutput, 0, len(transList))
		for _, t := range transList {
			outputs = append(outputs, transToOutput(t))
		}
		if err := printJSON(os.Stdout, outputs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	headers := []string{"Time", "Target", "Product", "Amount", "Type", "Status"}
	var rows [][]string
	for _, t := range transList {
		rows = append(rows, []string{
			t.GetCreatedTime(),
			t.GetTarget(),
			t.GetProduct(),
			fmt.Sprintf("%.2f", t.GetAmount()),
			t.GetFinType(),
			t.GetStatus(),
		})
	}
	printTable(os.Stdout, headers, rows)
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
