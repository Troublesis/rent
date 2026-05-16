package service

import (
	"fmt"
	"strings"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

type RoomInput struct {
	RoomNo      string
	Title       string
	Description string
	PriceYuan   string
	DepositYuan string
	Status      string
	Area        int
	Floor       int
	Tags        string
}

type RoomService struct {
	roomRepo   *repository.RoomRepository
	tenantRepo *repository.TenantRepository
}

func NewRoomService(roomRepo *repository.RoomRepository, tenantRepo *repository.TenantRepository) *RoomService {
	return &RoomService{roomRepo: roomRepo, tenantRepo: tenantRepo}
}

func (s *RoomService) ListRooms(filter repository.RoomFilter) ([]model.Room, error) {
	return s.roomRepo.ListRooms(filter)
}

func (s *RoomService) ListAvailableRooms(limit int, offset int) ([]model.Room, error) {
	return s.roomRepo.ListPublicAvailableRooms(limit, offset)
}

func (s *RoomService) GetRoom(id uint) (*model.Room, error) {
	return s.roomRepo.GetRoomWithMedia(id)
}

func (s *RoomService) CreateRoom(input RoomInput) (*model.Room, error) {
	room, err := buildRoom(input, 0)
	if err != nil {
		return nil, err
	}
	if err := s.roomRepo.CreateRoom(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) UpdateRoom(id uint, input RoomInput) (*model.Room, error) {
	current, err := s.roomRepo.GetRoom(id)
	if err != nil {
		return nil, err
	}
	room, err := buildRoom(input, id)
	if err != nil {
		return nil, err
	}
	room.CreatedAt = current.CreatedAt
	if err := s.roomRepo.UpdateRoom(room); err != nil {
		return nil, err
	}
	return s.roomRepo.GetRoomWithMedia(id)
}

func (s *RoomService) DeleteRoom(id uint) error {
	room, err := s.roomRepo.GetRoom(id)
	if err != nil {
		return err
	}
	if room.Status == model.RoomStatusOccupied {
		return fmt.Errorf("occupied rooms cannot be deleted")
	}
	if _, err := s.tenantRepo.GetActiveTenantByRoomID(id); err == nil {
		return fmt.Errorf("rooms with active tenants cannot be deleted")
	} else if !repository.IsNotFound(err) {
		return err
	}
	return s.roomRepo.DeleteRoom(id)
}

func (s *RoomService) AddRoomMedia(roomID uint, url string, mediaType string) error {
	if _, err := s.roomRepo.GetRoom(roomID); err != nil {
		return err
	}
	if mediaType != model.MediaTypeImage && mediaType != model.MediaTypeVideo {
		return fmt.Errorf("invalid media type")
	}
	media := &model.RoomMedia{RoomID: roomID, URL: strings.TrimSpace(url), MediaType: mediaType}
	return s.roomRepo.AddRoomMedia(media)
}

func buildRoom(input RoomInput, id uint) (*model.Room, error) {
	roomNo := strings.TrimSpace(input.RoomNo)
	title := strings.TrimSpace(input.Title)
	if roomNo == "" {
		return nil, fmt.Errorf("room number is required")
	}
	if title == "" {
		return nil, fmt.Errorf("room title is required")
	}
	price, err := ParseYuanToFen(input.PriceYuan)
	if err != nil {
		return nil, fmt.Errorf("invalid room price: %w", err)
	}
	deposit, err := ParseYuanToFen(input.DepositYuan)
	if err != nil {
		return nil, fmt.Errorf("invalid room deposit: %w", err)
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = model.RoomStatusVacant
	}
	if !model.ValidRoomStatus(status) {
		return nil, fmt.Errorf("invalid room status")
	}
	if input.Area < 0 || input.Floor < 0 {
		return nil, fmt.Errorf("area and floor must be non-negative")
	}
	return &model.Room{
		ID:          id,
		RoomNo:      roomNo,
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Price:       price,
		Deposit:     deposit,
		Status:      status,
		Area:        input.Area,
		Floor:       input.Floor,
		Tags:        strings.TrimSpace(input.Tags),
	}, nil
}
