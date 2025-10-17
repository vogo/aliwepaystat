# 架构设计文档（aliwepaystat）

## 概述
- 目标：对支付宝与微信账单进行统一解析、分类汇总，并提供 Web 界面进行配置管理、文件上传和统计查看。
- 运行模式：启动本地 Web 服务器（随机端口），自动打开浏览器，提供完整的 Web 应用体验。
- 输入：通过 Web 界面上传 CSV 文件（支付宝以 `alipay` 开头，微信以 `微信` 开头）。
- 输出：在 Web 界面中实时查看统计报表，支持配置管理和数据导出。
- 持久化：所有数据（交易、配置、统计）存储在本地 `SQLite` 数据库中，默认为 `aliwepaystat.db`。

## 目录结构
- `cmd/aliwepaystat/aliwepaystat.go`：应用入口，启动 Web 服务器，自动打开浏览器。
- `alipay.go` / `wechat.go`：两种平台的交易模型与解析器（`TransParser`），含编码、CSV 表头、字段数与特有逻辑。
- `trans.go`：交易抽象接口 `Trans`、解析器接口 `TransParser`、分组结构 `TransGroup`。
- `month.go`：核心分类与汇总逻辑，月度统计结构 `MonthStat` 与分类规则。
- `db.go`：SQLite 连接与表结构管理（交易、配置、统计表）、查询与写入。
- `config.go`：配置管理，从数据库读取配置项，支持 Web 界面动态修改。
- `web/`：Web 应用相关文件
  - `server.go`：HTTP 服务器实现，路由定义
  - `handlers.go`：Web 请求处理器（上传、配置、统计查看）
  - `static/`：静态资源（CSS、JS、图片）
  - `templates/`：HTML 模板文件
- `utils.go`：工具函数（字符串处理、数值计算等）。
- `utils.go`：工具方法，包含关键词匹配、金额格式化、投资识别等。
- `config.go`：配置解析，支持自定义关键词与显示阈值；内置默认配置。
- `layout.html` / `index.html` / `month-stat.html`：HTML 模板文件（被嵌入到程序）。

## 运行流程
1. **启动服务**：应用启动时初始化 SQLite 数据库，创建必要的表结构（交易、配置、统计）。
2. **Web 服务器**：启动 HTTP 服务器，监听随机可用端口，自动打开默认浏览器。
3. **配置管理**：用户通过 Web 界面查看和修改配置项（关键词、阈值等），实时保存到数据库。
4. **文件上传**：用户通过 Web 界面上传 CSV 文件，系统自动识别平台并解析。
5. **数据处理**：解析 CSV 文件，将交易数据存储到数据库，同时更新统计数据。
6. **统计查看**：用户通过 Web 界面查看实时统计报表，支持按月份、分类筛选。
7. **数据导出**：支持将统计结果导出为 JSON 或 CSV 格式。

## 数据模型

### 数据库表结构

#### 1. transactions 表（交易记录）
```sql
CREATE TABLE transactions (
    id TEXT PRIMARY KEY,
    order_id TEXT,
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
    platform TEXT,
    year_month TEXT
);
CREATE INDEX idx_year_month ON transactions(year_month);
```

