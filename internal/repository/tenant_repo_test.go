package repository

import (
	"testing"
	"time"

	"github.com/troublesis/rent/internal/model"
	"gorm.io/gorm"
)

func TestTenantRepositoryListTenantsSearchesTenantAndRoom(t *testing.T) {
	db := newTestDB(t)
	repo := NewTenantRepository(db)
	createTenantRepoTenant(t, db, "T101", "向阳主卧", "张三", "13800003001", model.TenantStatusActive, 100000, time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local))
	createTenantRepoTenant(t, db, "T202", "安静次卧", "李四", "13800003002", model.TenantStatusActive, 200000, time.Date(2026, time.May, 2, 0, 0, 0, 0, time.Local))

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{name: "name", query: "张三", want: "张三"},
		{name: "phone", query: "3002", want: "李四"},
		{name: "room number", query: "T101", want: "张三"},
		{name: "room title", query: "安静", want: "李四"},
		{name: "search option label", query: "张三 - T101 - 13800003001", want: "张三"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenants, err := repo.ListTenants(TenantFilter{Query: tt.query})
			if err != nil {
				t.Fatalf("ListTenants returned error: %v", err)
			}
			if len(tenants) != 1 || tenants[0].Name != tt.want {
				t.Fatalf("tenants = %#v, want only %s", tenants, tt.want)
			}
		})
	}
}

func TestTenantRepositoryListTenantsFiltersAndSorts(t *testing.T) {
	db := newTestDB(t)
	repo := NewTenantRepository(db)
	createTenantRepoTenant(t, db, "T301", "一号房", "王五", "13800003003", model.TenantStatusActive, 300000, time.Date(2026, time.May, 3, 0, 0, 0, 0, time.Local))
	createTenantRepoTenant(t, db, "T102", "二号房", "赵六", "13800003004", model.TenantStatusCheckout, 100000, time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local))
	createTenantRepoTenant(t, db, "T203", "三号房", "陈七", "13800003005", model.TenantStatusActive, 200000, time.Date(2026, time.May, 2, 0, 0, 0, 0, time.Local))

	activeTenants, err := repo.ListTenants(TenantFilter{Status: model.TenantStatusActive, SortBy: "rent_price", SortDir: "asc"})
	if err != nil {
		t.Fatalf("ListTenants active returned error: %v", err)
	}
	if len(activeTenants) != 2 || activeTenants[0].Name != "陈七" || activeTenants[1].Name != "王五" {
		t.Fatalf("active tenants = %#v, want 陈七 then 王五", activeTenants)
	}

	roomSorted, err := repo.ListTenants(TenantFilter{SortBy: "room", SortDir: "asc"})
	if err != nil {
		t.Fatalf("ListTenants room sort returned error: %v", err)
	}
	if len(roomSorted) != 3 || roomSorted[0].Room.RoomNo != "T102" || roomSorted[2].Room.RoomNo != "T301" {
		t.Fatalf("roomSorted = %#v, want room number ascending", roomSorted)
	}
}

func TestTenantRepositoryListTenantsByRoomID(t *testing.T) {
	db := newTestDB(t)
	repo := NewTenantRepository(db)

	room := model.Room{RoomNo: "R501", Title: "测试房源", RentType: model.RentTypeMonthly, RentPrice: 100000, PaymentTerms: model.PaymentTerms1M1D, Deposit: 100000, Status: model.RoomStatusOccupied}
	if err := db.Create(&room).Error; err != nil {
		t.Fatalf("create room: %v", err)
	}
	otherRoom := model.Room{RoomNo: "R502", Title: "其他房源", RentType: model.RentTypeMonthly, RentPrice: 100000, PaymentTerms: model.PaymentTerms1M1D, Deposit: 100000, Status: model.RoomStatusOccupied}
	if err := db.Create(&otherRoom).Error; err != nil {
		t.Fatalf("create other room: %v", err)
	}

	earlier := model.Tenant{Name: "早期租客", Phone: "13800009001", RoomID: room.ID, CheckinDate: time.Date(2026, time.March, 1, 0, 0, 0, 0, time.Local), RentPrice: 100000, RentType: model.RentTypeMonthly, PaymentTerms: model.PaymentTerms1M1D, Status: model.TenantStatusCheckout}
	if err := db.Create(&earlier).Error; err != nil {
		t.Fatalf("create earlier tenant: %v", err)
	}
	latest := model.Tenant{Name: "最近租客", Phone: "13800009002", RoomID: room.ID, CheckinDate: time.Date(2026, time.May, 10, 0, 0, 0, 0, time.Local), RentPrice: 100000, RentType: model.RentTypeMonthly, PaymentTerms: model.PaymentTerms1M1D, Status: model.TenantStatusActive}
	if err := db.Create(&latest).Error; err != nil {
		t.Fatalf("create latest tenant: %v", err)
	}
	unrelated := model.Tenant{Name: "无关租客", Phone: "13800009003", RoomID: otherRoom.ID, CheckinDate: time.Date(2026, time.April, 1, 0, 0, 0, 0, time.Local), RentPrice: 100000, RentType: model.RentTypeMonthly, PaymentTerms: model.PaymentTerms1M1D, Status: model.TenantStatusActive}
	if err := db.Create(&unrelated).Error; err != nil {
		t.Fatalf("create unrelated tenant: %v", err)
	}

	tenants, err := repo.ListTenantsByRoomID(room.ID)
	if err != nil {
		t.Fatalf("ListTenantsByRoomID returned error: %v", err)
	}
	if len(tenants) != 2 {
		t.Fatalf("got %d tenants, want 2", len(tenants))
	}
	if tenants[0].Name != "最近租客" || tenants[1].Name != "早期租客" {
		t.Fatalf("tenants order = [%s, %s], want [最近租客, 早期租客]", tenants[0].Name, tenants[1].Name)
	}

	empty, err := repo.ListTenantsByRoomID(9999)
	if err != nil {
		t.Fatalf("ListTenantsByRoomID empty returned error: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected no tenants for unknown room, got %d", len(empty))
	}
}

func createTenantRepoTenant(t *testing.T, db interface {
	Create(value interface{}) *gorm.DB
}, roomNo string, roomTitle string, name string, phone string, status string, rentPrice int, checkinDate time.Time) model.Tenant {
	t.Helper()
	room := model.Room{RoomNo: roomNo, Title: roomTitle, RentType: model.RentTypeMonthly, RentPrice: rentPrice, PaymentTerms: model.PaymentTerms1M1D, Deposit: rentPrice, Status: model.RoomStatusOccupied}
	if err := db.Create(&room).Error; err != nil {
		t.Fatalf("create room: %v", err)
	}
	tenant := model.Tenant{Name: name, Phone: phone, RoomID: room.ID, CheckinDate: checkinDate, RentPrice: rentPrice, RentType: model.RentTypeMonthly, PaymentTerms: model.PaymentTerms1M1D, Status: status}
	if err := db.Create(&tenant).Error; err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	return tenant
}
