package service

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

type RoomInput struct {
	RoomNo        string
	Title         string
	Description   string
	PriceYuan     string
	RentType      string
	RentPriceYuan string
	PaymentTerms  string
	DepositYuan   string
	Status        string
	Area          int
	Floor         int
	Address       string
	Bedrooms      int
	LivingRooms   int
	Bathrooms     int
	Orientation   string
	Tags          string
}

type RoomService struct {
	roomRepo   *repository.RoomRepository
	tenantRepo *repository.TenantRepository
}

func NewRoomService(roomRepo *repository.RoomRepository, tenantRepo *repository.TenantRepository) *RoomService {
	return &RoomService{roomRepo: roomRepo, tenantRepo: tenantRepo}
}

func (s *RoomService) ListRooms(filter repository.RoomFilter) ([]model.Room, error) {
	rooms, err := s.roomRepo.ListRooms(filter)
	if err != nil {
		return nil, err
	}
	return roomsWithOrderedMedia(rooms), nil
}

func (s *RoomService) ListAvailableRooms(limit int, offset int) ([]model.Room, error) {
	rooms, err := s.roomRepo.ListPublicAvailableRooms(limit, offset)
	if err != nil {
		return nil, err
	}
	return roomsWithOrderedMedia(rooms), nil
}

func (s *RoomService) GetRoom(id uint) (*model.Room, error) {
	room, err := s.roomRepo.GetRoomWithMedia(id)
	if err != nil {
		return nil, err
	}
	return roomWithOrderedMedia(room), nil
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
	return s.GetRoom(id)
}

func (s *RoomService) DeleteRoom(id uint) error {
	room, err := s.roomRepo.GetRoom(id)
	if err != nil {
		return err
	}
	if room.Status == model.RoomStatusOccupied {
		return fmt.Errorf("已出租房源不能删除")
	}
	if _, err := s.tenantRepo.GetActiveTenantByRoomID(id); err == nil {
		return fmt.Errorf("存在在租租客的房源不能删除")
	} else if !repository.IsNotFound(err) {
		return err
	}
	return s.roomRepo.DeleteRoom(id)
}

func (s *RoomService) AddRoomMedia(roomID uint, rawURL string, mediaType string) error {
	if _, err := s.roomRepo.GetRoom(roomID); err != nil {
		return err
	}
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return fmt.Errorf("媒体地址不能为空")
	}
	if !model.ValidMediaType(mediaType) {
		return fmt.Errorf("媒体类型不正确")
	}
	if mediaType == model.MediaTypeVideoLink {
		if err := validateHTTPURL(trimmedURL); err != nil {
			return err
		}
	}
	media := &model.RoomMedia{RoomID: roomID, URL: trimmedURL, MediaType: mediaType}
	return s.roomRepo.AddRoomMedia(media)
}

func (s *RoomService) AddRoomVideoLink(roomID uint, rawURL string) error {
	return s.AddRoomMedia(roomID, rawURL, model.MediaTypeVideoLink)
}

