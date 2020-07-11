# 支付宝微信账单统计

无现金时代，你几乎所有收支都在支付宝和微信支付里面，将两者合并统计就是你的整体财报了。

## 1. 准备工作

### 创建一个账单目录

比如 `/Users/wongoo/aliwepaystat`。

### 自定义配置文件

在账单目录中创建 `config.properties` 文件, 其内容为一些衣食住行的一些关键字，
以便程序能够将一笔交易正确分类，参考内容如下:
```
key.words.loan=放款
key.words.transfer=转账
key.words.inner-transfer=余额宝-自动转入,网商银行转入,余额宝-转出到余额,转出到网商银行
key.words.income=收入,红包奖励发放,收益发放,奖励,退款,收款,Collection Bill
key.words.repayment=还款
key.words.loan-repayment=蚂蚁借呗还款
key.words.eat=美团,饿了么,口碑,外卖,菜,餐饮,美食,饭,超市,汉堡,安德鲁森,节奏者,拉面,洪濑鸡爪,肉夹馍,麦之屋,沙县小吃,重庆小面,咖啡,85度C
key.words.travel=出行,交通,公交,车,打的,的士,taxi,滴滴,地铁
key.words.water-elect-gas=水费,电费,燃气
key.words.tel=话费,电信,移动,联通,手机充值
key.words.family=老公,老婆,张仁礼,李德明

list.min.amount=10.0
```

### 下载统计工具

在账单目录中执行以下命令下载工具
```bash
GOBIN=$(pwd) go get github.com/wongoo/aliwepaystat/cmd/aliwepaystat
```

## 2. 账单导出方式
- 微信支付账单导出: https://jingyan.baidu.com/article/95c9d20d04e8f8ec4e756182.html
- 支付宝账单导出: https://jingyan.baidu.com/article/00a07f38540b2782d028dc17.html

将以上下载的账单导出后解压到此账单目录，结构形如:
```
~  /Users/wongoo/aliwepaystat > tree
├── aliwepaystat            <--- 统计程序, windows下为aliwepaystat.exe
├── config.properties       <--- 交易分类配置文件
├── alipay_husband.csv      <--- 老公支付宝账单文件
├── alipay_wife.csv         <--- 老婆支付宝账单文件
├── 微信支付账单-husband.csv  <--- 老公微信账单文件
├── 微信支付账单-wife.csv     <--- 老婆微信账单文件
```
> 注意：支付宝的账单文件名以`alipay`开头, 微信账单文件名以`微信`开头。


## 3. 开始统计

点击执行 aliwepaystat 开始统计， 
统计结果位于账单目录的 stat 子目录下，点击打开 aliwepaystat-index.html 即可看到统计结果。