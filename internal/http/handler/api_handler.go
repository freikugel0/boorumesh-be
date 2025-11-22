package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/freikugel0/boorumesh-be/internal/service"
	"github.com/gin-gonic/gin"
)

type ApiHandler struct {
	fetchService service.SourceFetchService
}

func NewApiHandler(fetchService service.SourceFetchService) *ApiHandler {
	return &ApiHandler{fetchService: fetchService}
}

func (h *ApiHandler) GetImagesBySource(c *gin.Context) {
	code := strings.TrimSpace(c.Param("source"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "source is required",
		})
		return
	}

	// tags=tag1 tag2 tag3 (space separated)
	tagsRaw := strings.TrimSpace(c.Query("tags"))
	var tags []string
	if tagsRaw != "" {
		tags = strings.Fields(tagsRaw)
	}

	page := 1
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 0
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Option for disabling tags suffix
	raw := c.Query("raw") == "1"

	ctx := c.Request.Context()

	images, err := h.fetchService.FetchBySource(ctx, code, tags, page, limit, raw)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSourceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
		case errors.Is(err, service.ErrSourceDisabled):
			c.JSON(http.StatusBadRequest, gin.H{"error": "source is disabled"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "failed to fetch from source",
				"detail": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, images)
}
