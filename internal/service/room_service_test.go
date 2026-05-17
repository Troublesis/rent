package service

import (
	"testing"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

func TestCreateRoomStoresNewFields(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)

	input := validRoomInput("R101", "新户型房源", "1800")
	input.PaymentTerms = model.PaymentTerms6M0D
	input.Orientation = "东南"
	room, err := roomService.CreateRoom(input)
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	if room.RentPrice != 180000 || room.Price != 180000 {
		t.Fatalf("rent price/legacy price = %d/%d, want 180000/180000", room.RentPrice, room.Price)
	}
	if room.PaymentTerms != model.PaymentTerms6M0D || room.Address == "" || room.Orientation != "东南" {
		t.Fatalf("room fields not saved: %#v", room)
	}
}

func TestCreateRoomValidatesRequiredFields(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)

	input := validRoomInput("R102", "地址缺失房源", "1800")
	input.Address = "短"
	if _, err := roomService.CreateRoom(input); err == nil {
		t.Fatal("CreateRoom returned nil error for short address")
	}
}

func TestAddRoomVideoLink(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)

	room, err := roomService.CreateRoom(validRoomInput("R103", "视频房源", "1800"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	if err := roomService.AddRoomVideoLink(room.ID, "https://v.douyin.com/example"); err != nil {
		t.Fatalf("AddRoomVideoLink returned error: %v", err)
	}
	if err := roomService.AddRoomVideoLink(room.ID, "ftp://example.com/video"); err == nil {
		t.Fatal("AddRoomVideoLink returned nil error for invalid URL")
	}
	updatedRoom, err := roomService.GetRoom(room.ID)
	if err != nil {
		t.Fatalf("GetRoom returned error: %v", err)
	}
	if len(updatedRoom.Media) != 1 || updatedRoom.Media[0].MediaType != model.MediaTypeVideoLink {
		t.Fatalf("media = %#v, want one video link", updatedRoom.Media)
	}
}
