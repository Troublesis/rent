package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
	"github.com/troublesis/rent/internal/service"
)

type PublicHandler struct {
	renderer        Renderer
	roomService     *service.RoomService
	settingsService *service.SettingsService
}

type publicRoomFilter struct {
	Floor  string
	Layout string
}

type publicRoomFilterOption struct {
	Value string
	Label string
}

const (
	publicRoomViewList = "list"
	publicRoomViewGrid = "grid"
	publicRoomViewCard = "card"
)

type publicRoomFilterState struct {
	Floor     string
	Layout    string
	Floors    []publicRoomFilterOption
	Layouts   []publicRoomFilterOption
	HasActive bool
}

func NewPublicHandler(renderer Renderer, roomService *service.RoomService, settingsService *service.SettingsService) *PublicHandler {
	return &PublicHandler{renderer: renderer, roomService: roomService, settingsService: settingsService}
}

func (h *PublicHandler) Index(c *gin.Context) {
	data, err := h.publicRoomBrowserData(c, "/")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if c.Query("partial") == "1" {
		h.renderer.RenderPartial(c, http.StatusOK, "public/index.html", "public_room_page", data)
		return
	}
	settings, err := h.settingsService.GetSettings()
	if err != nil {
		c.String(http.StatusInternalServerError, "读取联系信息失败")
		return
	}
	data["Title"] = "安心租房"
	data["Settings"] = settings
	h.renderer.Render(c, http.StatusOK, "public_base.html", "public/index.html", data)
}

func (h *PublicHandler) Rooms(c *gin.Context) {
	target := "/"
	if c.Request.URL.RawQuery != "" {
		target += "?" + c.Request.URL.RawQuery
	}
	c.Redirect(http.StatusFound, target)
}

func (h *PublicHandler) publicRoomBrowserData(c *gin.Context, action string) (gin.H, error) {
	viewMode := publicRoomViewFromQuery(c)
	page := publicRoomPageFromQuery(c)
	pageSize := publicRoomPageSize(viewMode)
	facetRooms, err := h.roomService.ListAvailableRoomFacets()
	if err != nil {
		return nil, fmt.Errorf("读取房源失败")
	}
	filterState, _ := publicRoomFilterStateFor(facetRooms, publicRoomFilterFromQuery(c))
	filter := publicRoomFilter{Floor: filterState.Floor, Layout: filterState.Layout}
	repoFilter := publicRoomRepositoryFilter(filter, pageSize+1, (page-1)*pageSize)
	rooms, err := h.roomService.ListAvailableRoomsFiltered(repoFilter)
	if err != nil {
		return nil, fmt.Errorf("读取房源失败")
	}
	visibleRooms, hasMore := publicRoomPageRooms(rooms, pageSize)
	return gin.H{
		"Rooms":           visibleRooms,
		"Filter":          filterState,
		"FilterAction":    action,
		"ClearFilterURL":  publicRoomsURL(viewMode, publicRoomFilter{}, 1, false),
		"ViewMode":        viewMode,
		"ListViewURL":     publicRoomsURL(publicRoomViewList, filter, 1, false),
		"GridViewURL":     publicRoomsURL(publicRoomViewGrid, filter, 1, false),
		"CardViewURL":     publicRoomsURL(publicRoomViewCard, filter, 1, false),
		"HasMore":         hasMore,
		"NextPageURL":     publicRoomsURL(viewMode, filter, page+1, true),
		"NextPageFullURL": publicRoomsURL(viewMode, filter, page+1, false),
	}, nil
}

func (h *PublicHandler) RoomDetail(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	room, err := h.roomService.GetRoom(id)
	if err != nil || room.Status != model.RoomStatusVacant {
		c.String(http.StatusNotFound, "房源不存在或暂不可租")
		return
	}
	settings, err := h.settingsService.GetSettings()
	if err != nil {
		c.String(http.StatusInternalServerError, "读取联系信息失败")
		return
	}
	h.renderer.Render(c, http.StatusOK, "public_base.html", "public/room_detail.html", gin.H{
		"Title":    room.Title,
		"Room":     room,
		"Settings": settings,
	})
}

func publicRoomFilterFromQuery(c *gin.Context) publicRoomFilter {
	return publicRoomFilter{Floor: c.Query("floor"), Layout: c.Query("layout")}
}

func publicRoomViewFromQuery(c *gin.Context) string {
	switch c.Query("view") {
	case publicRoomViewGrid:
		return publicRoomViewGrid
	case publicRoomViewCard:
		return publicRoomViewCard
	default:
		return publicRoomViewList
	}
}

func publicRoomPageFromQuery(c *gin.Context) int {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		return 1
	}
	return page
}

func publicRoomPageSize(viewMode string) int {
	switch viewMode {
	case publicRoomViewGrid:
		return 16
	case publicRoomViewCard:
		return 8
	default:
		return 20
	}
}

func publicRoomPageRooms(rooms []model.Room, pageSize int) ([]model.Room, bool) {
	if pageSize <= 0 || len(rooms) <= pageSize {
		return append([]model.Room(nil), rooms...), false
	}
	return append([]model.Room(nil), rooms[:pageSize]...), true
}

