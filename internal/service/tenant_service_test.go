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

	room, err := roomService.CreateRoom(RoomInput{
		RoomNo:      "A101",
		Title:       "南向一居室",
		PriceYuan:   "1500",
		DepositYuan: "1500",
		Status:      model.RoomStatusVacant,
	})
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
	})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	if tenant.Status != model.TenantStatusActive {
		t.Fatalf("tenant status = %q, want active", tenant.Status)
	}

	updatedRoom, err := roomRepo.GetRoom(room.ID)
	if err != nil {
		t.Fatalf("GetRoom returned error: %v", err)
	}
	if updatedRoom.Status != model.RoomStatusOccupied {
		t.Fatalf("room status = %q, want occupied", updatedRoom.Status)
	}
}

func TestCheckInTenantFailsWhenRoomOccupied(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)

	room, err := roomService.CreateRoom(RoomInput{RoomNo: "A102", Title: "一居室", PriceYuan: "1600", DepositYuan: "1600", Status: model.RoomStatusVacant})
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

func TestCheckOutTenant(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)

	room, err := roomService.CreateRoom(RoomInput{RoomNo: "A103", Title: "两居室", PriceYuan: "2200", DepositYuan: "2200", Status: model.RoomStatusVacant})
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
