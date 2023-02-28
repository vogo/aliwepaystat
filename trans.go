// Copyright 2019 vogo. All rights reserved.

package aliwepaystat

import "golang.org/x/text/encoding"

// Trans transaction
type Trans interface {
	IsIncome() bool
	IsInnerTransfer() bool
	IsTransfer() bool
	IsClosed() bool
	YearMonth() string

	GetID() string
	GetOrderID() string
	GetCreatedTime() string
	GetSource() string
	GetType() string
	GetTarget() string
	GetProduct() string
	GetAmount() float64
	GetFormatAmount() float64
	GetFinType() string
	GetStatus() string
	GetRefund() float64
	GetComment() string
	IsShowInList() bool
}

type TransParser interface {
	NewTrans() Trans
	CsvHeader() string
	FieldNum() int
	Enc() encoding.Encoding
}

type TransGroup struct {
	Total     float64
	TransList []Trans
}

func (g *TransGroup) add(trans Trans) {
	g.Total += trans.GetAmount()
	g.TransList = append(g.TransList, trans)
}

func (g *TransGroup) FormatTotal() float64 {
	return RoundFloat(g.Total)
}
