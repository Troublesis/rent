package service

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseIntegerYuanToFen(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("金额不能为空")
	}
	if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "+") {
		return 0, fmt.Errorf("金额不能为负数")
	}
	if strings.Contains(trimmed, ".") {
		return 0, fmt.Errorf("金额需为整数")
	}
	yuan, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("金额需为整数")
	}
	return yuan * 100, nil
}

func ParseYuanToFen(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("金额不能为空")
	}
	if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "+") {
		return 0, fmt.Errorf("金额不能为负数")
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("金额格式不正确")
	}

	yuan, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("金额格式不正确")
	}

	fen := 0
	if len(parts) == 2 {
		fraction := parts[1]
		if len(fraction) > 2 {
			return 0, fmt.Errorf("金额最多支持两位小数")
		}
		if fraction == "" {
			fraction = "0"
		}
		for len(fraction) < 2 {
			fraction += "0"
		}
		fen, err = strconv.Atoi(fraction)
		if err != nil {
			return 0, fmt.Errorf("金额格式不正确")
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
	return fmt.Sprintf("%s%s.%02d", sign, withThousandSeparators(value/100), value%100)
}

func FormatFenAsYuanInt(fen int) string {
	return strconv.Itoa(fen / 100)
}

// FormatFenAsYuanIntDisplay formats a fen amount as an integer yuan string with
// thousand separators for display purposes (e.g. 150000 -> "1,500"). For form
// input field values use FormatFenAsYuanInt instead, which returns a raw
// integer that <input type="number"> can accept.
func FormatFenAsYuanIntDisplay(fen int) string {
	sign := ""
	value := fen
	if value < 0 {
		sign = "-"
		value = -value
	}
	return sign + withThousandSeparators(value/100)
}

// withThousandSeparators returns the absolute integer formatted with ASCII
// comma thousand separators, e.g. 1500 -> "1,500".
func withThousandSeparators(value int) string {
	if value < 0 {
		value = -value
	}
	raw := strconv.Itoa(value)
	if len(raw) <= 3 {
		return raw
	}
	var b strings.Builder
	remainder := len(raw) % 3
	if remainder > 0 {
		b.WriteString(raw[:remainder])
		if len(raw) > remainder {
			b.WriteByte(',')
		}
	}
	for i := remainder; i < len(raw); i += 3 {
		b.WriteString(raw[i : i+3])
		if i+3 < len(raw) {
			b.WriteByte(',')
		}
	}
	return b.String()
}
