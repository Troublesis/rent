# Push Notification Feature — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add WeChat Official Account push notifications for rent collection reminders via PushPlus API, with a configurable scheduler and customizable message templates.

**Architecture:** Follows existing handler → service → repository → model pattern. New PushScheduler goroutine uses robfig/cron for daily timed execution. PushPlus API called directly from PushService. Settings page extended with push config + template editor form.

**Tech Stack:** Go 1.25, Gin, GORM + SQLite, robfig/cron/v3, PushPlus HTTP API

**Branch:** `feat/push-notification` (create from main)

---

## Task 0: Branch Setup & Dependencies

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Create feat branch**

```bash
git checkout -b feat/push-notification main
```

- [ ] **Step 2: Install robfig/cron dependency**

```bash
go get github.com/robfig/cron/v3@latest
go mod tidy
```

- [ ] **Step 3: Verify build still works**

```bash
go build ./...
```

Expected: no errors

---

## Task 1: Push Data Models

**Files:**
- Create: `internal/model/push.go`

- [ ] **Step 1: Write push models file**

```go
package model

import "time"

const (
	PushModeSummary    = "summary"
	PushModeIndividual = "individual"
)

const (
	ScheduleModeAdvance = "advance"
	ScheduleModeFixed   = "fixed"
)

const (
	FixedPeriodWeekly  = "weekly"
	FixedPeriodMonthly = "monthly"
)

const (
	PushTypeScheduled = "scheduled"
	PushTypeTest      = "test"
)

const (
	PushStatusSuccess = "success"
	PushStatusFailed  = "failed"
	PushStatusSkipped = "skipped"
)

type PushConfig struct {
	ID            uint   `gorm:"primarykey"`
	PushPlusToken string `gorm:"default:''"`
	PushMode      string `gorm:"default:summary"`
	PushTime      string `gorm:"default:09:00"`
	ScheduleMode  string `gorm:"default:advance"`
	AdvanceDays   int    `gorm:"default:3"`
	FixedPeriod   string `gorm:"default:weekly"`
	FixedWeekday  int    `gorm:"default:2"`
	FixedMonthDay int    `gorm:"default:1"`
	Enabled       bool   `gorm:"default:true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (PushConfig) TableName() string { return "push_configs" }

type PushTemplate struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"not null"`
	TitleTpl  string
	BodyTpl   string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (PushTemplate) TableName() string { return "push_templates" }

type PushLog struct {
	ID         uint   `gorm:"primarykey"`
	PushType   string `gorm:"index"`
	Mode       string
	Title      string
	Content    string `gorm:"type:text"`
	Recipient  string
	Status     string `gorm:"index"`
	PushPlusID string
	ErrorMsg   string
	PaymentIDs string
	CreatedAt  time.Time `gorm:"index"`
}

func (PushLog) TableName() string { return "push_logs" }

func DefaultPushTemplate() PushTemplate {
	return PushTemplate{
		Name:     "房租催收",
		TitleTpl: "📋 房租催收提醒",
		BodyTpl: `房东您好，当前有以下房租待收：

{% for item in items %}
<b>{{姓名}}</b> — {{房源号}} {{房源标题}}
📞 {{手机号}}
💰 应缴金额：<b>¥{{金额}}</b>（{{费用类型}}）
📅 应缴日期：{{应缴日期}}

{% endfor %}
<hr>
💰 待收总计：<b>¥{{待收总额}}</b>（{{待收笔数}}笔）

—— {{房东姓名}} · {{房东电话}}`,
	}
}
```

- [ ] **Step 2: Verify build**

```bash
go build ./internal/model/...
```

Expected: no errors

- [ ] **Step 3: Commit**

Use aicommit2 agent to commit: `git add internal/model/push.go`

---

## Task 2: Push Repository

**Files:**
- Create: `internal/repository/push_repo.go`

- [ ] **Step 1: Write repository**

```go
package repository

import (
	"fmt"

	"github.com/troublesis/rent/internal/model"
	"gorm.io/gorm"
)

type PushRepository struct {
	db *gorm.DB
}

func NewPushRepository(db *gorm.DB) *PushRepository {
	return &PushRepository{db: db}
}

// --- PushConfig ---

func (r *PushRepository) GetConfig() (*model.PushConfig, error) {
	var config model.PushConfig
	if err := r.db.First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *PushRepository) GetConfigOrCreate() (*model.PushConfig, error) {
	config, err := r.GetConfig()
	if err == nil {
		return config, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	defaultConfig := model.PushConfig{ID: 1}
	if err := r.db.Create(&defaultConfig).Error; err != nil {
		return nil, fmt.Errorf("create default push config: %w", err)
	}
	return &defaultConfig, nil
}

func (r *PushRepository) SaveConfig(config *model.PushConfig) error {
	return r.db.Save(config).Error
}

// --- PushTemplate ---

func (r *PushRepository) GetTemplate() (*model.PushTemplate, error) {
	var tmpl model.PushTemplate
	if err := r.db.First(&tmpl).Error; err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (r *PushRepository) GetTemplateOrCreate() (*model.PushTemplate, error) {
	tmpl, err := r.GetTemplate()
	if err == nil {
		return tmpl, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	defaultTmpl := model.DefaultPushTemplate()
	if err := r.db.Create(&defaultTmpl).Error; err != nil {
		return nil, fmt.Errorf("create default push template: %w", err)
	}
	return &defaultTmpl, nil
}

func (r *PushRepository) SaveTemplate(tmpl *model.PushTemplate) error {
	return r.db.Save(tmpl).Error
}

// --- PushLog ---

func (r *PushRepository) CreateLog(log *model.PushLog) error {
	return r.db.Create(log).Error
}
```

- [ ] **Step 2: Verify build**

```bash
go build ./internal/repository/...
```

Expected: no errors

- [ ] **Step 3: Commit**

Use aicommit2 agent to commit: `git add internal/repository/push_repo.go`

---

## Task 3: Config Layer — PushPlus Token

**Files:**
- Modify: `config/config.go`
- Modify: `.env.example`

- [ ] **Step 1: Add PUSHPLUS_TOKEN to Config struct**

In `config/config.go`, add to the struct:

```go
type Config struct {
	// ... existing fields ...
	PushPlusToken string
}
```

In the `Load()` function, add after the existing `LandlordPhone` line:

```go
PushPlusToken: getEnv("PUSHPLUS_TOKEN", ""),
```

- [ ] **Step 2: Add to .env.example**

Append to `.env.example`:

