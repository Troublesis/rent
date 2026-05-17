package service

import (
	"testing"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

func TestCheckInTenant(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)

	room, err := roomService.CreateRoom(validRoomInput("A101", "南向一居室", "1500"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}

	tenant, err := tenantService.CheckInTenant(TenantInput{
		Name:          "张三",
		Phone:         "13800000000",
		RoomID:        room.ID,
		CheckinDate:   time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local),
		RentPriceYuan: "1500",
		DepositYuan:   "1500",
		Notes:         "合同已线下签署",
	})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	if tenant.Status != model.TenantStatusActive {
		t.Fatalf("tenant status = %q, want active", tenant.Status)
	}
	if tenant.RentType != model.RentTypeMonthly || tenant.PaymentTerms != model.PaymentTerms1M1D {
		t.Fatalf("tenant rent snapshot = %q/%q, want monthly/1m_1d", tenant.RentType, tenant.PaymentTerms)
	}
	if tenant.Notes != "合同已线下签署" {
		t.Fatalf("tenant notes = %q", tenant.Notes)
	}
	if tenant.Gender != "" {
		t.Fatalf("tenant gender = %q, want blank", tenant.Gender)
	}

	updatedRoom, err := roomRepo.GetRoom(room.ID)
	if err != nil {
		t.Fatalf("GetRoom returned error: %v", err)
	}
	if updatedRoom.Status != model.RoomStatusOccupied {
		t.Fatalf("room status = %q, want occupied", updatedRoom.Status)
	}
}

func TestBuildTenantAcceptsOptionalGender(t *testing.T) {
	baseInput := TenantInput{
		Name:          "性别租客",
		Phone:         "13800000010",
		RoomID:        1,
		RentPriceYuan: "1500",
		DepositYuan:   "1500",
	}

	maleInput := baseInput
	maleInput.Gender = model.TenantGenderMale
	maleTenant, err := buildTenant(maleInput)
	if err != nil {
		t.Fatalf("buildTenant male returned error: %v", err)
	}
	if maleTenant.Gender != model.TenantGenderMale {
		t.Fatalf("male gender = %q, want male", maleTenant.Gender)
	}

	femaleInput := baseInput
	femaleInput.Gender = model.TenantGenderFemale
	femaleTenant, err := buildTenant(femaleInput)
	if err != nil {
		t.Fatalf("buildTenant female returned error: %v", err)
	}
	if femaleTenant.Gender != model.TenantGenderFemale {
		t.Fatalf("female gender = %q, want female", femaleTenant.Gender)
	}
}

func TestBuildTenantRejectsInvalidGender(t *testing.T) {
	_, err := buildTenant(TenantInput{
		Name:          "错误性别",
		Phone:         "13800000011",
		Gender:        "unknown",
		RoomID:        1,
		RentPriceYuan: "1500",
		DepositYuan:   "1500",
	})
	if err == nil {
		t.Fatal("buildTenant returned nil error")
	}
}

func TestCheckInTenantDefaultsLeaseTermsFromRoom(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)

	input := validRoomInput("A104", "日租公寓", "80")
	input.RentType = model.RentTypeDaily
	input.PaymentTerms = model.PaymentTerms3M1D
	room, err := roomService.CreateRoom(input)
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{Name: "周九", Phone: "13800000009", RoomID: room.ID})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	if tenant.RentType != model.RentTypeDaily {
		t.Fatalf("tenant rent type = %q, want daily", tenant.RentType)
	}
	if tenant.PaymentTerms != model.PaymentTerms3M1D {
		t.Fatalf("tenant payment terms = %q, want 3m_1d", tenant.PaymentTerms)
	}
	if tenant.RentPrice != 8000 {
		t.Fatalf("tenant rent price = %d, want 8000", tenant.RentPrice)
	}
}