func buildRoom(input RoomInput, id uint) (*model.Room, error) {
	roomNo, err := validateRoomNo(input.RoomNo)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("标题不能为空")
	}
	rentType := strings.TrimSpace(input.RentType)
	if rentType == "" {
		rentType = model.RentTypeMonthly
	}
	if !model.ValidRentType(rentType) {
		return nil, fmt.Errorf("租金类型不正确")
	}
	paymentTerms := strings.TrimSpace(input.PaymentTerms)
	if paymentTerms == "" {
		paymentTerms = model.PaymentTerms1M1D
	}
	if !model.ValidPaymentTerms(paymentTerms) {
		return nil, fmt.Errorf("付款方式不正确")
	}
	rentPriceInput := strings.TrimSpace(input.RentPriceYuan)
	if rentPriceInput == "" {
		rentPriceInput = strings.TrimSpace(input.PriceYuan)
	}
	rentPrice, err := ParseIntegerYuanToFen(rentPriceInput)
	if err != nil {
		return nil, fmt.Errorf("租金金额不正确：%w", err)
	}
	if rentPrice <= 0 {
		return nil, fmt.Errorf("租金金额需大于 0")
	}
	deposit, err := ParseIntegerYuanToFen(input.DepositYuan)
	if err != nil {
		return nil, fmt.Errorf("押金金额不正确：%w", err)
	}
	if deposit < 0 {
		return nil, fmt.Errorf("押金需大于或等于 0")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = model.RoomStatusVacant
	}
	if !model.ValidRoomStatus(status) {
		return nil, fmt.Errorf("房源状态不正确")
	}
	if err := validateIntegerRange(input.Area, 1, 9999, "面积"); err != nil {
		return nil, err
	}
	address := strings.TrimSpace(input.Address)
	if len([]rune(address)) < 5 {
		return nil, fmt.Errorf("地址至少需要 5 个字符")
	}
	if err := validateIntegerRange(input.Bedrooms, 0, 20, "室数量"); err != nil {
		return nil, err
	}
	if err := validateIntegerRange(input.LivingRooms, 0, 20, "厅数量"); err != nil {
		return nil, err
	}
	if err := validateIntegerRange(input.Bathrooms, 0, 20, "卫数量"); err != nil {
		return nil, err
	}
	orientation := strings.TrimSpace(input.Orientation)
	if orientation != "" && !validOrientation(orientation) {
		return nil, fmt.Errorf("房屋朝向不正确")
	}
	return &model.Room{
		ID:           id,
		RoomNo:       roomNo,
		Title:        title,
		Description:  strings.TrimSpace(input.Description),
		Price:        rentPrice,
		RentType:     rentType,
		RentPrice:    rentPrice,
		PaymentTerms: paymentTerms,
		Deposit:      deposit,
		Status:       status,
		Area:         input.Area,
		Floor:        input.Floor,
		Address:      address,
		Bedrooms:     input.Bedrooms,
		LivingRooms:  input.LivingRooms,
		Bathrooms:    input.Bathrooms,
		Orientation:  orientation,
		Tags:         strings.TrimSpace(input.Tags),
	}, nil
}

func validateHTTPURL(rawURL string) error {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("视频链接格式不正确")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("视频链接需以 http:// 或 https:// 开头")
	}
	return nil
}

func validOrientation(orientation string) bool {
	switch orientation {
	case "东", "南", "西", "北", "东南", "东北", "西南", "西北", "南北", "东西":
		return true
	default:
		return false
	}
}

func roomWithOrderedMedia(room *model.Room) *model.Room {
	updatedRoom := *room
	updatedRoom.Media = orderedRoomMedia(room.Media)
	return &updatedRoom
}

func roomsWithOrderedMedia(rooms []model.Room) []model.Room {
	orderedRooms := make([]model.Room, len(rooms))
	for i, room := range rooms {
		orderedRooms[i] = room
		orderedRooms[i].Media = orderedRoomMedia(room.Media)
	}
	return orderedRooms
}

func orderedRoomMedia(media []model.RoomMedia) []model.RoomMedia {
	orderedMedia := make([]model.RoomMedia, len(media))
	copy(orderedMedia, media)
	sort.SliceStable(orderedMedia, func(i int, j int) bool {
		leftRank := mediaRank(orderedMedia[i].MediaType)
		rightRank := mediaRank(orderedMedia[j].MediaType)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if !orderedMedia[i].CreatedAt.Equal(orderedMedia[j].CreatedAt) {
			return orderedMedia[i].CreatedAt.Before(orderedMedia[j].CreatedAt)
		}
		return orderedMedia[i].ID < orderedMedia[j].ID
	})
	return orderedMedia
}

func mediaRank(mediaType string) int {
	switch mediaType {
	case model.MediaTypeVideoLink:
		return 0
	case model.MediaTypeImage:
		return 1
	case model.MediaTypeVideo:
		return 2
	default:
		return 3
	}
}
