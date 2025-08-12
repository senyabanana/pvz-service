package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/pvz-service/internal/dto"
	"github.com/senyabanana/pvz-service/internal/entity"
	"github.com/senyabanana/pvz-service/internal/service"
)

type ReceptionHandler struct {
	service service.ReceptionOperations
	log     *logrus.Logger
}

func NewReceptionHandler(service service.ReceptionOperations, log *logrus.Logger) *ReceptionHandler {
	return &ReceptionHandler{service: service, log: log}
}

// CreateReception godoc
// @Summary Create Reception
// @Tags reception
// @Description Создание приёмки для ПВЗ
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body dto.ReceptionRequest true "PVZ ID"
// @Success 201 {object} dto.ReceptionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /receptions [post]
func (h *ReceptionHandler) CreateReception(c *gin.Context) {
	var req dto.ReceptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnf("invalid reception input: %v", err)
		dto.BadRequest(c, "invalid pvzId")
		return
	}

	pvzID, err := uuid.Parse(req.PVZID)
	if err != nil {
		h.log.Warnf("invalid UUID format: %v", err)
		dto.BadRequest(c, "invalid UUID format")
		return
	}

	reception, err := h.service.CreateReception(c.Request.Context(), pvzID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrReceptionAlreadyExists):
			dto.BadRequest(c, "there is already an open reception for this PVZ")
			return
		case errors.Is(err, entity.ErrPVZNotFound):
			dto.NotFound(c, "pvz not found")
			return
		default:
			dto.InternalError(c, "failed to create reception")
			return
		}
	}

	c.JSON(http.StatusCreated, dto.ReceptionResponse{
		ID:       reception.ID.String(),
		DateTime: reception.DateTime.Format(time.RFC3339),
		PVZID:    reception.PVZID.String(),
		Status:   string(reception.Status),
	})
}

// CloseLastReception godoc
// @Summary Close Last Reception
// @Tags reception
// @Description Закрытие последней открытой приёмки у ПВЗ
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param pvzId path string true "PVZ ID"
// @Success 200 {object} dto.ReceptionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pvz/{pvzId}/close_last_reception [post]
func (h *ReceptionHandler) CloseLastReception(c *gin.Context) {
	pvzIDParam := c.Param("pvzId")
	pvzID, err := uuid.Parse(pvzIDParam)
	if err != nil {
		h.log.Warnf("invalid pvzId: %s", pvzIDParam)
		dto.BadRequest(c, "invalid pvzId")
		return
	}

	reception, err := h.service.CloseLastReception(c.Request.Context(), pvzID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNoOpenReception):
			dto.BadRequest(c, "no open reception found for this PVZ")
			return
		case errors.Is(err, entity.ErrReceptionAlreadyClosed):
			dto.BadRequest(c, "reception is already closed")
			return
		default:
			dto.InternalError(c, "failed to close reception")
			return
		}
	}

	c.JSON(http.StatusOK, dto.ReceptionResponse{
		ID:       reception.ID.String(),
		DateTime: reception.DateTime.Format(time.RFC3339),
		PVZID:    reception.PVZID.String(),
		Status:   string(reception.Status),
	})
}
