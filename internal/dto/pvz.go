package dto

type PVZRequest struct {
	City string `json:"city" binding:"required"`
}

type PVZResponse struct {
	ID               string `json:"id"`
	RegistrationDate string `json:"registrationDate"`
	City             string `json:"city"`
}

type FullPVZResponse struct {
	PVZ        PVZResponse             `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

type ReceptionWithProducts struct {
	Reception ReceptionResponse `json:"reception"`
	Products  []ProductResponse `json:"products"`
}

type FullPVZQueryParams struct {
	StartDate string `form:"startDate" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	EndDate   string `form:"endDate" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Page      int    `form:"page" binding:"omitempty,min=1"`
	Limit     int    `form:"limit" binding:"omitempty,min=1,max=30"`
}
