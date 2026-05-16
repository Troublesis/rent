package service

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseYuanToFen(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("amount is required")
	}
	if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "+") {
		return 0, fmt.Errorf("amount must be positive")
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("invalid amount format")
	}

	yuan, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid yuan amount")
	}

	fen := 0
	if len(parts) == 2 {
		fraction := parts[1]
		if len(fraction) > 2 {
			return 0, fmt.Errorf("amount supports at most two decimal places")
		}
		if fraction == "" {
			fraction = "0"
		}
		for len(fraction) < 2 {
			fraction += "0"
		}
		fen, err = strconv.Atoi(fraction)
		if err != nil {
			return 0, fmt.Errorf("invalid fen amount")
		}
	}

	return yuan*100 + fen, nil
}

func FormatFen(fen int) string {
	sign := ""
	value := fen
	if value < 0 {
		sign = "-"
		value = -value
	}
	return fmt.Sprintf("%s%d.%02d", sign, value/100, value%100)
}
