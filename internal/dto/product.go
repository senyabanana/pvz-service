package dto

type ProductRequest struct {
	Type  string `json:"type" binding:"required"`
	PVZID string `json:"pvzId" binding:"required,uuid"`
}

type ProductResponse struct {
	ID          string `json:"id"`
	DateTime    string `json:"dateTime"`
	Type        string `json:"type"`
	ReceptionID string `json:"receptionId"`
}
