// Package seed populates the development database with realistic fixture
// data so that the admin UI can be evaluated against meaningful numbers.
//
// The data set is intentionally deterministic — the same Run call produces
// the same rows every time so that screenshots and tests stay stable.
package seed

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/service"
	"gorm.io/gorm"
)

// Run seeds the database. Callers are expected to have already migrated (or
// reset) the schema. Run is safe to invoke against an empty database; calling
// it on a database that already contains rows simply appends another set.
func Run(db *gorm.DB, now time.Time) (Stats, error) {
	if now.IsZero() {
		now = time.Now()
	}
	rng := rand.New(rand.NewSource(42))
	var stats Stats
	err := db.Transaction(func(tx *gorm.DB) error {
		rooms, err := seedRooms(tx)
		if err != nil {
			return err
		}
		stats.Rooms = len(rooms)

		tenants, err := seedTenants(tx, rooms, now)
		if err != nil {
			return err
		}
		stats.Tenants = len(tenants)

		paymentCount, err := seedPayments(tx, tenants, now, rng)
		if err != nil {
			return err
		}
		stats.Payments = paymentCount

		settingsCount, err := seedSettings(tx)
		if err != nil {
			return err
		}
		stats.Settings = settingsCount
		return nil
	})
	return stats, err
}

// Stats summarises the rows that the seeder inserted.
type Stats struct {
	Rooms    int
	Tenants  int
	Payments int
	Settings int
}

// ----- rooms ----------------------------------------------------------------

type roomSpec struct {
	RoomNo       string
	Title        string
	Description  string
	RentPriceYn  int
	DepositYn    int
	RentType     string
	PaymentTerms string
	Status       string
	Area         int
	Floor        int
	Address      string
	Bedrooms     int
	LivingRooms  int
	Bathrooms    int
	Orientation  string
	Tags         string
	MediaURLs    []string
}

