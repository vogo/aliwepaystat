package aliwepaystat

import "testing"

func TestContains(t *testing.T) {
	if !Contains("hello world", "world") {
		t.Error("expected true")
	}
	if Contains("hello", "world") {
		t.Error("expected false")
	}
	if !Contains("abc", "") {
		t.Error("empty substr should match")
	}
}

func TestContainsAny(t *testing.T) {
	if !ContainsAny("hello world", "xyz", "world") {
		t.Error("expected true")
	}
	if ContainsAny("hello", "xyz", "abc") {
		t.Error("expected false")
	}
	if ContainsAny("hello") {
		t.Error("no args should return false")
	}
}

func TestEitherContainsAny(t *testing.T) {
	if !EitherContainsAny("hello", "world", "xyz", "llo") {
		t.Error("expected true from s1")
	}
	if !EitherContainsAny("hello", "world", "xyz", "orl") {
		t.Error("expected true from s2")
	}
	if EitherContainsAny("hello", "world", "xyz", "abc") {
		t.Error("expected false")
	}
}

func TestReplaceCsvLineFieldsSuffixBlank(t *testing.T) {
	input := []byte("hello   ,world  ,test")
	result := ReplaceCsvLineFieldsSuffixBlank(input)
	expected := "hello,world,test"
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}

	// no trailing blanks
	input2 := []byte("a,b,c")
	result2 := ReplaceCsvLineFieldsSuffixBlank(input2)
	if string(result2) != "a,b,c" {
		t.Errorf("got %q, want %q", string(result2), "a,b,c")
	}
}

func TestRoundFloat(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{1.23456, 1.2346},
		{0.0, 0.0},
		{100.0, 100.0},
		{1.99999, 2.0},
	}
	for _, tt := range tests {
		got := RoundFloat(tt.input)
		if got != tt.expected {
			t.Errorf("RoundFloat(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestIsInvestment(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"XX财富YY买入ZZ", true},
		{"XX基金YY买入ZZ", true},
		{"XX股票YY买入ZZ", true},
		{"买入基金", false},
		{"财富管理", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsInvestment(tt.input)
		if got != tt.expected {
			t.Errorf("IsInvestment(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
