package entity

import (
	"time"

	"github.com/google/uuid"
)

type ProductType string

const (
	ProductElectronics ProductType = "электроника"
	ProductClothing    ProductType = "одежда"
	ProductShoes       ProductType = "обувь"
)

type Product struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	DateTime    time.Time   `json:"dateTime" db:"date_time"`
	Type        ProductType `json:"type" db:"type"`
	ReceptionID uuid.UUID   `json:"receptionId" db:"reception_id"`
}

func IsValidProductType(t ProductType) bool {
	switch t {
	case ProductElectronics, ProductClothing, ProductShoes:
		return true
	default:
		return false
	}
}