func publicRoomRepositoryFilter(filter publicRoomFilter, limit int, offset int) repository.RoomFilter {
	roomFilter := repository.RoomFilter{Limit: limit, Offset: offset}
	if floor, err := strconv.Atoi(filter.Floor); err == nil {
		roomFilter.Floor = floor
		roomFilter.HasFloor = true
	}
	var bedrooms int
	var livingRooms int
	var bathrooms int
	if _, err := fmt.Sscanf(filter.Layout, "%d-%d-%d", &bedrooms, &livingRooms, &bathrooms); err == nil {
		roomFilter.Bedrooms = bedrooms
		roomFilter.HasBedrooms = true
		roomFilter.LivingRooms = livingRooms
		roomFilter.HasLivingRooms = true
		roomFilter.Bathrooms = bathrooms
		roomFilter.HasBathrooms = true
	}
	return roomFilter
}

func publicRoomsURL(viewMode string, filter publicRoomFilter, page int, partial bool) string {
	values := url.Values{}
	if filter.Floor != "" {
		values.Set("floor", filter.Floor)
	}
	if filter.Layout != "" {
		values.Set("layout", filter.Layout)
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}
	if partial {
		values.Set("partial", "1")
	}
	values.Set("view", viewMode)
	query := values.Encode()
	if query == "" {
		return "/"
	}
	return "/?" + query
}

func publicRoomFilterStateFor(rooms []model.Room, rawFilter publicRoomFilter) (publicRoomFilterState, []model.Room) {
	floors, layouts := publicRoomFilterOptions(rooms)
	filter := normalizedPublicRoomFilter(rawFilter, floors, layouts)
	state := publicRoomFilterState{
		Floor:     filter.Floor,
		Layout:    filter.Layout,
		Floors:    floors,
		Layouts:   layouts,
		HasActive: filter.Floor != "" || filter.Layout != "",
	}
	return state, filterPublicRooms(rooms, filter)
}

func publicRoomFilterOptions(rooms []model.Room) ([]publicRoomFilterOption, []publicRoomFilterOption) {
	floorValues := map[int]struct{}{}
	layoutValues := map[string]publicRoomFilterOption{}
	for _, room := range rooms {
		floorValues[room.Floor] = struct{}{}
		layoutValue := publicRoomLayoutValue(room)
		layoutValues[layoutValue] = publicRoomFilterOption{Value: layoutValue, Label: publicRoomLayoutLabel(room)}
	}
	floors := publicFloorOptions(floorValues)
	layouts := make([]publicRoomFilterOption, 0, len(layoutValues))
	for _, option := range layoutValues {
		layouts = append(layouts, option)
	}
	sort.Slice(layouts, func(i int, j int) bool {
		left := publicRoomLayoutSortKey(layouts[i].Value)
		right := publicRoomLayoutSortKey(layouts[j].Value)
		return left < right
	})
	return floors, layouts
}

func publicFloorOptions(values map[int]struct{}) []publicRoomFilterOption {
	floors := make([]int, 0, len(values))
	for floor := range values {
		floors = append(floors, floor)
	}
	sort.Ints(floors)
	options := make([]publicRoomFilterOption, 0, len(floors))
	for _, floor := range floors {
		value := strconv.Itoa(floor)
		options = append(options, publicRoomFilterOption{Value: value, Label: value + "层"})
	}
	return options
}

func normalizedPublicRoomFilter(rawFilter publicRoomFilter, floors []publicRoomFilterOption, layouts []publicRoomFilterOption) publicRoomFilter {
	return publicRoomFilter{
		Floor:  publicFilterValue(rawFilter.Floor, floors),
		Layout: publicFilterValue(rawFilter.Layout, layouts),
	}
}

func publicFilterValue(value string, options []publicRoomFilterOption) string {
	for _, option := range options {
		if option.Value == value {
			return value
		}
	}
	return ""
}

func filterPublicRooms(rooms []model.Room, filter publicRoomFilter) []model.Room {
	filteredRooms := make([]model.Room, 0, len(rooms))
	for _, room := range rooms {
		if filter.Floor != "" && strconv.Itoa(room.Floor) != filter.Floor {
			continue
		}
		if filter.Layout != "" && publicRoomLayoutValue(room) != filter.Layout {
			continue
		}
		filteredRooms = append(filteredRooms, room)
	}
	return filteredRooms
}

func limitPublicRooms(rooms []model.Room, limit int) []model.Room {
	if limit <= 0 || len(rooms) <= limit {
		return append([]model.Room(nil), rooms...)
	}
	return append([]model.Room(nil), rooms[:limit]...)
}

func publicRoomLayoutValue(room model.Room) string {
	return fmt.Sprintf("%d-%d-%d", room.Bedrooms, room.LivingRooms, room.Bathrooms)
}

func publicRoomLayoutLabel(room model.Room) string {
	return fmt.Sprintf("%d室%d厅%d卫", room.Bedrooms, room.LivingRooms, room.Bathrooms)
}

func publicRoomLayoutSortKey(value string) string {
	var bedrooms int
	var livingRooms int
	var bathrooms int
	if _, err := fmt.Sscanf(value, "%d-%d-%d", &bedrooms, &livingRooms, &bathrooms); err != nil {
		return value
	}
	return fmt.Sprintf("%02d-%02d-%02d", bedrooms, livingRooms, bathrooms)
}
