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
	if got := FormatFen(150050); got != "1500.50" {
		t.Fatalf("FormatFen = %q, want 1500.50", got)
	}
}
