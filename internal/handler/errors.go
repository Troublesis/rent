package handler

import (
	"errors"
	"strings"

	"github.com/troublesis/rent/internal/repository"
)

func TranslateDBError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "UNIQUE constraint failed: rooms.room_no"):
		return "该房间号已存在，请使用其他编号"
	case strings.Contains(msg, "UNIQUE constraint failed: tenants.phone"):
		return "该手机号已绑定其他租客"
	case strings.Contains(msg, "NOT NULL constraint failed"):
		return "存在必填项未填写，请检查表单"
	case strings.Contains(msg, "FOREIGN KEY constraint failed"):
		return "关联数据不存在，操作无法完成"
	default:
		return "操作失败，请稍后重试"
	}
}

func userFacingError(err error) string {
	if err == nil {
		return ""
	}
	if repository.IsNotFound(err) {
		return "数据不存在或已被删除"
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "操作失败，请稍后重试"
	}
	if containsChinese(message) && !containsRawDatabaseError(message) {
		return message
	}
	return TranslateDBError(errors.New(message))
}

func containsChinese(value string) bool {
	for _, r := range value {
		if r >= '一' && r <= '鿿' {
			return true
		}
	}
	return false
}

func containsRawDatabaseError(value string) bool {
	indicators := []string{"constraint failed", "UNIQUE constraint", "NOT NULL constraint", "FOREIGN KEY constraint", "SQL", "gorm", "sqlite"}
	for _, indicator := range indicators {
		if strings.Contains(value, indicator) {
			return true
		}
	}
	return false
}
