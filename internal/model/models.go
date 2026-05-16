package model

import "time"

const (
	RoomStatusVacant      = "vacant"
	RoomStatusOccupied    = "occupied"
	RoomStatusMaintenance = "maintenance"
)

const (
	TenantStatusActive   = "active"
	TenantStatusCheckout = "checkout"
)

const (
	PaymentTypeRent        = "rent"
	PaymentTypeWater       = "water"
	PaymentTypeElectricity = "electricity"
	PaymentTypeOther       = "other"
)

const (
	MediaTypeImage = "image"
	MediaTypeVideo = "video"
)

type Room struct {
	ID          uint   `gorm:"primarykey"`
	RoomNo      string `gorm:"uniqueIndex;not null"`
	Title       string `gorm:"not null"`
	Description string
	Price       int
	Deposit     int
	Status      string `gorm:"default:vacant;index"`
	Area        int
	Floor       int
	Tags        string
	Media       []RoomMedia `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RoomMedia struct {
	ID        uint   `gorm:"primarykey"`
	RoomID    uint   `gorm:"not null;index"`
	URL       string `gorm:"not null"`
	MediaType string `gorm:"not null"`
	CreatedAt time.Time
}

type Tenant struct {
	ID               uint   `gorm:"primarykey"`
	Name             string `gorm:"not null"`
	Phone            string `gorm:"not null"`
	EmergencyContact string
	RoomID           uint `gorm:"not null;index"`
	Room             Room
	CheckinDate      time.Time
	CheckoutDate     *time.Time
	RentPrice        int
	Deposit          int
	Status           string `gorm:"default:active;index"`
	Payments         []Payment
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Payment struct {
	ID        uint `gorm:"primarykey"`
	TenantID  uint `gorm:"not null;index"`
	Tenant    Tenant
	Amount    int
	Type      string `gorm:"not null;index"`
	Paid      bool   `gorm:"default:false;index"`
	PayDate   time.Time
	Note      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AppSetting struct {
	ID        uint   `gorm:"primarykey"`
	Key       string `gorm:"uniqueIndex;not null"`
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ValidRoomStatus(status string) bool {
	switch status {
	case RoomStatusVacant, RoomStatusOccupied, RoomStatusMaintenance:
		return true
	default:
		return false
	}
}

func ValidTenantStatus(status string) bool {
	switch status {
	case TenantStatusActive, TenantStatusCheckout:
		return true
	default:
		return false
	}
}

func ValidPaymentType(paymentType string) bool {
	switch paymentType {
	case PaymentTypeRent, PaymentTypeWater, PaymentTypeElectricity, PaymentTypeOther:
		return true
	default:
		return false
	}
}
