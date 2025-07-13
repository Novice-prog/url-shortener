package handler

import (
	"net/http"
	"net/url"
	"time"
	"url-shortener1/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	urlService *service.URLService
}

type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

type StatResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	VisitCount  int64  `json:"visit_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewHandler(urlService *service.URLService) *Handler {
	return &Handler{
		urlService: urlService,
	}
}

// validateURL checks if the URL is valid
func validateURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// POST /api/shorten
func (h *Handler) ShortenURL(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request format",
		})
		return
	}

	// Validate URL
	if !validateURL(req.URL) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid URL format. URL must start with http:// or https://",
		})
		return
	}

	// Create short URL
	sh, err := h.urlService.Create(c.Request.Context(), req.URL)
	if err != nil {
		if err == service.ErrExists {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: "URL already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to create short URL",
		})
		return
	}

	c.JSON(http.StatusCreated, ShortenResponse{
		ShortURL: sh.ShortURL,
	})
}

// GET /:short
func (h *Handler) Redirect(c *gin.Context) {
	short := c.Param("short")

	if short == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Short URL is required",
		})
		return
	}

	// Resolve short URL
	originalURL, err := h.urlService.Resolve(c.Request.Context(), short)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "URL not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to resolve URL",
		})
		return
	}

	// Redirect to original URL
	c.Redirect(http.StatusFound, originalURL)
}

func (h *Handler) Stat(c *gin.Context) {
	short := c.Param("short")
	stat, err := h.urlService.GetStat(c.Request.Context(), short)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "URL not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get stat",
		})
		return
	}
	c.JSON(http.StatusOK, StatResponse{
		OriginalURL: stat.OriginalURL,
		ShortURL:    stat.ShortURL,
		VisitCount:  stat.Visits,
		CreatedAt:   stat.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   stat.UpdatedAt.Format(time.RFC3339),
	})
}
