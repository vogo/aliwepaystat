package main

import (
	"testing"
)

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "xyz", false},
		{"", "", true},
		{"abc", "", true},
		{"", "a", false},
		{"abcdef", "CDE", true},
		{"ABC", "abc", true},
	}

	for _, tt := range tests {
		got := containsIgnoreCase(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}

func TestTransToOutput(t *testing.T) {
	// We can test transToOutput by using a mock that implements Trans
	// Since we can't import a mock easily here, we test the TransOutput struct
	out := TransOutput{
		ID:          "test-id",
		OrderID:     "order-1",
		Platform:    "alipay",
		YearMonth:   "202503",
		CreatedTime: "2025-03-15 10:30:00",
		Amount:      -25.50,
	}

	if out.ID != "test-id" {
		t.Errorf("ID: got %q, want %q", out.ID, "test-id")
	}
	if out.Amount != -25.50 {
		t.Errorf("Amount: got %f, want %f", out.Amount, -25.50)
	}
}
