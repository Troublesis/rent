package handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
)

func TestPaymentToAPIItemIncludesCanonicalFields(t *testing.T) {
	payment := model.Payment{
		ID:       7,
		TenantID: 3,
		Amount:   180000,
		Type:     model.PaymentTypeRent,
		Paid:     false,
		PayDate:  time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local),
		Tenant: model.Tenant{
			ID:        3,
			Name:      "张小明",
			Phone:     "13800000000",
			Status:    model.TenantStatusCheckout,
			RentPrice: 180000,
			RentType:  model.RentTypeMonthly,
			Room:      model.Room{RoomNo: "A101"},
		},
	}

	item := paymentToAPIItem(payment, time.Date(2026, time.May, 17, 0, 0, 0, 0, time.Local))
	if item.PaymentID != payment.ID {
		t.Fatalf("PaymentID = %d, want %d", item.PaymentID, payment.ID)
	}
	if item.Room != "A101" || item.RoomNo != "A101" {
		t.Fatalf("room fields = %q/%q, want A101", item.Room, item.RoomNo)
	}
	if item.TenantStatus != model.TenantStatusCheckout {
		t.Fatalf("TenantStatus = %q, want checkout", item.TenantStatus)
	}
	if item.PaymentStatus != "unpaid" {
		t.Fatalf("PaymentStatus = %q, want unpaid", item.PaymentStatus)
	}
	if item.StatusLabel != "未付款" {
		t.Fatalf("StatusLabel = %q, want 未付款", item.StatusLabel)
	}
}

func TestPaymentFilterFromQueryDefaultsToNormalRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		query     string
		wantNil   bool
		wantValue bool
	}{
		{name: "missing", query: "", wantValue: false},
		{name: "normal", query: "?excluded=false", wantValue: false},
		{name: "excluded", query: "?excluded=true", wantValue: true},
		{name: "all", query: "?excluded=all", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("GET", "/api/payments"+tt.query, nil)

			filter := paymentFilterFromQuery(ctx)
			if tt.wantNil {
				if filter.Excluded != nil {
					t.Fatalf("Excluded = %v, want nil", *filter.Excluded)
				}
				return
			}
			if filter.Excluded == nil {
				t.Fatal("Excluded = nil, want bool pointer")
			}
			if *filter.Excluded != tt.wantValue {
				t.Fatalf("Excluded = %v, want %v", *filter.Excluded, tt.wantValue)
			}
		})
	}
}
