package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/service"
)

type TemplateRenderer struct {
	root         string
	funcMap      template.FuncMap
	assetVersion string
}

func NewTemplateRenderer(root string) *TemplateRenderer {
	renderer := &TemplateRenderer{
		root:         root,
		assetVersion: computeAssetVersion("static"),
	}
	renderer.funcMap = template.FuncMap{
		"formatFen":          service.FormatFen,
		"formatYuanInt":      formatYuanInt,
		"formatYuanIntRaw":   formatYuanIntRaw,
		"divideBy100":        service.FormatFen,
		"formatDate":         formatDate,
		"formatOptionalDate": formatOptionalDate,
		"formatInputDate":    formatInputDate,
		"formatDateTime":     formatDateTime,
		"roomStatusLabel":    roomStatusLabel,
		"tenantStatusLabel":  tenantStatusLabel,
		"paymentTypeLabel":   paymentTypeLabel,
		"mediaTypeLabel":     mediaTypeLabel,
		"rentTypeLabel":      rentTypeLabel,
		"rentUnitLabel":      rentUnitLabel,
		"paymentTermsLabel":  paymentTermsLabel,
		"isRoomOccupied":     isRoomOccupied,
		"tenantGenderLabel":  tenantGenderLabel,
		"roomRentPrice":      model.RoomRentPrice,
		"floorPlanLabel":     floorPlanLabel,
		"firstImageURL":      firstImageURL,
		"mediaPosterURL":     mediaPosterURL,
		"isPlayableMedia":    isPlayableMedia,
		"isOverdue":          isOverdue,
		"deref":              derefTime,
		"dict":               dictHelper,
		"firstRune":          firstRune,
		"seq":                seq,
		"progressPercent":    progressPercent,
		"subInt":             subInt,
		"asset":              renderer.assetURL,
	}
	return renderer
}

// assetURL appends a cache-busting fingerprint to a static asset path so that
// browsers reload changed JS/CSS without manual hard-refresh. The version is
// computed once at startup from the latest mtime under ./static.
func (r *TemplateRenderer) assetURL(path string) string {
	if r.assetVersion == "" {
		return path
	}
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	return path + separator + "v=" + r.assetVersion
}

// computeAssetVersion walks the static asset directory once and returns a
// short fingerprint derived from the most recent file mtime. Falls back to
// the current time if the directory cannot be read (e.g. tests run from an
// unusual cwd) so that callers still get a non-empty version.
func computeAssetVersion(root string) string {
	var latest time.Time
	err := filepath.WalkDir(root, func(_ string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return nil
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
		return nil
	})
	if err != nil || latest.IsZero() {
		latest = time.Now()
	}
	return strconv.FormatInt(latest.Unix(), 36)
}

func (r *TemplateRenderer) Render(c *gin.Context, status int, layout string, page string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["CurrentPath"] = c.Request.URL.Path
	files, err := r.templateFiles(layout, page)
	if err != nil {
		c.String(http.StatusInternalServerError, "模板加载失败: %v", err)
		return
	}
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

func (r *TemplateRenderer) RenderPartial(c *gin.Context, status int, page string, templateName string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["CurrentPath"] = c.Request.URL.Path
	files, err := r.partialTemplateFiles(page)
	if err != nil {
		c.String(http.StatusInternalServerError, "模板加载失败: %v", err)
		return
	}
	tmpl, err := template.New(templateName).Funcs(r.funcMap).ParseFiles(files...)
	if err != nil {
		c.String(http.StatusInternalServerError, "模板加载失败: %v", err)
		return
	}
	c.Status(status)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, templateName, data); err != nil {
		c.String(http.StatusInternalServerError, "模板渲染失败: %v", err)
	}
}

func (r *TemplateRenderer) templateFiles(layout string, page string) ([]string, error) {
	componentFiles, err := filepath.Glob(filepath.Join(r.root, "components", "*.html"))
	if err != nil {
		return nil, err
	}
	files := []string{filepath.Join(r.root, "layout", layout)}
	files = append(files, componentFiles...)
	files = append(files, filepath.Join(r.root, page))
	return files, nil
}

func (r *TemplateRenderer) partialTemplateFiles(page string) ([]string, error) {
	componentFiles, err := filepath.Glob(filepath.Join(r.root, "components", "*.html"))
	if err != nil {
		return nil, err
	}
	files := append([]string{}, componentFiles...)
	files = append(files, filepath.Join(r.root, page))
	return files, nil
}

// formatYuanInt renders a fen amount as an integer yuan string with thousand
// separators for display contexts (e.g. 150000 -> "1,500"). Do NOT use this in
// <input type="number"> value attributes — browsers will reject commas. Use
// formatYuanIntRaw in form input values instead.
func formatYuanInt(fen int) string {
	return service.FormatFenAsYuanIntDisplay(fen)
}

// formatYuanIntRaw renders a fen amount as a plain integer yuan string without
// thousand separators (e.g. 150000 -> "1500"). Use only for form input values.
func formatYuanIntRaw(fen int) string {
	return service.FormatFenAsYuanInt(fen)
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return fmt.Sprintf("%d年%d月%d日", value.Year(), int(value.Month()), value.Day())
}

func formatOptionalDate(value *time.Time) string {
	if value == nil {
		return "-"
	}
	return formatDate(*value)
}

// formatInputDate keeps ISO format because HTML <input type="date"> requires it.
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
	return fmt.Sprintf("%d年%d月%d日 %02d:%02d", value.Year(), int(value.Month()), value.Day(), value.Hour(), value.Minute())
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

func tenantGenderLabel(gender string) string {
	switch gender {
	case model.TenantGenderMale:
		return "男性"
	case model.TenantGenderFemale:
		return "女性"
	default:
		return "未填写"
	}
}

func isRoomOccupied(status string) bool {
	return status == model.RoomStatusOccupied
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

func firstImageURL(media []model.RoomMedia) string {
	for _, item := range media {
		if item.MediaType == model.MediaTypeImage {
			return item.URL
		}
	}
	return ""
}

func mediaPosterURL(media model.RoomMedia, mediaList []model.RoomMedia) string {
	if media.MediaType == model.MediaTypeImage {
		return media.URL
	}
	if media.MediaType == model.MediaTypeVideoLink {
		return firstImageURL(mediaList)
	}
	return ""
}

func isPlayableMedia(mediaType string) bool {
	return mediaType == model.MediaTypeVideo || mediaType == model.MediaTypeVideoLink
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

func dictHelper(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict expects an even number of arguments")
	}
	result := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict key %v is not a string", values[i])
		}
		result[key] = values[i+1]
	}
	return result, nil
}

func derefTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func isOverdue(value time.Time) bool {
	if value.IsZero() {
		return false
	}
	now := time.Now()
	loc := now.Location()
	localValue := value.In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	valueDate := time.Date(localValue.Year(), localValue.Month(), localValue.Day(), 0, 0, 0, 0, loc)
	return valueDate.Before(today)
}

func firstRune(value string) string {
	for _, item := range value {
		return string(item)
	}
	return "租"
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

func progressPercent(value int, target int) int {
	if target <= 0 || value <= 0 {
		return 0
	}
	pct := value * 100 / target
	if pct > 100 {
		return 100
	}
	if pct < 0 {
		return 0
	}
	return pct
}

func subInt(a int, b int) int {
	result := a - b
	if result < 0 {
		return 0
	}
	return result
}
