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

func TestRoomRepositoryListRoomsSortsByRoomNumberAscending(t *testing.T) {
	db := newTestDB(t)
	repo := NewRoomRepository(db)
	rooms := []*model.Room{
		{RoomNo: "C301", Title: "三号", Status: model.RoomStatusVacant, RentPrice: 30000},
		{RoomNo: "A101", Title: "一号", Status: model.RoomStatusVacant, RentPrice: 10000},
		{RoomNo: "B201", Title: "二号", Status: model.RoomStatusVacant, RentPrice: 20000},
	}
	for _, room := range rooms {
		if err := repo.CreateRoom(room); err != nil {
			t.Fatalf("CreateRoom returned error: %v", err)
		}
	}

	sortedRooms, err := repo.ListRooms(RoomFilter{Status: model.RoomStatusVacant, SortBy: "room_no", SortDir: "asc"})
	if err != nil {
		t.Fatalf("ListRooms returned error: %v", err)
	}
	got := []string{sortedRooms[0].RoomNo, sortedRooms[1].RoomNo, sortedRooms[2].RoomNo}
	want := []string{"A101", "B201", "C301"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("room order = %v, want %v", got, want)
		}
	}
}

func TestRoomRepositoryListRoomsFiltersByRoomOptionLabel(t *testing.T) {
	db := newTestDB(t)
	repo := NewRoomRepository(db)
	rooms := []*model.Room{
		{RoomNo: "A101", Title: "温馨单间", Status: model.RoomStatusVacant},
		{RoomNo: "A102", Title: "整洁两居", Status: model.RoomStatusVacant},
	}
	for _, room := range rooms {
		if err := repo.CreateRoom(room); err != nil {
			t.Fatalf("CreateRoom returned error: %v", err)
		}
	}

	filteredRooms, err := repo.ListRooms(RoomFilter{Status: model.RoomStatusVacant, Query: "A101 - 温馨单间"})
	if err != nil {
		t.Fatalf("ListRooms returned error: %v", err)
	}
	if len(filteredRooms) != 1 || filteredRooms[0].RoomNo != "A101" {
		t.Fatalf("filteredRooms = %#v, want only A101", filteredRooms)
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

func TestRoomRepositoryListRoomsFiltersByFloorAndLayout(t *testing.T) {
	db := newTestDB(t)
	repo := NewRoomRepository(db)
	rooms := []*model.Room{
		{RoomNo: "A101", Title: "一楼一居", Status: model.RoomStatusVacant, Floor: 1, Bedrooms: 1, LivingRooms: 1, Bathrooms: 1},
		{RoomNo: "A202", Title: "二楼两居", Status: model.RoomStatusVacant, Floor: 2, Bedrooms: 2, LivingRooms: 1, Bathrooms: 1},
		{RoomNo: "A203", Title: "二楼三居", Status: model.RoomStatusVacant, Floor: 2, Bedrooms: 3, LivingRooms: 1, Bathrooms: 2},
	}
	for _, room := range rooms {
		if err := repo.CreateRoom(room); err != nil {
			t.Fatalf("CreateRoom returned error: %v", err)
		}
	}

	filteredRooms, err := repo.ListRooms(RoomFilter{Status: model.RoomStatusVacant, Floor: 2, Bedrooms: 3, LivingRooms: 1, Bathrooms: 2})
	if err != nil {
		t.Fatalf("ListRooms returned error: %v", err)
	}
	if len(filteredRooms) != 1 || filteredRooms[0].RoomNo != "A203" {
		t.Fatalf("filteredRooms = %#v, want only A203", filteredRooms)
	}
}

func TestRoomRepositoryListRoomsFiltersZeroValuedLayoutParts(t *testing.T) {
	db := newTestDB(t)
	repo := NewRoomRepository(db)
	rooms := []*model.Room{
		{RoomNo: "B101", Title: "一室无厅", Status: model.RoomStatusVacant, Floor: 1, Bedrooms: 1, LivingRooms: 0, Bathrooms: 1},
		{RoomNo: "B102", Title: "一室一厅", Status: model.RoomStatusVacant, Floor: 1, Bedrooms: 1, LivingRooms: 1, Bathrooms: 1},
	}
	for _, room := range rooms {
		if err := repo.CreateRoom(room); err != nil {
			t.Fatalf("CreateRoom returned error: %v", err)
		}
	}

	filteredRooms, err := repo.ListRooms(RoomFilter{Status: model.RoomStatusVacant, Bedrooms: 1, HasBedrooms: true, LivingRooms: 0, HasLivingRooms: true, Bathrooms: 1, HasBathrooms: true})
	if err != nil {
		t.Fatalf("ListRooms returned error: %v", err)
	}
	if len(filteredRooms) != 1 || filteredRooms[0].RoomNo != "B101" {
		t.Fatalf("filteredRooms = %#v, want only B101", filteredRooms)
	}
}

func TestRoomRepositoryListRoomsUsesLimitAndOffset(t *testing.T) {
	db := newTestDB(t)
	repo := NewRoomRepository(db)
	rooms := []*model.Room{
		{RoomNo: "A101", Title: "第一间", Status: model.RoomStatusVacant},
		{RoomNo: "A102", Title: "第二间", Status: model.RoomStatusVacant},
		{RoomNo: "A103", Title: "第三间", Status: model.RoomStatusVacant},
	}
	for _, room := range rooms {
		if err := repo.CreateRoom(room); err != nil {
			t.Fatalf("CreateRoom returned error: %v", err)
		}
	}

	pagedRooms, err := repo.ListRooms(RoomFilter{Status: model.RoomStatusVacant, Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("ListRooms returned error: %v", err)
	}
	if len(pagedRooms) != 1 {
		t.Fatalf("len(pagedRooms) = %d, want 1", len(pagedRooms))
	}
}
