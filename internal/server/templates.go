package server

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/service"
)

type TemplateRenderer struct {
	root    string
	funcMap template.FuncMap
}

func NewTemplateRenderer(root string) *TemplateRenderer {
	return &TemplateRenderer{root: root, funcMap: template.FuncMap{
		"formatFen":         service.FormatFen,
		"formatYuanInt":     formatYuanInt,
		"divideBy100":       service.FormatFen,
		"formatDate":        formatDate,
		"formatInputDate":   formatInputDate,
		"formatDateTime":    formatDateTime,
		"roomStatusLabel":   roomStatusLabel,
		"tenantStatusLabel": tenantStatusLabel,
		"paymentTypeLabel":  paymentTypeLabel,
		"mediaTypeLabel":    mediaTypeLabel,
		"rentTypeLabel":     rentTypeLabel,
		"rentUnitLabel":     rentUnitLabel,
		"paymentTermsLabel": paymentTermsLabel,
		"roomRentPrice":     model.RoomRentPrice,
		"floorPlanLabel":    floorPlanLabel,
		"isOverdue":         isOverdue,
		"seq":               seq,
	}}
}

func (r *TemplateRenderer) Render(c *gin.Context, status int, layout string, page string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["CurrentPath"] = c.Request.URL.Path
	files := []string{filepath.Join(r.root, "layout", layout), filepath.Join(r.root, page)}
	tmpl, err := template.New(layout).Funcs(r.funcMap).ParseFiles(files...)
	if err != nil {
		c.String(http.StatusInternalServerError, "模板加载失败: %v", err)
		return
	}
	layoutName := strings.TrimSuffix(layout, filepath.Ext(layout))
	c.Status(status)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, layoutName, data); err != nil {
		c.String(http.StatusInternalServerError, "模板渲染失败: %v", err)
	}
}

func formatYuanInt(fen int) int {
	return fen / 100
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.Format("2006-01-02")
}

func formatInputDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func formatDateTime(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.Format("2006-01-02 15:04")
}

func roomStatusLabel(status string) string {
	switch status {
	case model.RoomStatusVacant:
		return "空置"
	case model.RoomStatusOccupied:
		return "已出租"
	case model.RoomStatusMaintenance:
		return "维护中"
	default:
		return "未知"
	}
}

func tenantStatusLabel(status string) string {
	switch status {
	case model.TenantStatusActive:
		return "在租"
	case model.TenantStatusCheckout:
		return "已退租"
	default:
		return "未知"
	}
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

func mediaTypeLabel(mediaType string) string {
	switch mediaType {
	case model.MediaTypeImage:
		return "图片"
	case model.MediaTypeVideo:
		return "视频"
	case model.MediaTypeVideoLink:
		return "视频链接"
	default:
		return "文件"
	}
}

func rentTypeLabel(rentType string) string {
	switch model.RentTypeOrDefault(rentType) {
	case model.RentTypeDaily:
		return "日租"
	default:
		return "月租"
	}
}

func rentUnitLabel(rentType string) string {
	switch model.RentTypeOrDefault(rentType) {
	case model.RentTypeDaily:
		return "天"
	default:
		return "月"
	}
}

func paymentTermsLabel(paymentTerms string) string {
	switch model.PaymentTermsOrDefault(paymentTerms) {
	case model.PaymentTerms1M2D:
		return "付一押二"
	case model.PaymentTerms3M1D:
		return "付三押一"
	case model.PaymentTerms6M0D:
		return "半年付"
	case model.PaymentTerms12M0D:
		return "年付"
	default:
		return "付一押一"
	}
}

func floorPlanLabel(bedrooms int, livingRooms int, bathrooms int) string {
	return fmt.Sprintf("%d室%d厅%d卫", bedrooms, livingRooms, bathrooms)
}

func isOverdue(value time.Time) bool {
	if value.IsZero() {
		return false
	}
	today := time.Now().Truncate(24 * time.Hour)
	return value.Before(today)
}

func seq(start int, end int) []int {
	if end < start {
		return []int{}
	}
	values := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		values = append(values, i)
	}
	return values
}

func yuanJSON(fen int) string {
	return fmt.Sprintf("%s", service.FormatFen(fen))
}
