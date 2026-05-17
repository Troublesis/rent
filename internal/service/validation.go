package service

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	namePattern   = regexp.MustCompile(`^[\p{Han}A-Za-z]{2,20}$`)
	phonePattern  = regexp.MustCompile(`^1[3-9]\d{9}$`)
	roomNoPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,10}$`)
)

func validateName(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if !namePattern.MatchString(trimmed) {
		return "", fmt.Errorf("姓名需填写 2-20 个中文或英文字母")
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
	if !roomNoPattern.MatchString(trimmed) {
		return "", fmt.Errorf("房间号需为 1-10 位字母或数字")
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