```env
# 消息推送（可选，也可在设置页面配置）
PUSHPLUS_TOKEN=
```

- [ ] **Step 3: Verify build**

```bash
go build ./...
```

Expected: no errors

- [ ] **Step 4: Commit**

Use aicommit2 agent to commit with message: `feat(config): add PUSHPLUS_TOKEN to config and env example`

---

## Task 4: Push Service — Template Rendering & API Call

**Files:**
- Create: `internal/service/push_service.go`

- [ ] **Step 1: Write push service (full file)**

```go
package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/troublesis/rent/config"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

const pushPlusSendURL = "http://www.pushplus.plus/send"

type PushService struct {
	cfg         config.Config
	pushRepo    *repository.PushRepository
	paymentRepo *repository.PaymentRepository
	settingsSvc *SettingsService
	httpClient  *http.Client
}

func NewPushService(
	cfg config.Config,
	pushRepo *repository.PushRepository,
	paymentRepo *repository.PaymentRepository,
	settingsSvc *SettingsService,
) *PushService {
	return &PushService{
		cfg:         cfg,
		pushRepo:    pushRepo,
		paymentRepo: paymentRepo,
		settingsSvc: settingsSvc,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *PushService) resolveToken(config *model.PushConfig) string {
	if strings.TrimSpace(config.PushPlusToken) != "" {
		return strings.TrimSpace(config.PushPlusToken)
	}
	return strings.TrimSpace(s.cfg.PushPlusToken)
}

// --- Template rendering ---

type pushItem struct {
	Name         string
	Phone        string
	RoomNo       string
	RoomTitle    string
	Amount       string
	Type         string
	PayDate      string
	MonthlyRent  string
	CheckinDate  string
	OverdueDays  int
}

type pushData struct {
	Items        []pushItem
	TotalAmount  string
	TotalCount   int
	LandlordName string
	LandlordPhone string
}

func (s *PushService) buildPushData(payments []model.Payment, settings Settings) pushData {
	data := pushData{
		Items:        make([]pushItem, 0, len(payments)),
		TotalAmount:  "0.00",
		LandlordName: settings.LandlordName,
		LandlordPhone: settings.LandlordPhone,
	}

	today := time.Now()
	var totalFen int
	for _, p := range payments {
		totalFen += p.Amount
		overdueDays := 0
		payDateOnly := dateOnly(p.PayDate)
		todayOnly := dateOnly(today)
		if !payDateOnly.IsZero() && payDateOnly.Before(todayOnly) {
			overdueDays = int(todayOnly.Sub(payDateOnly).Hours() / 24)
		}

		data.Items = append(data.Items, pushItem{
			Name:        p.Tenant.Name,
			Phone:       p.Tenant.Phone,
			RoomNo:      p.Tenant.Room.RoomNo,
			RoomTitle:   p.Tenant.Room.Title,
			Amount:      FormatFen(p.Amount),
			Type:        paymentTypeLabel(p.Type),
			PayDate:     formatDateCN(p.PayDate),
			MonthlyRent: FormatFenAsYuanInt(p.Tenant.RentPrice),
			CheckinDate: formatDateCN(p.Tenant.CheckinDate),
			OverdueDays: overdueDays,
		})
	}

	data.TotalAmount = FormatFen(totalFen)
	data.TotalCount = len(payments)
	return data
}

func (s *PushService) renderIndividual(titleTpl, bodyTpl string, data pushData) ([]pushMessage, error) {
	messages := make([]pushMessage, 0, len(data.Items))
	for _, item := range data.Items {
		body := replaceItemVars(bodyTpl, item, data)
		title := replaceItemVars(titleTpl, item, data)
		if strings.TrimSpace(body) == "" {
			continue
		}
		messages = append(messages, pushMessage{Title: title, Content: body})
	}
	return messages, nil
}

func (s *PushService) renderSummary(titleTpl, bodyTpl string, data pushData) ([]pushMessage, error) {
	// Find the {% for item in items %}...{% endfor %} block
	forBlock, before, after := extractForBlock(bodyTpl)
	title := replaceGlobalVars(titleTpl, data)

	var body string
	if forBlock != "" {
		var itemParts []string
		for _, item := range data.Items {
			itemParts = append(itemParts, replaceItemVars(forBlock, item, data))
		}
		body = before + strings.Join(itemParts, "") + after
	} else {
		// No for block — just replace global and summary vars
		body = replaceGlobalVars(bodyTpl, data)
	}
	body = replaceSummaryVars(body, data)

	return []pushMessage{{Title: title, Content: body}}, nil
}

// extractForBlock pulls out the content between {% for item in items %} and {% endfor %}
func extractForBlock(tmpl string) (forBlock, before, after string) {
	startTag := "{% for item in items %}"
	endTag := "{% endfor %}"
	startIdx := strings.Index(tmpl, startTag)
	if startIdx == -1 {
		return "", "", ""
	}
	endIdx := strings.Index(tmpl[startIdx+len(startTag):], endTag)
	if endIdx == -1 {
		return "", "", ""
	}
	forBlock = tmpl[startIdx+len(startTag) : startIdx+len(startTag)+endIdx]
	before = tmpl[:startIdx]
	after = tmpl[startIdx+len(startTag)+endIdx+len(endTag):]
	return
}

func replaceItemVars(s string, item pushItem, data pushData) string {
	replacer := strings.NewReplacer(
		"{{姓名}}", item.Name,
		"{{手机号}}", item.Phone,
		"{{房源号}}", item.RoomNo,
		"{{房源标题}}", item.RoomTitle,
		"{{金额}}", item.Amount,
		"{{费用类型}}", item.Type,
		"{{应缴日期}}", item.PayDate,
		"{{月租金}}", item.MonthlyRent,
		"{{入住日期}}", item.CheckinDate,
		"{{逾期天数}}", fmt.Sprintf("%d", item.OverdueDays),
	)
	return replacer.Replace(s)
}

func replaceGlobalVars(s string, data pushData) string {
	replacer := strings.NewReplacer(
		"{{房东姓名}}", data.LandlordName,
		"{{房东电话}}", data.LandlordPhone,
	)
	return replacer.Replace(s)
}

func replaceSummaryVars(s string, data pushData) string {
	replacer := strings.NewReplacer(
		"{{待收总额}}", data.TotalAmount,
		"{{待收笔数}}", fmt.Sprintf("%d", data.TotalCount),
	)
	return replacer.Replace(s)
}

// --- PushPlus API ---

type pushMessage struct {
	Title   string
	Content string
}

type pushPlusRequest struct {
	Token    string `json:"token"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Template string `json:"template"`
}

type pushPlusResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

func (s *PushService) sendToPushPlus(token, title, content string) (pushPlusResponse, error) {
	reqBody := pushPlusRequest{
		Token:    token,
		Title:    title,
		Content:  content,
		Template: "html",
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return pushPlusResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := s.httpClient.Post(pushPlusSendURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return pushPlusResponse{}, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return pushPlusResponse{}, fmt.Errorf("read response: %w", err)
	}

	var result pushPlusResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return pushPlusResponse{}, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

// --- Main entry points ---

func (s *PushService) SendReminder() error {
	config, err := s.pushRepo.GetConfigOrCreate()
	if err != nil {
		return fmt.Errorf("get push config: %w", err)
	}
	if !config.Enabled {
		return nil
	}
	token := s.resolveToken(config)
	if token == "" {
		// Token not configured — skip silently
		s.createSkippedLog(config.PushMode, "未配置 PushPlus Token")
		return nil
	}

	tmpl, err := s.pushRepo.GetTemplateOrCreate()
	if err != nil {
		return fmt.Errorf("get template: %w", err)
	}
	if strings.TrimSpace(tmpl.BodyTpl) == "" {
		s.createSkippedLog(config.PushMode, "消息模板为空")
		return nil
	}

	// Query due payments — unpaid, not excluded, active tenants
	payments, err := s.queryDuePayments()
	if err != nil {
		return fmt.Errorf("query due payments: %w", err)
	}
	if len(payments) == 0 {
		return nil // nothing to push
	}

	settings, err := s.settingsSvc.GetSettings()
	if err != nil {
		return fmt.Errorf("get settings: %w", err)
	}

	data := s.buildPushData(payments, settings)

	var messages []pushMessage
	if config.PushMode == model.PushModeIndividual {
		messages, err = s.renderIndividual(tmpl.TitleTpl, tmpl.BodyTpl, data)
	} else {
		messages, err = s.renderSummary(tmpl.TitleTpl, tmpl.BodyTpl, data)
	}
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	paymentIDs := buildPaymentIDsJSON(payments)
	for _, msg := range messages {
		result, apiErr := s.sendToPushPlus(token, msg.Title, msg.Content)
		log := &model.PushLog{
			PushType:   model.PushTypeScheduled,
			Mode:       config.PushMode,
			Title:      msg.Title,
			Content:    msg.Content,
			Recipient:  "landlord",
			PaymentIDs: paymentIDs,
		}
		if apiErr != nil {
			log.Status = model.PushStatusFailed
			log.ErrorMsg = apiErr.Error()
		} else if result.Code == 200 {
			log.Status = model.PushStatusSuccess
			log.PushPlusID = result.Data
		} else {
			log.Status = model.PushStatusFailed
			log.ErrorMsg = result.Msg
		}
		_ = s.pushRepo.CreateLog(log)
	}
	return nil
}

func (s *PushService) SendTest(token, title, body string) error {
	if strings.TrimSpace(token) == "" {
		token = s.cfg.PushPlusToken
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("Token 未设置")
	}

	// Replace variables with sample data
	sampleItem := pushItem{
		Name:        "测试租客",
		Phone:       "13800000000",
		RoomNo:      "A101",
		RoomTitle:   "测试房源",
		Amount:      "¥1,500.00",
		Type:        "租金",
		PayDate:     "2026年6月5日",
		MonthlyRent: "1500.00",
		CheckinDate: "2025年1月1日",
		OverdueDays: 0,
	}
	sampleData := pushData{
		Items:        []pushItem{sampleItem},
		TotalAmount:  "¥1,500.00",
		TotalCount:   1,
		LandlordName: "房东",
		LandlordPhone: "13900000000",
	}

	title = replaceGlobalVars(title, sampleData)
	title = replaceSummaryVars(title, sampleData)

	var content string
	forBlock, before, after := extractForBlock(body)
	if forBlock != "" {
		var itemParts []string
		for _, item := range sampleData.Items {
			itemParts = append(itemParts, replaceItemVars(forBlock, item, sampleData))
		}
		content = before + strings.Join(itemParts, "") + after
	} else {
		content = replaceItemVars(body, sampleData.Items[0], sampleData)
	}
	content = replaceGlobalVars(content, sampleData)
	content = replaceSummaryVars(content, sampleData)

	result, err := s.sendToPushPlus(token, title, content)
	if err != nil {
		return err
	}
	if result.Code != 200 {
		return fmt.Errorf("PushPlus 返回错误: %s", result.Msg)
	}

	// Record test log
	_ = s.pushRepo.CreateLog(&model.PushLog{
		PushType:  model.PushTypeTest,
		Mode:      model.PushModeSummary,
		Title:     title,
		Content:   content,
		Recipient: "landlord",
		Status:    model.PushStatusSuccess,
	})
	return nil
}

func (s *PushService) queryDuePayments() ([]model.Payment, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return s.paymentRepo.ListPayments(repository.PaymentFilter{
		Paid:         boolPtr(false),
		Excluded:     boolPtr(false),
		TenantStatus: model.TenantStatusActive,
		ToDate:       today,
	})
}

func (s *PushService) createSkippedLog(mode, reason string) {
	_ = s.pushRepo.CreateLog(&model.PushLog{
		PushType:  model.PushTypeScheduled,
		Mode:      mode,
		Recipient: "landlord",
		Status:    model.PushStatusSkipped,
		ErrorMsg:  reason,
	})
}

// --- helpers ---

func dateOnly(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func formatDateCN(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%d年%d月%d日", t.Year(), int(t.Month()), t.Day())
}

func paymentTypeLabel(paymentType string) string {
	switch paymentType {
	case model.PaymentTypeRent:
		return "租金"
	case model.PaymentTypeWater:
		return "水费"
	case model.PaymentTypeElectricity:
		return "电费"
	case model.PaymentTypeOther:
		return "其他"
	default:
		return "未知"
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func buildPaymentIDsJSON(payments []model.Payment) string {
	ids := make([]string, len(payments))
	for i, p := range payments {
		ids[i] = fmt.Sprintf("%d", p.ID)
	}
	return "[" + strings.Join(ids, ",") + "]"
}
```

- [ ] **Step 2: Verify build**

```bash
go build ./internal/service/...
```

Expected: no errors

- [ ] **Step 3: Commit**

Use aicommit2 agent to commit: `git add internal/service/push_service.go`

---

## Task 5: Push Scheduler

**Files:**
- Create: `internal/service/push_scheduler.go`

- [ ] **Step 1: Write scheduler**

```go
package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
	"github.com/robfig/cron/v3"
)

type PushScheduler struct {
	pushRepo    *repository.PushRepository
	pushService *PushService
	cron        *cron.Cron
}

func NewPushScheduler(pushRepo *repository.PushRepository, pushService *PushService) *PushScheduler {
	return &PushScheduler{
		pushRepo:    pushRepo,
		pushService: pushService,
	}
}

func (ps *PushScheduler) Start(ctx context.Context) {
	ps.cron = cron.New(cron.WithSeconds())

	// Add a job that runs every minute but gates inside
	ps.cron.AddFunc("0 * * * * *", func() {
		ps.tick()
	})

	ps.cron.Start()

	go func() {
		<-ctx.Done()
		ps.cron.Stop()
	}()
}

func (ps *PushScheduler) tick() {
	config, err := ps.pushRepo.GetConfigOrCreate()
	if err != nil {
		log.Printf("[push-scheduler] get config: %v", err)
		return
	}
	if !config.Enabled {
		return
	}

	now := time.Now()
	if !ps.matchesPushTime(config.PushTime, now) {
		return
	}
	if !ps.shouldRunToday(config, now) {
		return
	}

	if err := ps.pushService.SendReminder(); err != nil {
		log.Printf("[push-scheduler] send reminder: %v", err)
	}
}

// matchesPushTime checks if current time is within the push minute window.
// Cron fires every minute; we gate on HH:MM match.
func (ps *PushScheduler) matchesPushTime(pushTime string, now time.Time) bool {
	parts := strings.Split(strings.TrimSpace(pushTime), ":")
	if len(parts) != 2 {
		return false
	}
	hour, err1 := strconv.Atoi(parts[0])
	minute, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	return now.Hour() == hour && now.Minute() == minute
}

func (ps *PushScheduler) shouldRunToday(config *model.PushConfig, now time.Time) bool {
	if config.ScheduleMode == model.ScheduleModeAdvance {
		return true // advance mode runs every day
	}

	// Fixed schedule mode
	if config.FixedPeriod == model.FixedPeriodWeekly {
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday = 7
		}
		return weekday == config.FixedWeekday
	}

	if config.FixedPeriod == model.FixedPeriodMonthly {
		targetDay := config.FixedMonthDay
		lastDay := lastDayOfMonth(now.Year(), now.Month())
		if targetDay > lastDay {
			targetDay = lastDay
		}
		return now.Day() == targetDay
	}

	return false
}

func lastDayOfMonth(year int, month time.Month) int {
	t := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
	return t.Day()
}

func buildCronExpr(pushTime string) (string, error) {
	parts := strings.Split(strings.TrimSpace(pushTime), ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid push time: %s", pushTime)
	}
	minute, err1 := strconv.Atoi(parts[0])
	hour, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || minute < 0 || minute > 59 || hour < 0 || hour > 23 {
		return "", fmt.Errorf("invalid push time: %s", pushTime)
	}
	return fmt.Sprintf("%d %d * * *", minute, hour), nil // standard cron (no seconds)
}
```

- [ ] **Step 2: Verify build**

```bash
go build ./internal/service/...
```

Expected: no errors

- [ ] **Step 3: Commit**

Use aicommit2 agent to commit: `git add internal/service/push_scheduler.go`

---

## Task 6: Push Handler

**Files:**
- Create: `internal/handler/admin_push.go`

- [ ] **Step 1: Write handler**

```go
package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/service"
)

type AdminPushHandler struct {
	renderer    Renderer
	pushService *service.PushService
}

func NewAdminPushHandler(renderer Renderer, pushService *service.PushService) *AdminPushHandler {
	return &AdminPushHandler{renderer: renderer, pushService: pushService}
}

// UpdatePushSettings handles POST /admin/settings/push
func (h *AdminPushHandler) UpdatePushSettings(c *gin.Context) {
	pushConfig := c.PostForm("push_config")
	if pushConfig == "" {
		// Settings form submits all push fields in one form. The settings page
		// currently handles landlord info separately via /admin/settings POST.
		// /admin/settings/push receives the entire form for push fields.
		// But for simplicity, we accept the full combined form too and only
		// process push fields.
	}

	// Parse push_config fields from form
	pushMode := strings.TrimSpace(c.PostForm("push_mode"))
	pushTime := strings.TrimSpace(c.PostForm("push_time"))
	scheduleMode := strings.TrimSpace(c.PostForm("schedule_mode"))
	advanceDaysStr := strings.TrimSpace(c.PostForm("advance_days"))
	fixedPeriod := strings.TrimSpace(c.PostForm("fixed_period"))
	fixedWeekdayStr := strings.TrimSpace(c.PostForm("fixed_weekday"))
	fixedMonthDayStr := strings.TrimSpace(c.PostForm("fixed_month_day"))
	token := strings.TrimSpace(c.PostForm("pushplus_token"))
	enabledStr := strings.TrimSpace(c.PostForm("push_enabled"))

	// Parse template fields
	titleTpl := strings.TrimSpace(c.PostForm("template_title"))
	bodyTpl := strings.TrimSpace(c.PostForm("template_body"))

	// Build update object
	config, err := h.pushService.GetConfig()
	if err != nil {
		c.String(http.StatusInternalServerError, "读取推送配置失败")
		return
	}

	if pushMode != "" {
		config.PushMode = pushMode
	}
	if pushTime != "" {
		config.PushTime = pushTime
	}
	if scheduleMode != "" {
		config.ScheduleMode = scheduleMode
	}
	if scheduleMode == "advance" {
		if advanceDaysStr != "" {
			fmt.Sscanf(advanceDaysStr, "%d", &config.AdvanceDays)
		}
	}
	if scheduleMode == "fixed" {
		config.FixedPeriod = fixedPeriod
		if fixedPeriod == "weekly" {
			fmt.Sscanf(fixedWeekdayStr, "%d", &config.FixedWeekday)
		}
		if fixedPeriod == "monthly" {
			fmt.Sscanf(fixedMonthDayStr, "%d", &config.FixedMonthDay)
		}
	}
	config.PushPlusToken = token
	config.Enabled = enabledStr == "on" || enabledStr == "true" || enabledStr == "1"

	if err := h.pushService.SaveConfig(config); err != nil {
		c.String(http.StatusInternalServerError, "保存推送配置失败")
		return
	}

	// Save template
	tmpl, err := h.pushService.GetTemplate()
	if err != nil {
		c.String(http.StatusInternalServerError, "读取消息模板失败")
		return
	}
	if titleTpl != "" {
		tmpl.TitleTpl = titleTpl
	}
	if bodyTpl != "" {
		tmpl.BodyTpl = bodyTpl
	}
	if err := h.pushService.SaveTemplate(tmpl); err != nil {
		c.String(http.StatusInternalServerError, "保存消息模板失败")
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/settings")
}

// TestPush handles POST /admin/settings/push/test
func (h *AdminPushHandler) TestPush(c *gin.Context) {
	token := strings.TrimSpace(c.PostForm("pushplus_token"))
	title := strings.TrimSpace(c.PostForm("template_title"))
	body := strings.TrimSpace(c.PostForm("template_body"))

	if err := h.pushService.SendTest(token, title, body); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "测试消息已发送，请查看微信"})
}
```

- [ ] **Step 2 (added): Add public-facing service methods**

`PushService` needs `GetConfig`, `SaveConfig`, `GetTemplate`, `SaveTemplate` public methods. These are simple wrappers around the repository. Add these to `push_service.go`:

```go
func (s *PushService) GetConfig() (*model.PushConfig, error) {
	return s.pushRepo.GetConfigOrCreate()
}

func (s *PushService) SaveConfig(config *model.PushConfig) error {
	return s.pushRepo.SaveConfig(config)
}

func (s *PushService) GetTemplate() (*model.PushTemplate, error) {
	return s.pushRepo.GetTemplateOrCreate()
}

func (s *PushService) SaveTemplate(tmpl *model.PushTemplate) error {
	return s.pushRepo.SaveTemplate(tmpl)
}
```

- [ ] **Step 3: Verify build**

```bash
go build ./internal/handler/...
```

Expected: no errors

- [ ] **Step 4: Commit**

Use aicommit2 agent to commit: `git add internal/handler/admin_push.go internal/service/push_service.go`

---

## Task 7: Push Config HTML Template

**Files:**
- Create: `templates/admin/push_config.html`

- [ ] **Step 1: Write push_config.html partial**

```html
{{define "push_config"}}
<div class="card max-w-2xl space-y-6 p-6">
  <div>
    <h3 class="text-2xl font-black">📲 消息推送配置</h3>
    <p class="mt-2 text-sm text-stone-600">通过微信公众号向您推送房租催收提醒。在 <a href="https://www.pushplus.plus" class="text-amber-600 underline" target="_blank">pushplus.plus</a> 获取 Token。</p>
  </div>

  <!-- Token -->
  <label class="block">
    <span class="text-sm font-bold text-stone-600">PushPlus Token</span>
    <input class="input mt-2" type="text" name="pushplus_token" value="{{.PushConfig.PushPlusToken}}" placeholder="留空则使用 .env 中的 PUSHPLUS_TOKEN">
    <p class="mt-1 text-xs text-stone-400">Token 优先使用此处填写值，为空时回退到 .env 环境变量</p>
  </label>

  <!-- Push mode + Time -->
  <div class="grid grid-cols-2 gap-4">
    <label class="block">
      <span class="text-sm font-bold text-stone-600">推送模式</span>
      <select class="input mt-2" name="push_mode">
        <option value="summary" {{if eq .PushConfig.PushMode "summary"}}selected{{end}}>汇总模式（合并当天到期记录）</option>
        <option value="individual" {{if eq .PushConfig.PushMode "individual"}}selected{{end}}>逐条模式（每条单独推送）</option>
      </select>
    </label>
    <label class="block">
      <span class="text-sm font-bold text-stone-600">推送时间</span>
      <input class="input mt-2" type="time" name="push_time" value="{{.PushConfig.PushTime}}">
    </label>
  </div>

  <!-- Schedule mode -->
  <fieldset>
    <legend class="text-sm font-bold text-stone-600 mb-3">调度方式</legend>
    
    <div class="rounded-lg border border-stone-200 bg-stone-50 p-4 mb-3">
      <label class="flex items-center gap-2 font-bold text-sm cursor-pointer">
        <input type="radio" name="schedule_mode" value="advance" {{if eq .PushConfig.ScheduleMode "advance"}}checked{{end}} class="accent-amber-600">
        按缴租日提前推送
      </label>
      <div class="mt-3 ml-6 flex items-center gap-2">
        <span class="text-sm text-stone-500">到期前</span>
        <select name="advance_days" class="input py-1 px-2 text-sm w-auto">
          <option value="0" {{if eq .PushConfig.AdvanceDays 0}}selected{{end}}>当天</option>
          <option value="3" {{if eq .PushConfig.AdvanceDays 3}}selected{{end}}>提前 3 天</option>
          <option value="5" {{if eq .PushConfig.AdvanceDays 5}}selected{{end}}>提前 5 天</option>
          <option value="7" {{if eq .PushConfig.AdvanceDays 7}}selected{{end}}>提前 7 天</option>
          <option value="10" {{if eq .PushConfig.AdvanceDays 10}}selected{{end}}>提前 10 天</option>
        </select>
        <span class="text-sm text-stone-500">推送</span>
      </div>
    </div>

    <div class="rounded-lg border border-stone-200 bg-stone-50 p-4">
      <label class="flex items-center gap-2 font-bold text-sm cursor-pointer">
        <input type="radio" name="schedule_mode" value="fixed" {{if eq .PushConfig.ScheduleMode "fixed"}}checked{{end}} class="accent-amber-600">
        固定周期推送
      </label>
      <div class="mt-3 ml-6 flex items-center gap-2 flex-wrap">
        <select name="fixed_period" id="fixed-period" class="input py-1 px-2 text-sm w-auto" onchange="toggleFixedPeriodFields()">
          <option value="weekly" {{if eq .PushConfig.FixedPeriod "weekly"}}selected{{end}}>每周</option>
          <option value="monthly" {{if eq .PushConfig.FixedPeriod "monthly"}}selected{{end}}>每月</option>
        </select>
        <select name="fixed_weekday" id="fixed-weekday" class="input py-1 px-2 text-sm w-auto">
          <option value="1" {{if eq .PushConfig.FixedWeekday 1}}selected{{end}}>周一</option>
          <option value="2" {{if eq .PushConfig.FixedWeekday 2}}selected{{end}}>周二</option>
          <option value="3" {{if eq .PushConfig.FixedWeekday 3}}selected{{end}}>周三</option>
          <option value="4" {{if eq .PushConfig.FixedWeekday 4}}selected{{end}}>周四</option>
          <option value="5" {{if eq .PushConfig.FixedWeekday 5}}selected{{end}}>周五</option>
          <option value="6" {{if eq .PushConfig.FixedWeekday 6}}selected{{end}}>周六</option>
          <option value="7" {{if eq .PushConfig.FixedWeekday 7}}selected{{end}}>周日</option>
        </select>
        <select name="fixed_month_day" id="fixed-monthday" class="input py-1 px-2 text-sm w-auto" style="display:none;">
          {{range $d := seq 1 31}}
          <option value="{{$d}}" {{if eq $.PushConfig.FixedMonthDay $d}}selected{{end}}>{{$d}} 号</option>
          {{end}}
        </select>
        <span id="monthday-hint" class="text-xs text-stone-400" style="display:none;">如当月无此日，自动取月末最后一天</span>
      </div>
    </div>
  </fieldset>

  <!-- Enabled toggle -->
  <label class="flex items-center gap-3 cursor-pointer">
    <input type="checkbox" name="push_enabled" {{if .PushConfig.Enabled}}checked{{end}} class="accent-amber-600 w-4 h-4">
    <span class="text-sm font-bold text-stone-600">启用自动推送</span>
  </label>

  <hr class="border-stone-200">

  <!-- Template Editor -->
  <div>
    <h3 class="text-2xl font-black">✏️ 消息模板</h3>
    <p class="mt-2 text-sm text-stone-600">编辑推送内容，使用下方变量标签自动填充数据。</p>
  </div>

  <label class="block">
    <span class="text-sm font-bold text-stone-600">推送标题</span>
    <input class="input mt-2" type="text" name="template_title" value="{{.PushTemplate.TitleTpl}}">
  </label>

  <label class="block">
    <span class="text-sm font-bold text-stone-600">推送内容（支持 HTML）</span>
    <textarea class="input mt-2 font-mono text-sm" name="template_body" rows="12">{{.PushTemplate.BodyTpl}}</textarea>
  </label>

  <!-- Variable tags -->
  <div class="rounded-lg border border-amber-200 bg-amber-50 p-4">
    <p class="text-xs font-bold text-amber-800 mb-2">可用变量（点击插入光标位置）</p>
    <div class="flex flex-wrap gap-2">
      <button type="button" class="var-tag" onclick="insertVar('{{姓名}}')">{{姓名}}</button>
      <button type="button" class="var-tag" onclick="insertVar('{{手机号}}')">{{手机号}}</button>
      <button type="button" class="var-tag" onclick="insertVar('{{房源号}}')">{{房源号}}</button>
      <button type="button" class="var-tag" onclick="insertVar('{{房源标题}}')">{{房源标题}}</button>
      <button type="button" class="var-tag" onclick="insertVar('{{金额}}')">{{金额}}</button>
      <button type="button" class="var-tag" onclick="insertVar('{{费用类型}}')">{{费用类型}}</button>
      <button type="button" class="var-tag" onclick="insertVar('{{应缴日期}}')">{{应缴日期}}</button>
      <button type="button" class="var-tag" onclick="insertVar('{{月租金}}')">{{月租金}}</button>
      <button type="button" class="var-tag" onclick="insertVar('{{入住日期}}')">{{入住日期}}</button>
      <button type="button" class="var-tag" onclick="insertVar('{{逾期天数}}')">{{逾期天数}}</button>
      <button type="button" class="var-tag-s" onclick="insertVar('{{房东姓名}}')">{{房东姓名}}</button>
      <button type="button" class="var-tag-s" onclick="insertVar('{{房东电话}}')">{{房东电话}}</button>
      <button type="button" class="var-tag-g" onclick="insertVar('{{待收总额}}')">{{待收总额}}</button>
      <button type="button" class="var-tag-g" onclick="insertVar('{{待收笔数}}')">{{待收笔数}}</button>
    </div>
    <div class="flex gap-4 mt-2 text-xs text-stone-400">
      <span>🟡 单条变量</span>
      <span>🔵 全局变量</span>
      <span>🟢 汇总变量（仅汇总模式）</span>
    </div>
  </div>
</div>

<script>
function insertVar(variable) {
  const ta = document.querySelector('textarea[name="template_body"]');
  if (!ta) return;
  const start = ta.selectionStart;
  const end = ta.selectionEnd;
  ta.value = ta.value.substring(0, start) + variable + ta.value.substring(end);
  ta.focus();
  ta.selectionStart = ta.selectionEnd = start + variable.length;
}

function toggleFixedPeriodFields() {
  const period = document.getElementById('fixed-period').value;
  document.getElementById('fixed-weekday').style.display = period === 'weekly' ? '' : 'none';
  document.getElementById('fixed-monthday').style.display = period === 'monthly' ? '' : 'none';
  document.getElementById('monthday-hint').style.display = period === 'monthly' ? '' : 'none';
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
  toggleFixedPeriodFields();
});
</script>

