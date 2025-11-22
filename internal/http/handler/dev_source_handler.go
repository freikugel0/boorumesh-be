package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/freikugel0/boorumesh-be/internal/repository"
	"github.com/freikugel0/boorumesh-be/internal/service"
)

type DevSourceHandler struct {
	svc service.DevSourceService
}

func NewDevSourceHandler(svc service.DevSourceService) *DevSourceHandler {
	return &DevSourceHandler{svc: svc}
}

func (h *DevSourceHandler) Create(c *gin.Context) {
	var in service.CreateSourceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid payload",
			"detail": err.Error(),
		})
		return
	}

	out, err := h.svc.CreateSource(c.Request.Context(), in)
	if err != nil {
		status := http.StatusInternalServerError

		switch err {
		case repository.ErrSourceExists:
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, out)
}

func (h *DevSourceHandler) GetSourceByCode(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "code is required",
		})
		return
	}

	ctx := c.Request.Context()

	src, err := h.svc.GetSourceByCode(ctx, code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSourceNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "source not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, src)
}