#### 2. config 表（配置项）
```sql
CREATE TABLE config (
    key TEXT PRIMARY KEY,
    value TEXT,
    description TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 3. month_stats 表（月度统计）
```sql
CREATE TABLE month_stats (
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
```

#### 4. category_stats 表（分类明细统计）
```sql
CREATE TABLE category_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year_month TEXT,
    category TEXT,
    subcategory TEXT,
    total_amount REAL,
    transaction_count INTEGER,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_category_year_month ON category_stats(year_month, category);
```

### 应用层数据结构
- **交易接口 `Trans`**：统一支付宝与微信的交易抽象，包含金额、时间、分类判断等方法。
- **解析器接口 `TransParser`**：定义平台特定的 CSV 解析规则（表头、字段数、编码）。
- **月度统计 `MonthStat`**：按优先级分组的交易汇总（放款、还款、投资、内部转账、收入、转账收入、转账支出、各类支出）。
- **交易分组 `TransGroup`**：单一分类下的交易列表与总金额。
- `Trans` 接口（统一访问交易属性与判定方法）：
  - 判定：`IsIncome`、`IsInnerTransfer`、`IsTransfer`、`IsClosed`、`YearMonth`。
  - 访问：`GetID`、`GetCreatedTime`、`GetTarget`、`GetProduct`、`GetAmount` 等。
- 平台交易结构体：
  - `AlipayTrans`：字段直接为 typed 值（含 `FinType`、`FundStatus` 等）。
  - `WechatTrans`：金额字段 `Amount`（字符串）在 `GetAmount` 中解析为浮点 `Amt`。
- 数据库存储结构：统一 `transactions` 表（含平台字段 `platform` 与 `year_month`）。读取时构造统一的 `DbTrans` 实例实现 `Trans` 接口，以适配既有分类逻辑。
- `TransGroup`：交易分组，含 `Total` 与 `TransList`，提供 `add` 与 `FormatTotal`。
- `MonthStat`：月度统计与所有分组，包含 `ExpenseTotal` 与各分组的 `TransGroup`。

## Web API 接口

### 配置管理
- `GET /api/config` - 获取所有配置项
- `PUT /api/config` - 更新配置项
- `POST /api/config/reset` - 重置为默认配置

### 文件上传
- `POST /api/upload` - 上传 CSV 文件并解析
- `GET /api/upload/status` - 获取上传处理状态

### 统计查询
- `GET /api/stats/overview` - 获取总览统计
- `GET /api/stats/monthly/:yearMonth` - 获取指定月份统计
- `GET /api/stats/categories` - 获取分类统计
- `GET /api/stats/export/:format` - 导出统计数据（JSON/CSV）

### 交易管理
- `GET /api/transactions` - 查询交易记录（支持分页、筛选）
- `DELETE /api/transactions/:id` - 删除指定交易
- `PUT /api/transactions/:id` - 更新交易信息

## 配置项（config.properties）
- `key.words.loan`：贷款关键词。
- `key.words.transfer`：转账关键词。（控制收入中的“转账收入”，支出中的“转账支出”）
- `key.words.inner-transfer`：内部转账关键词。
- `key.words.income`：普通收入关键词。
- `key.words.repayment`：信用还款关键词。
- `key.words.loan-repayment`：贷款还款关键词。
- `key.words.eat` / `travel` / `water-elect-gas` / `tel`：支出分类关键词。
- `key.words.family`：家庭成员列表（用于内部转账识别）。
- `list.min.amount`：明细展示的最小金额阈值（不影响累计）。
- 说明：配置会覆盖默认值，所有关键词按包含关系匹配。

## 模板与渲染
- 模板管理：
  - `template.go` 通过 `go:embed` 内嵌 `layout.html`、`index.html`、`month-stat.html`。
  - 统一以 `layout` 为母板，渲染内容区域 `content`。
- 数据绑定：
  - 总览：`yearMonths` 与 `monthStatsMap`。
  - 月度明细：`MonthStat`，循环各 `TransGroup` 生成表格。
  - 展示过滤：`Trans.IsShowInList` 依据 `list.min.amount` 控制。

## 关键依赖
- `github.com/jszwec/csvutil`：CSV 到结构体的高效映射（基于标签）。
- `golang.org/x/text/transform` + `encoding`：处理不同平台的文件编码（Alipay: `GBK`，Wechat: `UTF-8`）。
- `database/sql` + `github.com/mattn/go-sqlite3`：SQLite 持久层实现。
- 标准库：`flag`、`log`、`os`、`path/filepath`、`html/template`、`sort`、`bufio`、`encoding/csv` 等。

## 错误处理与边界
- 开闭与失败交易：`IsClosed` 直接忽略。
- 去重：`TransMap` 基于 `GetID` 防止重复累计。
- 非数据行：在定位到表头前后均打印但忽略（如描述或空行）。
- 编码异常与金额解析异常：记录日志并中断（必要时报错）。
- 年月提取：支付宝用交易号截取（兼容非“20”前缀），微信用交易时间 `YYYY-MM`。
- 持久化一致性：导入阶段同时计算并存储 `year_month`，避免平台差异导致的二次计算偏差；插入使用 `INSERT OR IGNORE` 防重。

## 安全与隐私
- 本地计算与生成报表，不上传数据；仅读取所给目录内文件。
- 输出为静态 HTML 文件，查看时不含远程网络访问。

## 性能与可维护性
- 适配日常规模的账单文件；线性遍历与分组，内存占用与时间复杂度随交易条数线性增长。
- 分类通过关键词与少量正则，易于通过配置扩展；新平台解析可通过实现 `TransParser` 扩展。
- 持久化后可增量导入，读取统计仅访问数据库，便于扩展查询与缓存。

## 扩展建议
- 新平台：新增 `Trans` 实现与 `TransParser`，在目录扫描时增加识别与分派逻辑。
- 分类优化：
  - 调整或细化关键词；
  - 增加新的分组（例如医疗、教育）；
  - 引入更精细的投资/还款识别策略。
- 交互增强：生成 JSON 数据与 Web 前端页面，以支持交互式筛选与图表。
- 数据库增强：引入索引优化、视图或物化视图；提供导出为 CSV/JSON 的命令；增加校验或清理工具。

## 构建与使用
- 构建：
  - `make build` 生成各平台发行包到 `dist/`。
  - `go install cmd/aliwepaystat/aliwepaystat.go` 安装到本地环境。
- 使用：
  - 将 CSV 与可执行文件放到同一目录，或用 `-d` 指定账单目录；可选 `-c` 指定配置文件；可选 `-db` 指定数据库文件。
  - 首次运行会导入 CSV 至数据库；随后读取数据库进行统计并生成报表，在 `stat/` 目录查看 `aliwepaystat-index.html` 与各月明细页面。

## 示例数据流（文字版时序）
1. CLI 读取参数，定位账单目录、配置文件与数据库文件。
2. 解析配置，得到关键词与阈值。
3. 初始化 SQLite，加载已存在交易 `id` 到内存集合。
4. 扫描目录并解析 CSV，将新交易入库（含 `year_month`、`platform`）。
5. 从数据库读取所有月份与对应交易，调用 `MonthStat.add` 分类与累计，构建统计。
6. 渲染总览与月度明细模板到 `stat/`，用户打开 `aliwepaystat-index.html` 查看结果。