func TestCheckInTenantStoresOptionalLeaseEndDate(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)

	room, err := roomService.CreateRoom(validRoomInput("A105", "长租公寓", "1800"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	leaseEndDate := time.Date(2027, time.May, 1, 0, 0, 0, 0, time.Local)
	tenant, err := tenantService.CheckInTenant(TenantInput{
		Name:          "郑十",
		Phone:         "13800000012",
		RoomID:        room.ID,
		CheckinDate:   time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local),
		LeaseEndDate:  leaseEndDate,
		RentPriceYuan: "1800",
		DepositYuan:   "1800",
	})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	if tenant.LeaseEndDate == nil || !tenant.LeaseEndDate.Equal(leaseEndDate) {
		t.Fatalf("LeaseEndDate = %v, want %v", tenant.LeaseEndDate, leaseEndDate)
	}
}

func TestCheckInTenantRejectsLeaseEndBeforeCheckin(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)

	room, err := roomService.CreateRoom(validRoomInput("A106", "短租公寓", "1800"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	_, err = tenantService.CheckInTenant(TenantInput{
		Name:          "王十",
		Phone:         "13800000013",
		RoomID:        room.ID,
		CheckinDate:   time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local),
		LeaseEndDate:  time.Date(2026, time.April, 30, 0, 0, 0, 0, time.Local),
		RentPriceYuan: "1800",
		DepositYuan:   "1800",
	})
	if err == nil {
		t.Fatal("CheckInTenant returned nil error")
	}
}

