package entity

import (
	"time"

	"github.com/google/uuid"
)

type PVZCity string

const (
	CityMoscow PVZCity = "Москва"
	CitySPB    PVZCity = "Санкт-Петербург"
	CityKazan  PVZCity = "Казань"
)

type PVZ struct {
	ID               uuid.UUID `json:"id" db:"id"`
	RegistrationDate time.Time `json:"registrationDate" db:"registration_date"`
	City             PVZCity   `json:"city" db:"city"`
}

func IsValidCity(city string) bool {
	switch PVZCity(city) {
	case CityMoscow, CitySPB, CityKazan:
		return true
	default:
		return false
	}
}

type FullPVZInfo struct {
	PVZ        PVZ                     `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

type ReceptionWithProducts struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
}
