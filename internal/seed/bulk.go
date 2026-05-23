package seed

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/service"
	"gorm.io/gorm"
)

// BulkConfig controls how many records RunBulk generates.
type BulkConfig struct {
	Rooms int
	Seed  int64
}

// RunBulk fills the database with a large volume of realistic data for
// performance testing. It runs inside a single transaction so a failure
// rolls everything back cleanly.
func RunBulk(db *gorm.DB, roomCount int, seedVal int64) (Stats, error) {
	now := time.Now()
	rng := rand.New(rand.NewSource(seedVal))
	cfg := BulkConfig{Rooms: roomCount, Seed: seedVal}

	var stats Stats
	err := db.Transaction(func(tx *gorm.DB) error {
		rooms, err := bulkRooms(tx, cfg, rng)
		if err != nil {
			return err
		}
		stats.Rooms = len(rooms)

		tenants, err := bulkTenants(tx, rooms, now, rng)
		if err != nil {
			return err
		}
		stats.Tenants = len(tenants)

		paymentCount, err := bulkPayments(tx, tenants, now, rng)
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

// ----- data pools ------------------------------------------------------------

// Common Chinese surnames (top 80).
var surnames = []string{
	"王", "李", "张", "刘", "陈", "杨", "赵", "黄", "周", "吴",
	"徐", "孙", "胡", "朱", "高", "林", "何", "郭", "马", "罗",
	"梁", "宋", "郑", "谢", "韩", "唐", "冯", "于", "董", "萧",
	"程", "曹", "袁", "邓", "许", "傅", "沈", "曾", "彭", "吕",
	"苏", "卢", "蒋", "蔡", "贾", "丁", "魏", "薛", "叶", "阎",
	"余", "潘", "杜", "戴", "夏", "钟", "汪", "田", "任", "姜",
	"范", "方", "石", "姚", "谭", "廖", "邹", "熊", "金", "陆",
	"郝", "孔", "白", "崔", "康", "毛", "邱", "秦", "江", "史",
}

// Common Chinese given-name characters (female-leaning + male-leaning mixed).
var givenChars = []string{
	"伟", "芳", "娜", "秀英", "敏", "静", "丽", "强", "磊", "军",
	"洋", "勇", "艳", "杰", "娟", "涛", "明", "超", "秀兰", "霞",
	"平", "刚", "桂英", "文", "华", "慧", "建", "国", "鑫", "晶",
	"晨", "宇", "婷", "浩", "欣", "佳", "俊", "博", "雨", "思",
	"子", "梓", "一", "子轩", "梓涵", "雨桐", "浩然", "诗涵", "可欣", "梦琪",
	"志远", "天佑", "嘉豪", "锦程", "瑞雪", "若曦", "星辰", "逸飞", "嘉琪", "紫萱",
}

// Street name pool for addresses.
var addresses = []string{
	"锦绣街 %d 号",
	"怡园路 %d 号",
	"云溪大道 %d 号",
	"翠竹巷 %d 号",
	"枫林路 %d 号",
	"桃源街 %d 号",
	"松柏巷 %d 号",
	"百合路 %d 号",
}

// Building name prefixes for titles.
var buildingNames = []string{
	"锦绣里", "怡园", "云溪", "翠竹苑", "枫林居",
	"桃源阁", "松柏轩", "百合庭", "紫荆园", "银杏居",
	"桂花苑", "柳荫轩", "兰亭居", "梅园", "竹韵庭",
	"荷塘月", "橡树湾", "棠梨苑", "梧桐里", "紫薇园",
	"青竹居", "丁香庭", "茉莉轩",
}

// Room type labels used in titles.
var roomTypeLabels = map[int]string{
	1: "一居",
	2: "两居",
	3: "三居",
}

var orientations = []string{"朝南", "朝北", "朝东", "朝西", "南北通透", "朝东南", "朝西南"}
var tagSets = [][]string{
	{"电梯", "空调"},
	{"电梯", "阳台", "空调"},
	{"近地铁", "采光佳"},
	{"整租", "拎包入住"},
	{"空调", "洗衣机"},
	{"电梯", "通风好"},
	{"近公园", "采光佳"},
	{"精装", "空调", "洗衣机"},
}

// Rent price ranges per bedroom count (yuan).
type priceRange struct{ min, max int }

var rentByBedrooms = map[int]priceRange{
	1: {2500, 3800},
	2: {4000, 5800},
	3: {6000, 8500},
}

// ----- rooms -----------------------------------------------------------------

func bulkRooms(tx *gorm.DB, cfg BulkConfig, rng *rand.Rand) ([]model.Room, error) {
	buildings := 'D'
	maxBuilding := 'Z'
	roomsPerFloor := 4
	floorsMin, floorsMax := 1, 25
	bedroomsOpts := []int{1, 2, 3}

	var rooms []model.Room
	rooms = make([]model.Room, 0, cfg.Rooms)

	buildingIdx := 0
	for b := buildings; b <= maxBuilding && len(rooms) < cfg.Rooms; b++ {
		addr := fmt.Sprintf(addresses[buildingIdx%len(addresses)], 10+rng.Intn(200))
		bName := buildingNames[buildingIdx%len(buildingNames)]

		for floor := floorsMin; floor <= floorsMax && len(rooms) < cfg.Rooms; floor++ {
			for unit := 1; unit <= roomsPerFloor && len(rooms) < cfg.Rooms; unit++ {
				bedrooms := bedroomsOpts[rng.Intn(len(bedroomsOpts))]
				pr := rentByBedrooms[bedrooms]
				rentYn := pr.min + rng.Intn(pr.max-pr.min+1)
				depositYn := rentYn * (1 + rng.Intn(3))

				rentFen, _ := service.ParseIntegerYuanToFen(fmt.Sprintf("%d", rentYn))
				depositFen, _ := service.ParseIntegerYuanToFen(fmt.Sprintf("%d", depositYn))

				// Distribute status: ~60% occupied, ~25% vacant, ~15% maintenance
				roll := rng.Intn(100)
				status := model.RoomStatusVacant
				switch {
				case roll < 60:
					status = model.RoomStatusOccupied
				case roll < 85:
					status = model.RoomStatusVacant
				default:
					status = model.RoomStatusMaintenance
				}

				orient := orientations[rng.Intn(len(orientations))]
				tags := tagSets[rng.Intn(len(tagSets))]
				tagStr := tags[0]
				for _, t := range tags[1:] {
					tagStr += "," + t
				}

				roomNo := fmt.Sprintf("%c%02d%02d", b, floor, unit)
				label := roomTypeLabels[bedrooms]
				area := 25 + bedrooms*15 + rng.Intn(20)

				rooms = append(rooms, model.Room{
					RoomNo:       roomNo,
					Title:        fmt.Sprintf("%s · %d楼%s", bName, floor, label),
					Description:  fmt.Sprintf("%d室%d厅，%s，面积%d㎡。", bedrooms, bedrooms-1+1, orient, area),
					Price:        rentFen,
					RentType:     model.RentTypeMonthly,
					RentPrice:    rentFen,
					PaymentTerms: model.PaymentTerms1M1D,
					Deposit:      depositFen,
					Status:       status,
					Area:         area,
					Floor:        floor,
					Address:      addr,
					Bedrooms:     bedrooms,
					LivingRooms:  bedrooms,
					Bathrooms:    1 + rng.Intn(2),
					Orientation:  orient,
					Tags:         tagStr,
				})
			}
		}
		buildingIdx++
	}

	if err := tx.CreateInBatches(&rooms, 100).Error; err != nil {
		return nil, fmt.Errorf("bulk insert rooms: %w", err)
	}
	log.Printf("  rooms: %d inserted", len(rooms))
	return rooms, nil
}

// ----- tenants ---------------------------------------------------------------

func bulkTenants(tx *gorm.DB, rooms []model.Room, now time.Time, rng *rand.Rand) ([]model.Tenant, error) {
	today := dateOnly(now)

	// Split rooms by status.
	var occupied, vacant []model.Room
	for _, r := range rooms {
		switch r.Status {
		case model.RoomStatusOccupied:
			occupied = append(occupied, r)
		case model.RoomStatusVacant:
			vacant = append(vacant, r)
		}
	}

	// One active tenant per occupied room.
	activeCount := len(occupied)
	checkoutCount := len(vacant) / 4 // use some vacant rooms for checkout tenants
	if checkoutCount > 250 {
		checkoutCount = 250
	}

	total := activeCount + checkoutCount
	tenants := make([]model.Tenant, 0, total)
	genderOpts := []string{model.TenantGenderMale, model.TenantGenderFemale}
	leaseOpts := []int{6, 12, 24}
	termsOpts := []string{model.PaymentTerms1M1D, model.PaymentTerms3M1D}

	phoneSeq := 10000000

	// Active tenants.
	for _, room := range occupied {
		monthsAgo := 1 + rng.Intn(35)
		checkin := today.AddDate(0, -monthsAgo, -rng.Intn(28))
		leaseMonths := leaseOpts[rng.Intn(len(leaseOpts))]
		leaseEnd := checkin.AddDate(0, leaseMonths, 0)
		gender := genderOpts[rng.Intn(len(genderOpts))]

		tenants = append(tenants, model.Tenant{
			Name:             randomName(rng),
			Phone:            fmt.Sprintf("138%08d", phoneSeq),
			EmergencyContact: randomName(rng) + " " + fmt.Sprintf("139%08d", phoneSeq+10000000),
			Gender:           gender,
			RoomID:           room.ID,
			CheckinDate:      checkin,
			LeaseEndDate:     &leaseEnd,
			RentPrice:        room.RentPrice,
			RentType:         model.RentTypeMonthly,
			PaymentTerms:     termsOpts[rng.Intn(len(termsOpts))],
			Deposit:          room.Deposit,
			Notes:            "",
			Status:           model.TenantStatusActive,
		})
		phoneSeq++
	}

	// Checkout tenants — multiple per vacant room is fine.
	for i := 0; i < checkoutCount; i++ {
		room := vacant[i%len(vacant)]
		monthsAgo := 6 + rng.Intn(30)
		checkin := today.AddDate(0, -monthsAgo, -rng.Intn(28))
		checkoutDaysAgo := 7 + rng.Intn(90)
		checkout := today.AddDate(0, 0, -checkoutDaysAgo)
		gender := genderOpts[rng.Intn(len(genderOpts))]

		tenants = append(tenants, model.Tenant{
			Name:             randomName(rng),
			Phone:            fmt.Sprintf("138%08d", phoneSeq),
			EmergencyContact: randomName(rng) + " " + fmt.Sprintf("139%08d", phoneSeq+10000000),
			Gender:           gender,
			RoomID:           room.ID,
			CheckinDate:      checkin,
			CheckoutDate:     &checkout,
			RentPrice:        room.RentPrice,
			RentType:         model.RentTypeMonthly,
			PaymentTerms:     model.PaymentTerms1M1D,
			Deposit:          room.Deposit,
			Status:           model.TenantStatusCheckout,
		})
		phoneSeq++
	}

	if err := tx.CreateInBatches(&tenants, 100).Error; err != nil {
		return nil, fmt.Errorf("bulk insert tenants: %w", err)
	}
	log.Printf("  tenants: %d inserted (active=%d, checkout=%d)", len(tenants), activeCount, checkoutCount)
	return tenants, nil
}

func randomName(rng *rand.Rand) string {
	return surnames[rng.Intn(len(surnames))] + givenChars[rng.Intn(len(givenChars))]
}

// ----- payments --------------------------------------------------------------

func bulkPayments(tx *gorm.DB, tenants []model.Tenant, now time.Time, rng *rand.Rand) (int, error) {
	today := dateOnly(now)
	total := 0

	for i, tenant := range tenants {
		end := today
		if tenant.Status == model.TenantStatusCheckout && tenant.CheckoutDate != nil {
			end = dateOnly(*tenant.CheckoutDate)
		}

		// Rent dues via business logic.
		dues := service.RentDueSchedule(tenant, end)
		for _, due := range dues {
			payment := buildRentPayment(tenant, due, today, rng)
			if err := tx.Create(&payment).Error; err != nil {
				return total, fmt.Errorf("rent payment tenant=%d: %w", tenant.ID, err)
			}
			total++
		}

		// Utility payments: ~3-7 per tenant.
		utilCount := 3 + rng.Intn(5)
		monthsBack := 0
		if tenant.Status == model.TenantStatusCheckout && tenant.CheckoutDate != nil {
			monthsBack = int(today.Sub(dateOnly(*tenant.CheckoutDate)).Hours() / 24 / 30)
		} else {
			monthsBack = int(today.Sub(dateOnly(tenant.CheckinDate)).Hours() / 24 / 30)
		}
		if monthsBack > 36 {
			monthsBack = 36
		}

		for j := 0; j < utilCount; j++ {
			mBack := rng.Intn(monthsBack + 1)
			payDate := today.AddDate(0, -mBack, -rng.Intn(28))
			if payDate.Before(dateOnly(tenant.CheckinDate)) {
				payDate = dateOnly(tenant.CheckinDate).AddDate(0, 0, rng.Intn(20))
			}
			if payDate.After(today) {
				continue
			}

			ptype := model.PaymentTypeWater
			amount := 4000 + rng.Intn(8000)
			if rng.Intn(2) == 0 {
				ptype = model.PaymentTypeElectricity
				amount = 8000 + rng.Intn(18000)
			}

			if err := tx.Create(&model.Payment{
				TenantID: tenant.ID,
				Amount:   amount,
				Type:     ptype,
				Paid:     rng.Intn(10) > 0,
				PayDate:  payDate,
				Note:     utilityNote(ptype),
			}).Error; err != nil {
				return total, fmt.Errorf("utility payment: %w", err)
			}
			total++
		}

		if (i+1)%100 == 0 {
			log.Printf("  payments: %d tenants processed, %d payments so far", i+1, total)
		}
	}

	log.Printf("  payments: %d total inserted", total)
	return total, nil
}
