package repository

import (
	"testing"

	"github.com/troublesis/rent/internal/model"
)

func TestRoomRepositoryListAndPreloadMedia(t *testing.T) {
	db := newTestDB(t)
	repo := NewRoomRepository(db)
	room := &model.Room{
		RoomNo:       "C301",
		Title:        "带阳台单间",
		Price:        180000,
		RentType:     model.RentTypeMonthly,
		RentPrice:    180000,
		PaymentTerms: model.PaymentTerms1M1D,
		Deposit:      180000,
		Status:       model.RoomStatusVacant,
		Address:      "北京市朝阳区示例路 1 号",
		Bedrooms:     1,
		LivingRooms:  1,
		Bathrooms:    1,
		Orientation:  "南",
	}
	if err := repo.CreateRoom(room); err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	if err := repo.AddRoomMedia(&model.RoomMedia{RoomID: room.ID, URL: "/uploads/1/a.jpg", MediaType: model.MediaTypeImage}); err != nil {
		t.Fatalf("AddRoomMedia image returned error: %v", err)
	}
	if err := repo.AddRoomMedia(&model.RoomMedia{RoomID: room.ID, URL: "https://v.douyin.com/example", MediaType: model.MediaTypeVideoLink}); err != nil {
		t.Fatalf("AddRoomMedia video link returned error: %v", err)
	}

	rooms, err := repo.ListRooms(RoomFilter{Status: model.RoomStatusVacant})
	if err != nil {
		t.Fatalf("ListRooms returned error: %v", err)
	}
	if len(rooms) != 1 {
		t.Fatalf("len(rooms) = %d, want 1", len(rooms))
	}
	if len(rooms[0].Media) != 2 {
		t.Fatalf("len(Media) = %d, want 2", len(rooms[0].Media))
	}
	if rooms[0].RentType != model.RentTypeMonthly || rooms[0].PaymentTerms != model.PaymentTerms1M1D || rooms[0].Address == "" {
		t.Fatalf("room new fields were not persisted: %#v", rooms[0])
	}
}

func TestRoomRepositoryCountByStatus(t *testing.T) {
	db := newTestDB(t)
	repo := NewRoomRepository(db)
	rooms := []*model.Room{
		{RoomNo: "C302", Title: "一号", Status: model.RoomStatusVacant},
		{RoomNo: "C303", Title: "二号", Status: model.RoomStatusOccupied},
	}
	for _, room := range rooms {
		if err := repo.CreateRoom(room); err != nil {
			t.Fatalf("CreateRoom returned error: %v", err)
		}
	}
	count, err := repo.CountByStatus(model.RoomStatusVacant)
	if err != nil {
		t.Fatalf("CountByStatus returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}