func roomSpecs() []roomSpec {
	// 18 rooms: 11 will be occupied (active tenants), 2 maintenance, 5 vacant.
	return []roomSpec{
		// A 楼层 — 一房一厅
		{"A101", "锦绣里 · 朝南一居", "南向采光，靠近地铁口，适合单身上班族。", 2800, 5000, model.RentTypeMonthly, model.PaymentTerms1M1D, model.RoomStatusOccupied, 32, 1, "锦绣街 18 号", 1, 1, 1, "朝南", "电梯,阳台,空调", []string{"/static/img/seed/a101-1.jpg", "/static/img/seed/a101-2.jpg"}},
		{"A102", "锦绣里 · 北向一居", "安静内庭，适合居家办公。", 2600, 5000, model.RentTypeMonthly, model.PaymentTerms1M1D, model.RoomStatusOccupied, 30, 1, "锦绣街 18 号", 1, 1, 1, "朝北", "空调,洗衣机", nil},
		{"A103", "锦绣里 · 短租日付", "可日租，含家具家电，适合短期居住。", 180, 1000, model.RentTypeDaily, model.PaymentTerms1M1D, model.RoomStatusOccupied, 28, 1, "锦绣街 18 号", 1, 1, 1, "朝东", "短租,日付", nil},
		{"A104", "锦绣里 · 装修升级", "装修升级中，预计两周后开放。", 2900, 5000, model.RentTypeMonthly, model.PaymentTerms1M1D, model.RoomStatusMaintenance, 33, 1, "锦绣街 18 号", 1, 1, 1, "朝南", "翻新中", nil},
		{"A105", "锦绣里 · 待出租", "刚清洁完毕，随时可入住。", 2700, 5000, model.RentTypeMonthly, model.PaymentTerms1M1D, model.RoomStatusVacant, 31, 1, "锦绣街 18 号", 1, 1, 1, "朝东", "电梯,采光佳", nil},
		{"A106", "锦绣里 · 待出租", "近地铁口三百米。", 2750, 5000, model.RentTypeMonthly, model.PaymentTerms1M1D, model.RoomStatusVacant, 32, 1, "锦绣街 18 号", 1, 1, 1, "朝西", "电梯", nil},
		// B 楼层 — 两房一厅
		{"B201", "怡园 · 精装两居", "南北通透，主卧带飘窗。", 4500, 9000, model.RentTypeMonthly, model.PaymentTerms3M1D, model.RoomStatusOccupied, 68, 2, "怡园路 6 号", 2, 1, 1, "南北通透", "电梯,飘窗,空调,洗衣机", []string{"/static/img/seed/b201-1.jpg", "/static/img/seed/b201-2.jpg", "/static/img/seed/b201-3.jpg"}},
		{"B202", "怡园 · 朝南两居", "对面就是公园，环境舒适。", 4300, 9000, model.RentTypeMonthly, model.PaymentTerms1M2D, model.RoomStatusOccupied, 65, 2, "怡园路 6 号", 2, 1, 1, "朝南", "近公园,采光佳", nil},
		{"B203", "怡园 · 经济两居", "整租，月付，押一付一。", 4100, 4100, model.RentTypeMonthly, model.PaymentTerms1M1D, model.RoomStatusOccupied, 62, 2, "怡园路 6 号", 2, 1, 1, "朝东", "月付,押一付一", nil},
		{"B204", "怡园 · 半年付优惠", "半年付直降两百，长租优先。", 4200, 9000, model.RentTypeMonthly, model.PaymentTerms6M0D, model.RoomStatusOccupied, 66, 2, "怡园路 6 号", 2, 1, 1, "朝南", "半年付,优惠", nil},
		{"B205", "怡园 · 待出租", "刚做完保洁，欢迎看房。", 4400, 9000, model.RentTypeMonthly, model.PaymentTerms1M1D, model.RoomStatusVacant, 67, 2, "怡园路 6 号", 2, 1, 1, "朝南", "整租,通风好", nil},
		{"B206", "怡园 · 维修中", "热水器需更换，下周完成。", 4250, 9000, model.RentTypeMonthly, model.PaymentTerms1M1D, model.RoomStatusMaintenance, 64, 2, "怡园路 6 号", 2, 1, 1, "朝东", "维修中", nil},
		// C 楼层 — 三房
		{"C301", "云溪 · 精装三居", "高楼层视野无敌，主卧带卫。", 6800, 13600, model.RentTypeMonthly, model.PaymentTerms3M1D, model.RoomStatusOccupied, 110, 3, "云溪大道 9 号", 3, 2, 2, "南北通透", "电梯,主卧带卫,精装,空调,洗衣机", []string{"/static/img/seed/c301-1.jpg", "/static/img/seed/c301-2.jpg"}},
		{"C302", "云溪 · 年付优惠", "年付直降一千，签约即送清洁。", 6500, 13000, model.RentTypeMonthly, model.PaymentTerms12M0D, model.RoomStatusOccupied, 108, 3, "云溪大道 9 号", 3, 2, 2, "南北通透", "年付,长租优惠", nil},
		{"C303", "云溪 · 朝南三居", "全屋朝南，孩子上学方便。", 6600, 13200, model.RentTypeMonthly, model.PaymentTerms3M1D, model.RoomStatusOccupied, 112, 3, "云溪大道 9 号", 3, 2, 2, "朝南", "学区,采光佳", nil},
		{"C304", "云溪 · 行政套间", "短租为主，配套商务家具。", 320, 5000, model.RentTypeDaily, model.PaymentTerms1M1D, model.RoomStatusOccupied, 95, 3, "云溪大道 9 号", 2, 1, 1, "朝南", "短租,日付,商务", nil},
		{"C305", "云溪 · 待出租", "拎包入住，钥匙在管理处。", 6300, 12600, model.RentTypeMonthly, model.PaymentTerms3M1D, model.RoomStatusVacant, 105, 3, "云溪大道 9 号", 3, 2, 2, "朝南", "整租,拎包入住", nil},
		{"C306", "云溪 · 待出租", "中楼层，性价比高。", 6100, 12200, model.RentTypeMonthly, model.PaymentTerms1M1D, model.RoomStatusVacant, 100, 3, "云溪大道 9 号", 3, 2, 2, "朝东", "整租", nil},
	}
}

