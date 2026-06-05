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

func (s *PushService) resolveToken(config *model.PushConfig) string {
	if strings.TrimSpace(config.PushPlusToken) != "" {
		return strings.TrimSpace(config.PushPlusToken)
	}
	return strings.TrimSpace(s.cfg.PushPlusToken)
}

type pushItem struct {
	Name        string
	Phone       string
	RoomNo      string
	RoomTitle   string
	Amount      string
	Type        string
	PayDate     string
	MonthlyRent string
	CheckinDate string
	OverdueDays int
}

type pushData struct {
	Items         []pushItem
	TotalAmount   string
	TotalCount    int
	LandlordName  string
	LandlordPhone string
}

func (s *PushService) buildPushData(payments []model.Payment, settings Settings) pushData {
	data := pushData{
		Items:         make([]pushItem, 0, len(payments)),
		TotalAmount:   "0.00",
		LandlordName:  settings.LandlordName,
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
			Type:        paymentTypeLabelCN(p.Type),
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

// renderIndividual renders one message per payment item.
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

// renderSummary finds the {% for item in items %}...{% endfor %} block and loops over items.
func (s *PushService) renderSummary(titleTpl, bodyTpl string, data pushData) ([]pushMessage, error) {
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
		body = replaceGlobalVars(bodyTpl, data)
	}
	body = replaceSummaryVars(body, data)

	return []pushMessage{{Title: title, Content: body}}, nil
}

// extractForBlock pulls content between {% for item in items %} and {% endfor %}
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

// Variable replacement helpers
func replaceItemVars(s string, item pushItem, data pushData) string {
	return strings.NewReplacer(
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
	).Replace(s)
}

func replaceGlobalVars(s string, data pushData) string {
	return strings.NewReplacer(
		"{{房东姓名}}", data.LandlordName,
		"{{房东电话}}", data.LandlordPhone,
	).Replace(s)
}

func replaceSummaryVars(s string, data pushData) string {
	return strings.NewReplacer(
		"{{待收总额}}", data.TotalAmount,
		"{{待收笔数}}", fmt.Sprintf("%d", data.TotalCount),
	).Replace(s)
}

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

// SendReminder scans due payments and sends push notification(s)
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

	payments, err := s.queryDuePayments()
	if err != nil {
		return fmt.Errorf("query due payments: %w", err)
	}
	if len(payments) == 0 {
		return nil
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

// SendTest sends a test push with sample data
func (s *PushService) SendTest(token, title, body string) error {
	if strings.TrimSpace(token) == "" {
		token = s.cfg.PushPlusToken
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("Token 未设置")
	}

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
		Items:         []pushItem{sampleItem},
		TotalAmount:   "¥1,500.00",
		TotalCount:    1,
		LandlordName:  "房东",
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

func formatDateCN(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%d年%d月%d日", t.Year(), int(t.Month()), t.Day())
}

func paymentTypeLabelCN(paymentType string) string {
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

func boolPtr(b bool) *bool { return &b }

func buildPaymentIDsJSON(payments []model.Payment) string {
	ids := make([]string, len(payments))
	for i, p := range payments {
		ids[i] = fmt.Sprintf("%d", p.ID)
	}
	return "[" + strings.Join(ids, ",") + "]"
}
