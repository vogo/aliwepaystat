package aliwepaystat

import "testing"

func TestPrintDataDescLine(t *testing.T) {
	// These should not panic (they just log output)
	printDataDescLine("some description line")
	printDataDescLine("----separator----")
	printDataDescLine(",,,,empty")
	printDataDescLine("")
}
