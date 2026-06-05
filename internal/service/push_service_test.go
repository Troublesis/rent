package service

import (
	"testing"
	"time"
)

func TestExtractForBlock(t *testing.T) {
	tmpl := "开头内容\n{% for item in items %}\n<b>{{姓名}}</b> — {{房源号}}\n{% endfor %}\n结尾内容"
	forBlock, before, after := extractForBlock(tmpl)
	expectedFor := "\n<b>{{姓名}}</b> — {{房源号}}\n"
	if forBlock != expectedFor {
		t.Errorf("forBlock:\ngot  %q\nwant %q", forBlock, expectedFor)
	}
	if before != "开头内容\n" {
		t.Errorf("before: got %q", before)
	}
	if after != "\n结尾内容" {
		t.Errorf("after: got %q", after)
	}
}

func TestExtractForBlock_NoBlock(t *testing.T) {
	tmpl := "无循环模板 {{姓名}} 内容"
	forBlock, _, _ := extractForBlock(tmpl)
	if forBlock != "" {
		t.Errorf("expected empty forBlock, got %q", forBlock)
	}
}

func TestReplaceItemVars(t *testing.T) {
	item := pushItem{
		Name:        "张三",
		Phone:       "13800000001",
		RoomNo:      "A101",
		RoomTitle:   "阳光大单间",
		Amount:      "¥1,500.00",
		Type:        "租金",
		PayDate:     "2026年6月5日",
		MonthlyRent: "1500.00",
		CheckinDate: "2025年1月1日",
		OverdueDays: 3,
	}
	data := pushData{LandlordName: "李四", LandlordPhone: "13900000000"}

	s := "<b>{{姓名}}</b> — {{房源号}} {{房源标题}}\n📞 {{手机号}}\n💰 {{金额}}（{{费用类型}}）\n📅 {{应缴日期}}\n⚠️ 逾期{{逾期天数}}天"
	result := replaceItemVars(s, item, data)
	expected := "<b>张三</b> — A101 阳光大单间\n📞 13800000001\n💰 ¥1,500.00（租金）\n📅 2026年6月5日\n⚠️ 逾期3天"
	if result != expected {
		t.Errorf("replaceItemVars:\ngot  %q\nwant %q", result, expected)
	}
}

func TestReplaceGlobalVars(t *testing.T) {
	data := pushData{LandlordName: "李四", LandlordPhone: "13900000000"}
	s := "—— {{房东姓名}} · {{房东电话}}"
	result := replaceGlobalVars(s, data)
	expected := "—— 李四 · 13900000000"
	if result != expected {
		t.Errorf("replaceGlobalVars:\ngot  %q\nwant %q", result, expected)
	}
}

func TestReplaceSummaryVars(t *testing.T) {
	data := pushData{TotalAmount: "¥4,500.00", TotalCount: 3}
	s := "待收总计：{{待收总额}}（{{待收笔数}}笔）"
	result := replaceSummaryVars(s, data)
	expected := "待收总计：¥4,500.00（3笔）"
	if result != expected {
		t.Errorf("replaceSummaryVars:\ngot  %q\nwant %q", result, expected)
	}
}

func TestLastDayOfMonth(t *testing.T) {
	tests := []struct {
		year     int
		month    time.Month
		expected int
	}{
		{2026, 1, 31},
		{2026, 2, 28},
		{2024, 2, 29},
		{2026, 4, 30},
	}
	for _, tc := range tests {
		got := lastDayOfMonth(tc.year, tc.month)
		if got != tc.expected {
			t.Errorf("lastDayOfMonth(%d, %d) = %d, want %d", tc.year, tc.month, got, tc.expected)
		}
	}
}

func TestFormatDateCN(t *testing.T) {
	// Test zero time
	if s := formatDateCN(time.Time{}); s != "" {
		t.Errorf("expected empty for zero, got %q", s)
	}
	// Test valid time
	t2026 := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	expected := "2026年6月5日"
	if s := formatDateCN(t2026); s != expected {
		t.Errorf("got %q, want %q", s, expected)
	}
}

func TestPaymentTypeLabelCN(t *testing.T) {
	tests := [][2]string{
		{"rent", "租金"},
		{"water", "水费"},
		{"electricity", "电费"},
		{"other", "其他"},
		{"unknown", "未知"},
	}
	for _, tc := range tests {
		got := paymentTypeLabelCN(tc[0])
		if got != tc[1] {
			t.Errorf("paymentTypeLabelCN(%q) = %q, want %q", tc[0], got, tc[1])
		}
	}
}

func TestMatchesPushTime(t *testing.T) {
	ps := &PushScheduler{}
	// Valid time match
	t1 := time.Date(2026, 6, 5, 9, 0, 0, 0, time.Local)
	if !ps.matchesPushTime("09:00", t1) {
		t.Error("expected 09:00 to match 9:00 AM")
	}
	// Valid time mismatch
	t2 := time.Date(2026, 6, 5, 9, 5, 0, 0, time.Local)
	if ps.matchesPushTime("09:00", t2) {
		t.Error("expected 09:00 to NOT match 9:05 AM")
	}
	// Hour mismatch
	t3 := time.Date(2026, 6, 5, 10, 0, 0, 0, time.Local)
	if ps.matchesPushTime("09:00", t3) {
		t.Error("expected 09:00 to NOT match 10:00 AM")
	}
	// Invalid format
	if ps.matchesPushTime("abcd", t1) {
		t.Error("expected invalid push time to not match")
	}
	if ps.matchesPushTime("", t1) {
		t.Error("expected empty push time to not match")
	}
}