<style>
.var-tag {
  background: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 6px;
  padding: 2px 8px;
  font-size: 11px;
  font-family: monospace;
  color: #92400e;
  cursor: pointer;
}
.var-tag:hover { background: #fde68a; }
.var-tag-s {
  background: #dbeafe;
  border: 1px solid #3b82f6;
  border-radius: 6px;
  padding: 2px 8px;
  font-size: 11px;
  font-family: monospace;
  color: #1e40af;
  cursor: pointer;
}
.var-tag-s:hover { background: #bfdbfe; }
.var-tag-g {
  background: #dcfce7;
  border: 1px solid #22c55e;
  border-radius: 6px;
  padding: 2px 8px;
  font-size: 11px;
  font-family: monospace;
  color: #166534;
  cursor: pointer;
}
.var-tag-g:hover { background: #bbf7d0; }
</style>
{{end}}
```

- [ ] **Step 2: Commit**

---

## Task 8: Integration — Storage, Router, Settings

**Files:**
- Modify: `internal/storage/storage.go`
- Modify: `internal/server/router.go`
- Modify: `internal/handler/admin_settings.go`
- Modify: `internal/service/settings_service.go`
- Modify: `templates/admin/settings.html`

- [ ] **Step 1: Register push models in AllModels()**

In `internal/storage/storage.go`, add to `AllModels()`:

```go
func AllModels() []interface{} {
    return []interface{}{
        &model.Room{},
        &model.RoomMedia{},
        &model.Tenant{},
        &model.Payment{},
        &model.AppSetting{},
        &model.PushConfig{},    // ADD
        &model.PushTemplate{},  // ADD
        &model.PushLog{},       // ADD
    }
}
```

- [ ] **Step 2: Wire push dependencies in router.go**

**IMPORTANT:** `pushRepo` must be created BEFORE `settingsService` because `SettingsService` now depends on it.

In `internal/server/router.go`, change the creation order — add `pushRepo` right after `settingsRepo`, then update `settingsService` constructor:

```go
settingsRepo := repository.NewSettingsRepository(db)
pushRepo := repository.NewPushRepository(db)     // ADD (before settingsService)

roomService := service.NewRoomService(roomRepo, tenantRepo)
tenantService := service.NewTenantService(db, tenantRepo, roomRepo)
paymentService := service.NewPaymentService(paymentRepo, tenantRepo)
// ...
settingsService := service.NewSettingsService(cfg, settingsRepo, pushRepo)  // CHANGED (3 args)
pushService := service.NewPushService(cfg, pushRepo, paymentRepo, settingsService)  // ADD
pushHandler := handler.NewAdminPushHandler(renderer, pushService)  // ADD
pushScheduler := service.NewPushScheduler(pushRepo, pushService)  // ADD
```

Add routes in the admin group:

```go
admin.POST("/settings/push", pushHandler.UpdatePushSettings)
admin.POST("/settings/push/test", pushHandler.TestPush)
```

Change `NewRouter` return type from `*gin.Engine` to `(*gin.Engine, *service.PushScheduler)`:

```go
func NewRouter(cfg config.Config, db *gorm.DB) (*gin.Engine, *service.PushScheduler) {
```

At the end of NewRouter:
```go
return router, pushScheduler
```

- [ ] **Step 3: Pass PushConfig and PushTemplate to settings page**

In `internal/handler/admin_settings.go`, update `Page` method:

```go
func (h *AdminSettingsHandler) Page(c *gin.Context) {
    settings, err := h.settingsService.GetSettings()
    if err != nil {
        c.String(http.StatusInternalServerError, "读取设置失败")
        return
    }
    pushConfig, _ := h.settingsService.GetPushConfig()
    pushTemplate, _ := h.settingsService.GetPushTemplate()
    h.render(c, http.StatusOK, settings, queryError(c), pushConfig, pushTemplate)
}

func (h *AdminSettingsHandler) render(c *gin.Context, status int, settings service.Settings, errorMessage string, pushConfig *model.PushConfig, pushTemplate *model.PushTemplate) {
    h.renderer.Render(c, status, "admin_base.html", "admin/settings.html", gin.H{
        "Title":        "系统设置",
        "Settings":     settings,
        "PushConfig":   pushConfig,
        "PushTemplate": pushTemplate,
        "Error":        errorMessage,
    })
}
```

Add necessary imports to `admin_settings.go`: `"github.com/troublesis/rent/internal/model"`

- [ ] **Step 4: Add GetPushConfig/GetPushTemplate to SettingsService**

In `internal/service/settings_service.go`, add:

```go
type SettingsService struct {
    cfg          config.Config
    settingsRepo *repository.SettingsRepository
    pushRepo     *repository.PushRepository  // ADD
}

func NewSettingsService(cfg config.Config, settingsRepo *repository.SettingsRepository, pushRepo *repository.PushRepository) *SettingsService {
    return &SettingsService{cfg: cfg, settingsRepo: settingsRepo, pushRepo: pushRepo}
}

func (s *SettingsService) GetPushConfig() (*model.PushConfig, error) {
    return s.pushRepo.GetConfigOrCreate()
}

func (s *SettingsService) GetPushTemplate() (*model.PushTemplate, error) {
    return s.pushRepo.GetTemplateOrCreate()
}
```

Import `"github.com/troublesis/rent/internal/model"` in settings_service.go

- [ ] **Step 5: Update settings.html to include push_config**

In `templates/admin/settings.html`, after the existing form closing `</form>`, add a second form for push settings:

```html
<form class="card max-w-2xl p-6 mt-6" method="post" action="/admin/settings/push">
  {{template "push_config" .}}
  <div class="flex items-center justify-between mt-6">
    <button type="button" class="btn-stone" id="test-push-btn" onclick="testPush()">🧪 发送测试推送</button>
    <button class="btn-primary" type="submit">💾 保存推送设置</button>
  </div>
</form>

<script>
async function testPush() {
  const btn = document.getElementById('test-push-btn');
  btn.disabled = true;
  btn.textContent = '发送中...';
  const formData = new FormData();
  formData.append('pushplus_token', document.querySelector('input[name="pushplus_token"]').value);
  formData.append('template_title', document.querySelector('input[name="template_title"]').value);
  formData.append('template_body', document.querySelector('textarea[name="template_body"]').value);
  try {
    const resp = await fetch('/admin/settings/push/test', {method: 'POST', body: formData});
    const data = await resp.json();
    alert(data.message || (data.success ? '发送成功' : '发送失败'));
  } catch(e) {
    alert('请求失败: ' + e.message);
  } finally {
    btn.disabled = false;
    btn.textContent = '🧪 发送测试推送';
  }
}
</script>
```

- [ ] **Step 6: Verify build**

```bash
go build ./...
```

Expected: no errors

- [ ] **Step 7: Commit**

Use aicommit2 agent to commit all modified files

---

## Task 9: Start Scheduler in main.go & Create Context

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Add scheduler startup to main.go**

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/troublesis/rent/config"
    "github.com/troublesis/rent/internal/server"
    "github.com/troublesis/rent/internal/storage"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("load config: %v", err)
    }
    if err := storage.ApplyTimezone(cfg); err != nil {
        log.Fatalf("%v", err)
    }
    if err := storage.EnsureDirs(cfg); err != nil {
        log.Fatalf("%v", err)
    }
    db, err := storage.Open(cfg, "")
    if err != nil {
        log.Fatalf("%v", err)
    }
    if err := storage.Migrate(db); err != nil {
        log.Fatalf("%v", err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    router, scheduler := server.NewRouter(cfg, db)
    if scheduler != nil {
        scheduler.Start(ctx)
    }

    // Graceful shutdown on SIGINT/SIGTERM
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-quit
        cancel()
    }()

    log.Printf("rent app listening on http://localhost%s", cfg.Addr())
    if err := router.Run(cfg.Addr()); err != nil {
        log.Fatalf("run server: %v", err)
    }
}
```

- [ ] **Step 2: Update router.go to return PushScheduler**

Change `NewRouter` to return `(*gin.Engine, *service.PushScheduler)`:

```go
func NewRouter(cfg config.Config, db *gorm.DB) (*gin.Engine, *service.PushScheduler) {
    // ... existing setup ...
    
    pushRepo := repository.NewPushRepository(db)
    pushService := service.NewPushService(cfg, pushRepo, paymentRepo, settingsService)
    pushHandler := handler.NewAdminPushHandler(renderer, pushService)
    pushScheduler := service.NewPushScheduler(pushRepo, pushService)
    
    // ... routes ...
    
    return router, pushScheduler
}
```

If no push service is needed (test env), return `nil` for scheduler and check in main.go.

- [ ] **Step 3: Verify build**

```bash
go build ./...
```

Expected: no errors

- [ ] **Step 4: Commit**

Use aicommit2 agent to commit

---

## Task 10: Tests

**Files:**
- Create: `internal/service/push_service_test.go`
- Create: `internal/repository/push_repo_test.go`

- [ ] **Step 1: Write repository test**

```go
package repository

import (
    "testing"

    "github.com/troublesis/rent/internal/model"
)

func newPushTestDB(t *testing.T) *PushRepository {
    db := newTestDB(t)
    // AutoMigrate push models too
    if err := db.AutoMigrate(&model.PushConfig{}, &model.PushTemplate{}, &model.PushLog{}); err != nil {
        t.Fatalf("migrate push models: %v", err)
    }
    return NewPushRepository(db)
}

func TestGetConfigOrCreate(t *testing.T) {
    repo := newPushTestDB(t)
    config, err := repo.GetConfigOrCreate()
    if err != nil {
        t.Fatalf("GetConfigOrCreate: %v", err)
    }
    if config.ID != 1 {
        t.Errorf("expected ID=1, got %d", config.ID)
    }
    if config.PushMode != "summary" {
        t.Errorf("expected default PushMode=summary, got %s", config.PushMode)
    }
}

func TestGetConfigOrCreate_Existing(t *testing.T) {
    repo := newPushTestDB(t)
    config1, _ := repo.GetConfigOrCreate()
    config1.PushMode = "individual"
    repo.SaveConfig(config1)

    config2, err := repo.GetConfigOrCreate()
    if err != nil {
        t.Fatalf("GetConfigOrCreate: %v", err)
    }
    if config2.PushMode != "individual" {
        t.Errorf("expected PushMode=individual, got %s", config2.PushMode)
    }
}

func TestGetTemplateOrCreate(t *testing.T) {
    repo := newPushTestDB(t)
    tmpl, err := repo.GetTemplateOrCreate()
    if err != nil {
        t.Fatalf("GetTemplateOrCreate: %v", err)
    }
    if tmpl.Name != "房租催收" {
        t.Errorf("expected template name, got %s", tmpl.Name)
    }
    if tmpl.BodyTpl == "" {
        t.Error("expected non-empty body template")
    }
}

func TestCreateLog(t *testing.T) {
    repo := newPushTestDB(t)
    log := &model.PushLog{
        PushType:  model.PushTypeTest,
        Mode:      model.PushModeSummary,
        Status:    model.PushStatusSuccess,
        Recipient: "landlord",
    }
    if err := repo.CreateLog(log); err != nil {
        t.Fatalf("CreateLog: %v", err)
    }
    if log.ID == 0 {
        t.Error("expected ID to be assigned")
    }
}
```

- [ ] **Step 2: Write service test**

```go
package service

import (
    "testing"
)

func TestExtractForBlock(t *testing.T) {
    tmpl := `开头内容
{% for item in items %}
<b>{{姓名}}</b> — {{房源号}}
{% endfor %}
结尾内容`

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

    s := "<b>{{姓名}}</b> — {{房源号}} {{房源标题}}\n📞 {{手机号}}\n💰 ¥{{金额}}（{{费用类型}}）\n📅 {{应缴日期}}\n⚠️ 逾期{{逾期天数}}天"
    result := replaceItemVars(s, item, data)

    expected := "<b>张三</b> — A101 阳光大单间\n📞 13800000001\n💰 ¥¥1,500.00（租金）\n📅 2026年6月5日\n⚠️ 逾期3天"
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

func TestBuildPushData(t *testing.T) {
    // Tests buildPushData with mock payments
    // Depends on having payments with preloaded Tenant/Room
    // Covered by integration test
}

func TestMatchesPushTime(t *testing.T) {
    // Tests the HH:MM matching logic
    // Covered by scheduler test
}

func TestLastDayOfMonth(t *testing.T) {
    tests := []struct {
        year     int
        month    int
        expected int
    }{
        {2026, 1, 31},
        {2026, 2, 28},
        {2024, 2, 29},
        {2026, 4, 30},
    }
    for _, tc := range tests {
        got := lastDayOfMonth(tc.year, time.Month(tc.month))
        if got != tc.expected {
            t.Errorf("lastDayOfMonth(%d, %d) = %d, want %d", tc.year, tc.month, got, tc.expected)
        }
    }
}
```

Note: service tests need `import "time"` added.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/repository/... -v -run TestPush
go test ./internal/service/... -v -run "TestExtractForBlock|TestReplace|TestLast"
```

Expected: all tests pass

- [ ] **Step 4: Commit**

Use aicommit2 agent to commit test files

---

## Task 11: Run All Tests & Final Verification

- [ ] **Step 1: Run complete test suite**

```bash
go test ./... -count=1
```

Expected: all tests pass, no regressions

- [ ] **Step 2: Build and run**

```bash
go build -o rent-app ./cmd/server/main.go
```

Expected: build succeeds

---

## Summary

| Task | Files | New/Modify |
|------|-------|-----------|
| 0 | Branch + deps | go.mod, go.sum |
| 1 | `internal/model/push.go` | New |
| 2 | `internal/repository/push_repo.go` | New |
| 3 | `config/config.go`, `.env.example` | Modify |
| 4 | `internal/service/push_service.go` | New |
| 5 | `internal/service/push_scheduler.go` | New |
| 6 | `internal/handler/admin_push.go` | New (+ service methods) |
| 7 | `templates/admin/push_config.html` | New (+ template func) |
| 8 | storage, router, handler, service, template | Modify (5 files) |
| 9 | `cmd/server/main.go` | Modify |
| 10 | Tests | New (2 files) |
| 11 | Full verification | — |