func seedRooms(tx *gorm.DB) (map[string]*model.Room, error) {
	specs := roomSpecs()
	rooms := make(map[string]*model.Room, len(specs))
	for _, spec := range specs {
		rentFen, err := service.ParseIntegerYuanToFen(fmt.Sprintf("%d", spec.RentPriceYn))
		if err != nil {
			return nil, fmt.Errorf("room %s rent: %w", spec.RoomNo, err)
		}
		depositFen, err := service.ParseIntegerYuanToFen(fmt.Sprintf("%d", spec.DepositYn))
		if err != nil {
			return nil, fmt.Errorf("room %s deposit: %w", spec.RoomNo, err)
		}
		room := &model.Room{
			RoomNo:       spec.RoomNo,
			Title:        spec.Title,
			Description:  spec.Description,
			Price:        rentFen,
			RentPrice:    rentFen,
			RentType:     spec.RentType,
			PaymentTerms: spec.PaymentTerms,
			Deposit:      depositFen,
			Status:       spec.Status,
			Area:         spec.Area,
			Floor:        spec.Floor,
			Address:      spec.Address,
			Bedrooms:     spec.Bedrooms,
			LivingRooms:  spec.LivingRooms,
			Bathrooms:    spec.Bathrooms,
			Orientation:  spec.Orientation,
			Tags:         spec.Tags,
		}
		if err := tx.Create(room).Error; err != nil {
			return nil, fmt.Errorf("create room %s: %w", spec.RoomNo, err)
		}
		for _, url := range spec.MediaURLs {
			if err := tx.Create(&model.RoomMedia{RoomID: room.ID, URL: url, MediaType: model.MediaTypeImage}).Error; err != nil {
				return nil, fmt.Errorf("create media for %s: %w", spec.RoomNo, err)
			}
		}
		rooms[spec.RoomNo] = room
	}
	return rooms, nil
}

// ----- tenants --------------------------------------------------------------

type tenantSpec struct {
	Name             string
	Phone            string
	EmergencyContact string
	Gender           string
	RoomNo           string
	CheckinMonthsAgo int
	CheckinDayShift  int
	LeaseMonths      int // 0 → no lease end date
	LeaseEndOverride *time.Time
	RentPriceYn      int // overrides room rent price when > 0
	DepositYn        int // overrides room deposit when > 0
	RentType         string
	PaymentTerms     string
	Notes            string
	Status           string
	CheckoutShift    int // days before now when checking out (only for checkout tenants)
}

func tenantSpecs(now time.Time) []tenantSpec {
	// 11 active tenants + 3 checkout tenants. One active tenant has a lease
	// end date in the past to exercise the overdue badge.
	overdueLease := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -45)
	return []tenantSpec{
		// Active — 11 tenants
		{"陈玉婷", "13800001001", "陈父 13900001001", model.TenantGenderFemale, "A101", 14, 0, 24, nil, 0, 0, "", "", "稳定租客，按时缴费。", model.TenantStatusActive, 0},
		{"林子轩", "13800001002", "林妈 13900001002", model.TenantGenderMale, "A102", 9, -3, 12, nil, 0, 0, "", "", "", model.TenantStatusActive, 0},
		{"周婧怡", "13800001003", "周哥 13900001003", model.TenantGenderFemale, "A103", 1, 0, 0, nil, 0, 0, "", "", "短租客，按日计租。", model.TenantStatusActive, 0},
		{"赵伟", "13800001004", "赵姐 13900001004", model.TenantGenderMale, "B201", 18, 0, 36, nil, 0, 0, "", "", "长期租客。", model.TenantStatusActive, 0},
		{"王怀珺", "13800001005", "王父 13900001005", model.TenantGenderFemale, "B202", 7, 2, 12, nil, 0, 0, "", "", "", model.TenantStatusActive, 0},
		{"刘晓东", "13800001006", "刘妻 13900001006", model.TenantGenderMale, "B203", 4, 0, 12, nil, 0, 0, "", "", "新签约。", model.TenantStatusActive, 0},
		{"杨思琪", "13800001007", "杨母 13900001007", model.TenantGenderFemale, "B204", 12, 0, 12, &overdueLease, 0, 0, "", "", "合同已到期，待续签。", model.TenantStatusActive, 0},
		{"郑伟", "13800001008", "郑兄 13900001008", model.TenantGenderMale, "C301", 11, -5, 24, nil, 0, 0, "", "", "", model.TenantStatusActive, 0},
		{"黄丽华", "13800001009", "黄父 13900001009", model.TenantGenderFemale, "C302", 10, 0, 12, nil, 0, 0, "", "", "年付优惠租客。", model.TenantStatusActive, 0},
		{"高俊杰", "13800001010", "高母 13900001010", model.TenantGenderMale, "C303", 6, 4, 24, nil, 0, 0, "", "", "", model.TenantStatusActive, 0},
		{"徐晨曦", "13800001011", "徐父 13900001011", model.TenantGenderMale, "C304", 0, -10, 0, nil, 0, 0, "", "", "行政短租。", model.TenantStatusActive, 0},
		// Checkout — 3 tenants. Rooms remain in the vacant state we picked
		// above (A105/A106/B205/C305/C306 etc are vacant). We reuse one
		// vacant-now room per checkout tenant so foreign keys remain valid.
		{"古亦凡", "13800002001", "古母 13900002001", model.TenantGenderMale, "A105", 15, 0, 12, nil, 0, 0, "", "", "已退租，押金已退。", model.TenantStatusCheckout, 30},
		{"沈晓彤", "13800002002", "沈父 13900002002", model.TenantGenderFemale, "B205", 19, 2, 12, nil, 0, 0, "", "", "搬迁至外地。", model.TenantStatusCheckout, 60},
		{"邓佳琪", "13800002003", "邓哥 13900002003", model.TenantGenderFemale, "C305", 22, -4, 12, nil, 0, 0, "", "", "毕业返乡。", model.TenantStatusCheckout, 90},
	}
}

