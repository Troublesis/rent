package repository

import (
	"testing"

	"github.com/troublesis/rent/internal/model"
)

func TestRoomRepositoryListAndPreloadMedia(t *testing.T) {
	db := newTestDB(t)
	repo := NewRoomRepository(db)
	room := &model.Room{RoomNo: "C301", Title: "带阳台单间", Price: 180000, Deposit: 180000, Status: model.RoomStatusVacant}
	if err := repo.CreateRoom(room); err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	if err := repo.AddRoomMedia(&model.RoomMedia{RoomID: room.ID, URL: "/uploads/1/a.jpg", MediaType: model.MediaTypeImage}); err != nil {
		t.Fatalf("AddRoomMedia returned error: %v", err)
	}

	rooms, err := repo.ListRooms(RoomFilter{Status: model.RoomStatusVacant})
	if err != nil {
		t.Fatalf("ListRooms returned error: %v", err)
	}
	if len(rooms) != 1 {
		t.Fatalf("len(rooms) = %d, want 1", len(rooms))
	}
	if len(rooms[0].Media) != 1 {
		t.Fatalf("len(Media) = %d, want 1", len(rooms[0].Media))
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
