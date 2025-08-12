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

type ProductHandler struct {
	service service.ProductOperations
	log     *logrus.Logger
}

func NewProductHandler(service service.ProductOperations, log *logrus.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		log:     log,
	}
}

// AddProduct godoc
// @Summary Add Product
// @Tags product
// @Description Добавление товара в текущую приёмку ПВЗ
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body dto.ProductRequest true "Product payload"
// @Success 201 {object} dto.ProductResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products [post]
func (h *ProductHandler) AddProduct(c *gin.Context) {
	var req dto.ProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnf("invalid product input: %v", err)
		dto.BadRequest(c, "invalid request body")
		return
	}

	pvzID, err := uuid.Parse(req.PVZID)
	if err != nil {
		h.log.Warnf("invalid UUID format: %v", err)
		dto.BadRequest(c, "invalid UUID format")
		return
	}

	product, err := h.service.AddProduct(c.Request.Context(), pvzID, entity.ProductType(req.Type))
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNoActiveReception):
			dto.BadRequest(c, "no open reception for this PVZ")
			return
		case errors.Is(err, entity.ErrInvalidProductType):
			dto.BadRequest(c, "invalid product type")
			return
		default:
			dto.InternalError(c, "failed to add product")
			return
		}
	}

	c.JSON(http.StatusCreated, dto.ProductResponse{
		ID:          product.ID.String(),
		DateTime:    product.DateTime.Format(time.RFC3339),
		Type:        string(product.Type),
		ReceptionID: product.ReceptionID.String(),
	})
}

// DeleteLastProduct godoc
// @Summary Delete Last Product
// @Tags product
// @Description Удаление последнего добавленного товара из текущей приёмки
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param pvzId path string true "PVZ ID"
// @Success 200
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pvz/{pvzId}/delete_last_product [post]
func (h *ProductHandler) DeleteLastProduct(c *gin.Context) {
	pvzIDParam := c.Param("pvzId")
	pvzID, err := uuid.Parse(pvzIDParam)
	if err != nil {
		h.log.Warnf("invalid pvzId: %s", pvzIDParam)
		dto.BadRequest(c, "invalid pvzId")
		return
	}

	if err := h.service.DeleteLastProduct(c.Request.Context(), pvzID); err != nil {
		switch {
		case errors.Is(err, entity.ErrNoOpenReception):
			dto.BadRequest(c, "no open reception for this PVZ")
			return
		case errors.Is(err, entity.ErrNoProductsToDelete):
			dto.BadRequest(c, "no products to delete")
			return
		default:
			dto.InternalError(c, "failed to delete product")
			return
		}
	}

	c.Status(http.StatusOK)
}
