# 支付宝微信账单统计

无现金时代，你几乎所有收支都在支付宝和微信支付里面，将两者合并统计就是你的整体财报了。

本工具将支付宝和微信支付的 CSV 账单导入 SQLite 数据库，后续所有查询和统计分析都基于 SQLite 进行。

## 1. 安装

### 下载

下载地址: https://github.com/vogo/aliwepaystat/releases

### 自行编译

```bash
go build ./cmd/aliwepaystat/
```

## 2. 导出账单

从支付宝和微信支付导出 CSV 账单文件：
- 微信支付账单导出: https://jingyan.baidu.com/article/95c9d20d04e8f8ec4e756182.html
- 支付宝账单导出: https://jingyan.baidu.com/article/00a07f38540b2782d028dc17.html

> 注意：支付宝的账单文件名需以 `alipay` 开头，微信账单文件名需以 `微信` 开头。

## 3. 导入数据库

将 CSV 账单文件导入 SQLite 数据库：

```bash
# 导入单个文件
aliwepaystat import alipay_202503.csv

# 导入目录下所有 CSV 文件
aliwepaystat import /path/to/csv/dir/
```

导入完成后，所有后续操作都基于数据库，原始 CSV 文件可自行归档。

## 4. 查询统计

```bash
# 查看已导入的月份
aliwepaystat query months

# 查看所有月份统计
aliwepaystat query stats

# 查看指定月份统计
aliwepaystat query stats 202503

# 查看指定月份交易明细
aliwepaystat query transactions --month 202503

# 按分类关键词筛选
aliwepaystat query transactions --month 202503 --category 美团

# JSON 格式输出
aliwepaystat --json query stats
```

## 5. Web 界面

启动 Web 服务，通过浏览器查看统计和管理数据：

```bash
aliwepaystat web

# 指定端口
aliwepaystat web --port 8080
```

Web 界面支持上传 CSV、查看统计图表、管理交易和配置分类关键词。

## 6. 配置管理

应用配置文件默认路径为 `~/.aliwepaystat.conf`，可通过 `-c` 参数指定。

```bash
# 查看配置
aliwepaystat config list

# 设置数据库路径
aliwepaystat config set db /path/to/data.db

# 设置默认导入目录
aliwepaystat config set dir /path/to/csv/dir/
```

## 完整用法

```
aliwepaystat [-c <config-path>] [--json] <command> [args...]

Commands:
  config    管理应用配置
  import    导入 CSV 账单文件到数据库
  query     查询交易数据和统计
  web       启动 Web 界面
  help      显示帮助信息

Global Flags:
  -c <path>   配置文件路径 (默认: ~/.aliwepaystat.conf)
  --json      以 JSON 格式输出
```
