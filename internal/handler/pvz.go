package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/pvz-service/internal/dto"
	"github.com/senyabanana/pvz-service/internal/entity"
	"github.com/senyabanana/pvz-service/internal/service"
)

type PVZHandler struct {
	service service.PVZOperations
	log     *logrus.Logger
}

func NewPVZHandler(service service.PVZOperations, log *logrus.Logger) *PVZHandler {
	return &PVZHandler{
		service: service,
		log:     log,
	}
}

// CreatePVZ godoc
// @Summary Create PVZ
// @Tags pvz
// @Description Создание ПВЗ
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.PVZRequest true "Город ПВЗ"
// @Success 201 {object} dto.PVZResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pvz [post]
func (h *PVZHandler) CreatePVZ(c *gin.Context) {
	var req dto.PVZRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnf("invalid create PVZ input: %v", err)
		dto.BadRequest(c, "invalid city")
		return
	}

	pvz, err := h.service.CreatePVZ(c.Request.Context(), req.City)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidCity) {
			h.log.Warnf("create pvz failed: unsupported city: %s", req.City)
			dto.BadRequest(c, "unsupported city")
			return
		}

		h.log.Errorf("failed to create PVZ: %v", err)
		dto.InternalError(c, "internal error")
		return
	}

	c.JSON(http.StatusCreated, dto.PVZResponse{
		ID:               pvz.ID.String(),
		RegistrationDate: pvz.RegistrationDate.Format(time.RFC3339),
		City:             string(pvz.City),
	})
}

// GetFullInfoPVZ godoc
// @Summary Get Full Info PVZ
// @Tags pvz
// @Description Получение информации о ПВЗ с приёмками и товарами
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param startDate query string false "Фильтрация по дате начала (RFC3339)"
// @Param endDate query string false "Фильтрация по дате окончания (RFC3339)"
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Лимит элементов на странице (по умолчанию 10)"
// @Success 200 {array} dto.FullPVZResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pvz [get]
func (h *PVZHandler) GetFullInfoPVZ(c *gin.Context) {
	var query dto.FullPVZQueryParams

	if err := c.ShouldBindQuery(&query); err != nil {
		dto.BadRequest(c, "invalid query parameters")
		return
	}

	startDate := parseQueryTime(query.StartDate, "startDate", c, h.log)
	if c.IsAborted() {
		return
	}
	endDate := parseQueryTime(query.EndDate, "endDate", c, h.log)
	if c.IsAborted() {
		return
	}

	page := query.Page
	if page == 0 {
		page = 1
	}
	limit := query.Limit
	if limit == 0 {
		limit = 10
	}

	pvzInfo, err := h.service.GetFullPVZInfo(c.Request.Context(), startDate, endDate, page, limit)
	if err != nil {
		h.log.Errorf("failed to get full PVZ info: %v", err)
		dto.InternalError(c, "failed to get PVZ list")
		return
	}

	c.JSON(http.StatusOK, convertToResponse(pvzInfo))
}

func parseQueryTime(raw string, field string, c *gin.Context, log *logrus.Logger) *time.Time {
	if raw == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		log.Warnf("invalid %s: %v", field, err)
		dto.BadRequest(c, fmt.Sprintf("invalid %s format", field))
		return nil
	}
	return &t
}

func convertToResponse(pvzInfo []entity.FullPVZInfo) []dto.FullPVZResponse {
	result := make([]dto.FullPVZResponse, 0, len(pvzInfo))

	for _, info := range pvzInfo {
		var receptions []dto.ReceptionWithProducts
		for _, rec := range info.Receptions {
			var products []dto.ProductResponse
			for _, p := range rec.Products {
				products = append(products, dto.ProductResponse{
					ID:          p.ID.String(),
					DateTime:    p.DateTime.Format(time.RFC3339),
					Type:        string(p.Type),
					ReceptionID: p.ReceptionID.String(),
				})
			}

			receptions = append(receptions, dto.ReceptionWithProducts{
				Reception: dto.ReceptionResponse{
					ID:       rec.Reception.ID.String(),
					DateTime: rec.Reception.DateTime.Format(time.RFC3339),
					PVZID:    rec.Reception.PVZID.String(),
					Status:   string(rec.Reception.Status),
				},
				Products: products,
			})
		}

		result = append(result, dto.FullPVZResponse{
			PVZ: dto.PVZResponse{
				ID:               info.PVZ.ID.String(),
				RegistrationDate: info.PVZ.RegistrationDate.Format(time.RFC3339),
				City:             string(info.PVZ.City),
			},
			Receptions: receptions,
		})
	}

	return result
}
