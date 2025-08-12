package dto

type ReceptionRequest struct {
	PVZID string `json:"pvzId" binding:"required,uuid"`
}

type ReceptionResponse struct {
	ID       string `json:"id"`
	DateTime string `json:"dateTime"`
	PVZID    string `json:"pvzId"`
	Status   string `json:"status"`
}