func TestCheckInTenantFailsWhenRoomOccupied(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)

	room, err := roomService.CreateRoom(validRoomInput("A102", "一居室", "1600"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	input := TenantInput{Name: "李四", Phone: "13800000001", RoomID: room.ID, RentPriceYuan: "1600", DepositYuan: "1600"}
	if _, err := tenantService.CheckInTenant(input); err != nil {
		t.Fatalf("first CheckInTenant returned error: %v", err)
	}
	if _, err := tenantService.CheckInTenant(TenantInput{Name: "王五", Phone: "13800000002", RoomID: room.ID, RentPriceYuan: "1600", DepositYuan: "1600"}); err == nil {
		t.Fatal("second CheckInTenant returned nil error")
	}
}

func TestUpdateTenantChangesInfoAndRoomStatus(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)

	oldRoom, err := roomService.CreateRoom(validRoomInput("A107", "旧房源", "1800"))
	if err != nil {
		t.Fatalf("CreateRoom old returned error: %v", err)
	}
	newRoom, err := roomService.CreateRoom(validRoomInput("A108", "新房源", "2200"))
	if err != nil {
		t.Fatalf("CreateRoom new returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{Name: "编辑前", Phone: "13800000020", RoomID: oldRoom.ID, RentPriceYuan: "1800", DepositYuan: "1800"})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	leaseEndDate := time.Date(2027, time.June, 1, 0, 0, 0, 0, time.Local)

	updatedTenant, err := tenantService.UpdateTenant(tenant.ID, TenantInput{
		Name:             "编辑后",
		Phone:            "13800000021",
		EmergencyContact: "13800000022",
		Gender:           model.TenantGenderFemale,
		RoomID:           newRoom.ID,
		CheckinDate:      time.Date(2026, time.June, 1, 0, 0, 0, 0, time.Local),
		LeaseEndDate:     leaseEndDate,
		RentPriceYuan:    "2200",
		RentType:         model.RentTypeMonthly,
		PaymentTerms:     model.PaymentTerms3M1D,
		DepositYuan:      "3000",
		Notes:            "已更新资料",
	})
	if err != nil {
		t.Fatalf("UpdateTenant returned error: %v", err)
	}
	if updatedTenant.Name != "编辑后" || updatedTenant.Phone != "13800000021" || updatedTenant.RoomID != newRoom.ID {
		t.Fatalf("updated tenant = %#v, want edited identity and room", updatedTenant)
	}
	if updatedTenant.Status != model.TenantStatusActive || updatedTenant.CheckoutDate != nil {
		t.Fatalf("tenant status = %q checkout = %v, want active without checkout", updatedTenant.Status, updatedTenant.CheckoutDate)
	}
	if updatedTenant.LeaseEndDate == nil || !updatedTenant.LeaseEndDate.Equal(leaseEndDate) {
		t.Fatalf("lease end date = %v, want %v", updatedTenant.LeaseEndDate, leaseEndDate)
	}
	updatedOldRoom, err := roomRepo.GetRoom(oldRoom.ID)
	if err != nil {
		t.Fatalf("GetRoom old returned error: %v", err)
	}
	if updatedOldRoom.Status != model.RoomStatusVacant {
		t.Fatalf("old room status = %q, want vacant", updatedOldRoom.Status)
	}
	updatedNewRoom, err := roomRepo.GetRoom(newRoom.ID)
	if err != nil {
		t.Fatalf("GetRoom new returned error: %v", err)
	}
	if updatedNewRoom.Status != model.RoomStatusOccupied {
		t.Fatalf("new room status = %q, want occupied", updatedNewRoom.Status)
	}
}

func TestUpdateTenantRejectsOccupiedTargetRoom(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)

	firstRoom, err := roomService.CreateRoom(validRoomInput("A109", "第一房源", "1800"))
	if err != nil {
		t.Fatalf("CreateRoom first returned error: %v", err)
	}
	secondRoom, err := roomService.CreateRoom(validRoomInput("A110", "第二房源", "2200"))
	if err != nil {
		t.Fatalf("CreateRoom second returned error: %v", err)
	}
	firstTenant, err := tenantService.CheckInTenant(TenantInput{Name: "第一租客", Phone: "13800000023", RoomID: firstRoom.ID, RentPriceYuan: "1800", DepositYuan: "1800"})
	if err != nil {
		t.Fatalf("CheckInTenant first returned error: %v", err)
	}
	if _, err := tenantService.CheckInTenant(TenantInput{Name: "第二租客", Phone: "13800000024", RoomID: secondRoom.ID, RentPriceYuan: "2200", DepositYuan: "2200"}); err != nil {
		t.Fatalf("CheckInTenant second returned error: %v", err)
	}

	_, err = tenantService.UpdateTenant(firstTenant.ID, TenantInput{Name: "第一租客", Phone: "13800000023", RoomID: secondRoom.ID, RentPriceYuan: "1800", DepositYuan: "1800"})
	if err == nil {
		t.Fatal("UpdateTenant returned nil error")
	}
}

func TestCheckOutTenant(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)

	room, err := roomService.CreateRoom(validRoomInput("A103", "两居室", "2200"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{Name: "赵六", Phone: "13800000003", RoomID: room.ID, RentPriceYuan: "2200", DepositYuan: "2200"})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	if err := tenantService.CheckOutTenant(tenant.ID); err != nil {
		t.Fatalf("CheckOutTenant returned error: %v", err)
	}

	updatedTenant, err := tenantRepo.GetTenant(tenant.ID)
	if err != nil {
		t.Fatalf("GetTenant returned error: %v", err)
	}
	if updatedTenant.Status != model.TenantStatusCheckout {
		t.Fatalf("tenant status = %q, want checkout", updatedTenant.Status)
	}
	updatedRoom, err := roomRepo.GetRoom(room.ID)
	if err != nil {
		t.Fatalf("GetRoom returned error: %v", err)
	}
	if updatedRoom.Status != model.RoomStatusVacant {
		t.Fatalf("room status = %q, want vacant", updatedRoom.Status)
	}
}

func validRoomInput(roomNo string, title string, rent string) RoomInput {
	return RoomInput{
		RoomNo:        roomNo,
		Title:         title,
		RentType:      model.RentTypeMonthly,
		RentPriceYuan: rent,
		PaymentTerms:  model.PaymentTerms1M1D,
		DepositYuan:   rent,
		Status:        model.RoomStatusVacant,
		Area:          35,
		Floor:         3,
		Address:       "北京市朝阳区示例路 1 号",
		Bedrooms:      1,
		LivingRooms:   1,
		Bathrooms:     1,
	}
}
