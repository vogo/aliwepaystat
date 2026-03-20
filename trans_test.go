package aliwepaystat

import "testing"

func TestTransGroupAdd(t *testing.T) {
	g := &TransGroup{}
	trans1 := &AlipayTrans{ID: "001", Amount: 10.0}
	trans2 := &AlipayTrans{ID: "002", Amount: 20.5}

	g.add(trans1)
	g.add(trans2)

	if g.Total != 30.5 {
		t.Errorf("Total = %v, want 30.5", g.Total)
	}
	if len(g.TransList) != 2 {
		t.Errorf("TransList length = %d, want 2", len(g.TransList))
	}
}

func TestTransGroupFormatTotal(t *testing.T) {
	g := &TransGroup{Total: 99.12345}
	if ft := g.FormatTotal(); ft != 99.1235 {
		t.Errorf("FormatTotal() = %v, want 99.1235", ft)
	}
}
