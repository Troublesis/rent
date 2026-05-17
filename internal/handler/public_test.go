package handler

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
)

func TestPublicRoomFilterStateUsesAvailableRoomsOnly(t *testing.T) {
	rooms := []model.Room{
		{RoomNo: "A101", Floor: 1, Bedrooms: 1, LivingRooms: 1, Bathrooms: 1},
		{RoomNo: "A202", Floor: 2, Bedrooms: 2, LivingRooms: 1, Bathrooms: 1},
		{RoomNo: "A203", Floor: 2, Bedrooms: 2, LivingRooms: 1, Bathrooms: 1},
	}

	state, filteredRooms := publicRoomFilterStateFor(rooms, publicRoomFilter{})

	if len(filteredRooms) != 3 {
		t.Fatalf("len(filteredRooms) = %d, want 3", len(filteredRooms))
	}
	floorLabels := optionLabels(state.Floors)
	if !reflect.DeepEqual(floorLabels, []string{"1层", "2层"}) {
		t.Fatalf("floorLabels = %#v, want 1层 and 2层", floorLabels)
	}
	layoutLabels := optionLabels(state.Layouts)
	if !reflect.DeepEqual(layoutLabels, []string{"1室1厅1卫", "2室1厅1卫"}) {
		t.Fatalf("layoutLabels = %#v, want unique available layouts", layoutLabels)
	}
}

func TestPublicRoomFilterStateFiltersByFloorAndLayout(t *testing.T) {
	rooms := []model.Room{
		{RoomNo: "A101", Floor: 1, Bedrooms: 1, LivingRooms: 1, Bathrooms: 1},
		{RoomNo: "A202", Floor: 2, Bedrooms: 2, LivingRooms: 1, Bathrooms: 1},
		{RoomNo: "A301", Floor: 3, Bedrooms: 2, LivingRooms: 1, Bathrooms: 1},
	}

	state, filteredRooms := publicRoomFilterStateFor(rooms, publicRoomFilter{Floor: "2", Layout: "2-1-1"})

	if !state.HasActive {
		t.Fatal("state.HasActive = false, want true")
	}
	if len(filteredRooms) != 1 || filteredRooms[0].RoomNo != "A202" {
		t.Fatalf("filteredRooms = %#v, want only A202", filteredRooms)
	}
}

func TestPublicRoomFilterStateIgnoresUnknownValues(t *testing.T) {
	rooms := []model.Room{{RoomNo: "A101", Floor: 1, Bedrooms: 1, LivingRooms: 1, Bathrooms: 1}}

	state, filteredRooms := publicRoomFilterStateFor(rooms, publicRoomFilter{Floor: "9", Layout: "9-9-9"})

	if state.HasActive {
		t.Fatal("state.HasActive = true, want false for unavailable filter values")
	}
	if len(filteredRooms) != 1 {
		t.Fatalf("len(filteredRooms) = %d, want 1", len(filteredRooms))
	}
}

func TestPublicRoomViewFromQueryDefaultsToList(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "missing", path: "/rooms", want: publicRoomViewList},
		{name: "empty", path: "/rooms?view=", want: publicRoomViewList},
		{name: "unknown", path: "/rooms?view=table", want: publicRoomViewList},
		{name: "grid", path: "/rooms?view=grid", want: publicRoomViewGrid},
		{name: "card", path: "/rooms?view=card", want: publicRoomViewCard},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newPublicTestContext(tt.path)
			if got := publicRoomViewFromQuery(c); got != tt.want {
				t.Fatalf("publicRoomViewFromQuery = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPublicRoomPageFromQueryDefaultsToFirstPage(t *testing.T) {
	tests := []struct {
		name string
		path string
		want int
	}{
		{name: "missing", path: "/rooms", want: 1},
		{name: "invalid", path: "/rooms?page=abc", want: 1},
		{name: "zero", path: "/rooms?page=0", want: 1},
		{name: "negative", path: "/rooms?page=-2", want: 1},
		{name: "valid", path: "/rooms?page=3", want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newPublicTestContext(tt.path)
			if got := publicRoomPageFromQuery(c); got != tt.want {
				t.Fatalf("publicRoomPageFromQuery = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPublicRoomFilterRepositoryFilter(t *testing.T) {
	filter := publicRoomFilter{Floor: "2", Layout: "3-1-2"}
	repoFilter := publicRoomRepositoryFilter(filter, 10, 20)

	if repoFilter.Floor != 2 || !repoFilter.HasFloor {
		t.Fatalf("Floor = %d with HasFloor %v, want 2 with true", repoFilter.Floor, repoFilter.HasFloor)
	}
	if repoFilter.Bedrooms != 3 || repoFilter.LivingRooms != 1 || repoFilter.Bathrooms != 2 {
		t.Fatalf("layout filter = %d-%d-%d, want 3-1-2", repoFilter.Bedrooms, repoFilter.LivingRooms, repoFilter.Bathrooms)
	}
	if !repoFilter.HasBedrooms || !repoFilter.HasLivingRooms || !repoFilter.HasBathrooms {
		t.Fatalf("layout filter flags = %v/%v/%v, want all true", repoFilter.HasBedrooms, repoFilter.HasLivingRooms, repoFilter.HasBathrooms)
	}
	if repoFilter.Limit != 10 || repoFilter.Offset != 20 {
		t.Fatalf("limit/offset = %d/%d, want 10/20", repoFilter.Limit, repoFilter.Offset)
	}
}

func TestPublicRoomURLPreservesFiltersAndView(t *testing.T) {
	filter := publicRoomFilter{Floor: "2", Layout: "3-1-2"}
	got := publicRoomsURL("card", filter, 3, true)
	want := "/?floor=2&layout=3-1-2&page=3&partial=1&view=card"

	if got != want {
		t.Fatalf("publicRoomsURL = %q, want %q", got, want)
	}
}

func optionLabels(options []publicRoomFilterOption) []string {
	labels := make([]string, 0, len(options))
	for _, option := range options {
		labels = append(labels, option.Label)
	}
	return labels
}

func newPublicTestContext(path string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, path, nil)
	return c
}