func seedTenants(tx *gorm.DB, rooms map[string]*model.Room, now time.Time) ([]model.Tenant, error) {
	specs := tenantSpecs(now)
	tenants := make([]model.Tenant, 0, len(specs))
	today := dateOnly(now)
	for _, spec := range specs {
		room, ok := rooms[spec.RoomNo]
		if !ok {
			return nil, fmt.Errorf("tenant %s references unknown room %s", spec.Name, spec.RoomNo)
		}
		rentFen := room.RentPrice
		if spec.RentPriceYn > 0 {
			parsed, err := service.ParseIntegerYuanToFen(fmt.Sprintf("%d", spec.RentPriceYn))
			if err != nil {
				return nil, fmt.Errorf("tenant %s rent: %w", spec.Name, err)
			}
			rentFen = parsed
		}
		depositFen := room.Deposit
		if spec.DepositYn > 0 {
			parsed, err := service.ParseIntegerYuanToFen(fmt.Sprintf("%d", spec.DepositYn))
			if err != nil {
				return nil, fmt.Errorf("tenant %s deposit: %w", spec.Name, err)
			}
			depositFen = parsed
		}
		rentType := spec.RentType
		if rentType == "" {
			rentType = room.RentType
		}
		paymentTerms := spec.PaymentTerms
		if paymentTerms == "" {
			paymentTerms = room.PaymentTerms
		}
		checkin := today.AddDate(0, -spec.CheckinMonthsAgo, spec.CheckinDayShift)
		tenant := model.Tenant{
			Name:             spec.Name,
			Phone:            spec.Phone,
			EmergencyContact: spec.EmergencyContact,
			Gender:           spec.Gender,
			RoomID:           room.ID,
			CheckinDate:      checkin,
			RentPrice:        rentFen,
			RentType:         rentType,
			PaymentTerms:     paymentTerms,
			Deposit:          depositFen,
			Notes:            spec.Notes,
			Status:           spec.Status,
		}
		switch {
		case spec.LeaseEndOverride != nil:
			t := *spec.LeaseEndOverride
			tenant.LeaseEndDate = &t
		case spec.LeaseMonths > 0:
			lease := checkin.AddDate(0, spec.LeaseMonths, 0)
			tenant.LeaseEndDate = &lease
		}
		if spec.Status == model.TenantStatusCheckout {
			checkout := today.AddDate(0, 0, -spec.CheckoutShift)
			tenant.CheckoutDate = &checkout
		}
		if err := tx.Create(&tenant).Error; err != nil {
			return nil, fmt.Errorf("create tenant %s: %w", spec.Name, err)
		}
		// Keep room.status consistent with tenant status. Active tenants
		// flip their room to occupied; checkout tenants leave the room
		// vacant (matching the production check-out flow).
		if spec.Status == model.TenantStatusActive {
			if err := tx.Model(&model.Room{}).Where("id = ?", room.ID).Update("status", model.RoomStatusOccupied).Error; err != nil {
				return nil, fmt.Errorf("flip room %s to occupied: %w", room.RoomNo, err)
			}
			room.Status = model.RoomStatusOccupied
		}
		tenants = append(tenants, tenant)
	}
	return tenants, nil
}

// ----- payments -------------------------------------------------------------

