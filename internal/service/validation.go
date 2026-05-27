package service

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// namePattern accepts Chinese characters, ASCII letters, digits and spaces.
	// Leading/trailing whitespace is trimmed by validateName before matching;
	// inner runs of spaces are allowed (e.g. "张 三", "John Smith").
	namePattern   = regexp.MustCompile(`^[\p{Han}A-Za-z0-9 ]+$`)
	phonePattern  = regexp.MustCompile(`^1[3-9]\d{9}$`)
	roomNoPattern = regexp.MustCompile(`^[\p{Han}A-Za-z0-9 ]+$`)
)

func validateName(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	length := len([]rune(trimmed))
	if length < 2 || length > 20 || !namePattern.MatchString(trimmed) {
		return "", fmt.Errorf("姓名需为 2-20 位中文、字母、数字或空格")
	}
	return trimmed, nil
}

func validatePhone(value string, required bool, label string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" && !required {
		return "", nil
	}
	if !phonePattern.MatchString(trimmed) {
		return "", fmt.Errorf("%s需填写正确的中国大陆手机号", label)
	}
	return trimmed, nil
}

func validateRoomNo(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	length := len([]rune(trimmed))
	if length < 1 || length > 20 || !roomNoPattern.MatchString(trimmed) {
		return "", fmt.Errorf("房间号需为 1-20 位中文、字母、数字或空格")
	}
	return trimmed, nil
}

func validateNotes(value string, label string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if len([]rune(trimmed)) > 1000 {
		return "", fmt.Errorf("%s不能超过 1000 字", label)
	}
	return trimmed, nil
}

func validateIntegerRange(value int, min int, max int, label string) error {
	if value < min || value > max {
		return fmt.Errorf("%s需为 %d-%d 之间的整数", label, min, max)
	}
	return nil
}
