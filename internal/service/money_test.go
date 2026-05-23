package service

import "testing"

func TestParseYuanToFen(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "whole yuan", input: "1500", want: 150000},
		{name: "one decimal", input: "1500.5", want: 150050},
		{name: "two decimals", input: "1500.50", want: 150050},
		{name: "fen only precision", input: "0.01", want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseYuanToFen(tt.input)
			if err != nil {
				t.Fatalf("ParseYuanToFen returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseYuanToFen(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseYuanToFenRejectsInvalidInput(t *testing.T) {
	inputs := []string{"", "abc", "1.234", "-1", ".5"}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if _, err := ParseYuanToFen(input); err == nil {
				t.Fatalf("ParseYuanToFen(%q) returned nil error", input)
			}
		})
	}
}

func TestFormatFen(t *testing.T) {
	tests := []struct {
		name string
		fen  int
		want string
	}{
		{name: "thousands", fen: 150050, want: "1,500.50"},
		{name: "millions", fen: 100000000, want: "1,000,000.00"},
		{name: "small", fen: 5050, want: "50.50"},
		{name: "negative", fen: -150050, want: "-1,500.50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatFen(tt.fen); got != tt.want {
				t.Fatalf("FormatFen(%d) = %q, want %q", tt.fen, got, tt.want)
			}
		})
	}
}

func TestFormatFenAsYuanIntDisplay(t *testing.T) {
	tests := []struct {
		name string
		fen  int
		want string
	}{
		{name: "thousands", fen: 150000, want: "1,500"},
		{name: "millions", fen: 100000000, want: "1,000,000"},
		{name: "negative", fen: -150050, want: "-1,500"},
		{name: "small", fen: 9900, want: "99"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatFenAsYuanIntDisplay(tt.fen); got != tt.want {
				t.Fatalf("FormatFenAsYuanIntDisplay(%d) = %q, want %q", tt.fen, got, tt.want)
			}
		})
	}
}

func TestFormatFenAsYuanInt(t *testing.T) {
	tests := []struct {
		name string
		fen  int
		want string
	}{
		{name: "whole yuan", fen: 150000, want: "1500"},
		{name: "positive fen remainder", fen: 150050, want: "1500"},
		{name: "negative fen remainder", fen: -150050, want: "-1500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatFenAsYuanInt(tt.fen); got != tt.want {
				t.Fatalf("FormatFenAsYuanInt(%d) = %q, want %q", tt.fen, got, tt.want)
			}
		})
	}
}
