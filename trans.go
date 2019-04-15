package main

//AlipayTrans alipay transaction
type AlipayTrans struct {
	ID           string  `csv:"id"`
	OrderID      string  `csv:"order_id"`
	CreatedTime  string  `csv:"created_time"`
	PaidTime     string  `csv:"paid_time"`
	ModifiedTime string  `csv:"modified_time"`
	Source       string  `csv:"source"`
	Type         string  `csv:"type"`
	Target       string  `csv:"target"`
	Product      string  `csv:"product"`
	Amount       float32 `csv:"amount"`
	FinType      string  `csv:"fin_type"`
	Status       string  `csv:"status"`
	Charge       float32 `csv:"charge"`
	Refund       float32 `csv:"refund"`
	Comment      string  `csv:"comment"`
	FundStatus   string  `csv:"fund_status"`
}

type TransGroup struct {
	Total     float32
	TransList []*AlipayTrans
}

func (g *TransGroup) add(trans *AlipayTrans) {
	g.Total += trans.Amount
	g.TransList = append(g.TransList, trans)
}

func (t *AlipayTrans) isIncome() bool {
	return Contains(t.FinType, "收入") || ContainsAny(t.Product, incomeKeyWords...)
}

func (t *AlipayTrans) isInnerTransfer() bool {
	return ContainsAny(t.Target, familyMembers...) || ContainsAny(t.Product, innerTransferKeyWords...)
}

func (t *AlipayTrans) isClosed() bool {
	return Contains(t.Status, "交易关闭")
}

func (t *AlipayTrans) yearMonth() string {
	if t.ID[:2] != "20" {
		return "20" + t.ID[:4]
	}
	return t.ID[:6]
}