func seedPayments(tx *gorm.DB, tenants []model.Tenant, now time.Time, rng *rand.Rand) (int, error) {
	count := 0
	today := dateOnly(now)
	for _, tenant := range tenants {
		end := today
		if tenant.Status == model.TenantStatusCheckout && tenant.CheckoutDate != nil {
			end = dateOnly(*tenant.CheckoutDate)
		}
		dues := service.RentDueSchedule(tenant, end)
		for _, due := range dues {
			payment := buildRentPayment(tenant, due, today, rng)
			if err := tx.Create(&payment).Error; err != nil {
				return count, fmt.Errorf("seed rent payment tenant=%s date=%s: %w", tenant.Name, due.DueDate.Format("2006-01-02"), err)
			}
			count++
		}
	}

	// Ad-hoc utility payments — water/electricity scattered across the last
	// six months. Distribute roughly 25 entries across the active tenants.
	utilityTenants := make([]model.Tenant, 0, len(tenants))
	for _, t := range tenants {
		if t.Status == model.TenantStatusActive && t.ID != 0 {
			utilityTenants = append(utilityTenants, t)
		}
	}
	if len(utilityTenants) > 0 {
		for i := 0; i < 25; i++ {
			t := utilityTenants[rng.Intn(len(utilityTenants))]
			monthsBack := rng.Intn(6)
			payDate := today.AddDate(0, -monthsBack, -rng.Intn(28))
			if payDate.Before(dateOnly(t.CheckinDate)) {
				payDate = dateOnly(t.CheckinDate).AddDate(0, 0, rng.Intn(20))
				if payDate.After(today) {
					continue
				}
			}
			ptype := model.PaymentTypeWater
			amount := 4000 + rng.Intn(8000) // 40–119 元
			if rng.Intn(2) == 0 {
				ptype = model.PaymentTypeElectricity
				amount = 8000 + rng.Intn(18000) // 80–259 元
			}
			payment := model.Payment{
				TenantID: t.ID,
				Amount:   amount,
				Type:     ptype,
				Paid:     rng.Intn(10) > 0, // ~90% paid
				PayDate:  payDate,
				Note:     utilityNote(ptype),
			}
			if err := tx.Create(&payment).Error; err != nil {
				return count, fmt.Errorf("seed utility payment: %w", err)
			}
			count++
		}
	}
	return count, nil
}

func buildRentPayment(tenant model.Tenant, due service.RentDue, today time.Time, rng *rand.Rand) model.Payment {
	payment := model.Payment{
		TenantID:      tenant.ID,
		Amount:        due.Amount,
		Type:          model.PaymentTypeRent,
		PayDate:       due.DueDate,
		AutoGenerated: true,
		Note:          "系统自动生成租金应收",
	}
	// Decide paid/unpaid/excluded distribution.
	roll := rng.Intn(100)
	switch {
	case tenant.Status == model.TenantStatusCheckout:
		// Checkout tenants: most paid, a couple excluded as the "已退租待处理" demo.
		if roll < 92 {
			payment.Paid = true
			payment.Note = "已收"
		} else {
			payment.Excluded = true
			payment.ExclusionNote = "已退租，押金抵扣"
			payment.Note = "退租后排除"
		}
	default:
		// Active tenants — only mark records up to today as resolved.
		if due.DueDate.After(today) {
			payment.Paid = false
			payment.Note = "待收"
			break
		}
		switch {
		case roll < 85:
			payment.Paid = true
			payment.Note = "已收"
		case roll < 93:
			payment.Paid = true
			lateDays := 1 + rng.Intn(7)
			late := due.DueDate.AddDate(0, 0, lateDays)
			if late.After(today) {
				late = today
			}
			payment.PayDate = late
			payment.Note = fmt.Sprintf("晚交 %d 天已收", lateDays)
		default:
			payment.Paid = false
			payment.Note = "逾期未收"
		}
	}
	return payment
}

func utilityNote(t string) string {
	switch t {
	case model.PaymentTypeWater:
		return "水费"
	case model.PaymentTypeElectricity:
		return "电费"
	default:
		return ""
	}
}

// ----- settings -------------------------------------------------------------

func seedSettings(tx *gorm.DB) (int, error) {
	entries := []model.AppSetting{
		{Key: "landlord_name", Value: "李房东"},
		{Key: "landlord_phone", Value: "13800000000"},
	}
	for _, entry := range entries {
		current := model.AppSetting{}
		if err := tx.Where("key = ?", entry.Key).Attrs(model.AppSetting{Key: entry.Key, Value: entry.Value}).FirstOrCreate(&current).Error; err != nil {
			return 0, fmt.Errorf("seed setting %s: %w", entry.Key, err)
		}
		if current.Value != entry.Value {
			current.Value = entry.Value
			if err := tx.Save(&current).Error; err != nil {
				return 0, fmt.Errorf("update setting %s: %w", entry.Key, err)
			}
		}
	}
	return len(entries), nil
}

func dateOnly(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
