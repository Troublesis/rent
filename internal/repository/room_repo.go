package repository

import (
	"errors"
	"strings"

	"github.com/troublesis/rent/internal/model"
	"gorm.io/gorm"
)

type RoomFilter struct {
	Status         string
	Query          string
	SortBy         string
	SortDir        string
	Floor          int
	HasFloor       bool
	Bedrooms       int
	HasBedrooms    bool
	LivingRooms    int
	HasLivingRooms bool
	Bathrooms      int
	HasBathrooms   bool
	Limit          int
	Offset         int
}

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) WithDB(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) ListRooms(filter RoomFilter) ([]model.Room, error) {
	query := r.db.Model(&model.Room{}).Preload("Media")
	query = applyRoomFilter(query, filter)
	query = applyRoomSort(query, filter)
	var rooms []model.Room
	if err := query.Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *RoomRepository) ListPublicAvailableRooms(limit int, offset int) ([]model.Room, error) {
	return r.ListRooms(RoomFilter{Status: model.RoomStatusVacant, Limit: limit, Offset: offset})
}

func (r *RoomRepository) ListPublicAvailableRoomFacets() ([]model.Room, error) {
	var rooms []model.Room
	if err := r.db.Model(&model.Room{}).
		Select("floor", "bedrooms", "living_rooms", "bathrooms").
		Where("status = ?", model.RoomStatusVacant).
		Order("floor ASC, bedrooms ASC, living_rooms ASC, bathrooms ASC").
		Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *RoomRepository) GetRoom(id uint) (*model.Room, error) {
	var room model.Room
	if err := r.db.First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) GetRoomWithMedia(id uint) (*model.Room, error) {
	var room model.Room
	if err := r.db.Preload("Media").First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) CreateRoom(room *model.Room) error {
	return r.db.Create(room).Error
}

func (r *RoomRepository) UpdateRoom(room *model.Room) error {
	return r.db.Save(room).Error
}

func (r *RoomRepository) DeleteRoom(id uint) error {
	result := r.db.Delete(&model.Room{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *RoomRepository) AddRoomMedia(media *model.RoomMedia) error {
	return r.db.Create(media).Error
}

func (r *RoomRepository) DeleteRoomMedia(id uint) error {
	result := r.db.Delete(&model.RoomMedia{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *RoomRepository) CountByStatus(status string) (int64, error) {
	query := r.db.Model(&model.Room{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func applyRoomFilter(query *gorm.DB, filter RoomFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Query != "" {
		like := "%" + filter.Query + "%"
		query = query.Where("room_no LIKE ? OR title LIKE ? OR (room_no || ' - ' || title) LIKE ?", like, like, like)
	}
	if filter.HasFloor || filter.Floor > 0 {
		query = query.Where("floor = ?", filter.Floor)
	}
	if filter.HasBedrooms || filter.Bedrooms > 0 {
		query = query.Where("bedrooms = ?", filter.Bedrooms)
	}
	if filter.HasLivingRooms || filter.LivingRooms > 0 {
		query = query.Where("living_rooms = ?", filter.LivingRooms)
	}
	if filter.HasBathrooms || filter.Bathrooms > 0 {
		query = query.Where("bathrooms = ?", filter.Bathrooms)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	return query
}

func applyRoomSort(query *gorm.DB, filter RoomFilter) *gorm.DB {
	direction := "DESC"
	if strings.EqualFold(filter.SortDir, "asc") {
		direction = "ASC"
	}
	switch filter.SortBy {
	case "room_no":
		return query.Order("room_no " + direction).Order("id ASC")
	case "title":
		return query.Order("title " + direction).Order("id ASC")
	case "rent_price":
		return query.Order("COALESCE(NULLIF(rent_price, 0), price) " + direction).Order("id ASC")
	case "status":
		return query.Order("status " + direction).Order("id ASC")
	default:
		return query.Order("created_at DESC, id DESC")
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
