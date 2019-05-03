# 支付宝微信账单统计

无现金时代，你几乎所有收支都在支付宝和微信支付里面，将两者合并统计就是你的整体财报了。

## 账单导出方式
- 微信支付账单导出: https://jingyan.baidu.com/article/95c9d20d04e8f8ec4e756182.html
- 支付宝账单导出: https://jingyan.baidu.com/article/00a07f38540b2782d028dc17.html

导出后解压，将所有csv文件放到一个目录下，形如:
```
~  /Users/wongoo/aliwepaystat > tree
├── alipay_husband.csv
├── alipay_wife.csv
├── 微信支付账单-husband.csv
├── 微信支付账单-wife.csv
```

## 下载工具

```bash
go get github.com/wongoo/aliwepaystat/cmd/aliwepaystat
```

## 自定义配置文件

如 mystatconfig.properties:
```
key.words.loan=放款
key.words.transfer=转账
key.words.inner-transfer=余额宝-自动转入,网商银行转入,余额宝-转出到余额,转出到网商银行
key.words.income=收入,红包奖励发放,收益发放,奖励,退款
key.words.repayment=还款
key.words.loan-repayment=蚂蚁借呗还款
key.words.eat=美团,饿了么,口碑,外卖,菜,餐饮,美食,饭,超市,汉堡,安德鲁森,节奏者,拉面,洪濑鸡爪,肉夹馍,麦之屋,沙县小吃,重庆小面,咖啡,85度C
key.words.travel=出行,交通,公交,车,打的,的士,taxi,滴滴,地铁
key.words.water-elect-gas=水费,电费,燃气
key.words.tel=话费,电信,移动,联通,手机充值
key.words.family=老公,老婆,张仁礼,李德明
```

## 统计

```bash
aliwepaystat -c mystatconfig.properties -d /Users/wongoo/aliwepaystat/
```