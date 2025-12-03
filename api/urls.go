package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/datmaithanh/URL-Shortener-Service/config"
	db "github.com/datmaithanh/URL-Shortener-Service/db/sqlc"
	"github.com/datmaithanh/URL-Shortener-Service/utils"
	"github.com/gin-gonic/gin"
)

type urlsRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	Title       string `json:"title"`
}

type urlsResponse struct {
	Code        string    `json:"code"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	Title       string    `json:"title,omitempty"`
	Clicks      int64     `json:"clicks"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
}

func (server *Server) newUrlsResponse(ctx *gin.Context) {
	var req urlsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	
	payload := db.CreateUrlParams{
		OriginalUrl: req.OriginalURL,
		Title:       req.Title,
		ExpiresAt:   time.Now().Add(config.URL_EXPIRE_DURATION),
	}
	url, err := server.store.CreateUrl(ctx, payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	code := utils.EncodeBase62(url.ID)
	shortUrl := fmt.Sprintf("%s/%s", config.DOMAIN_NAME, code)
	_, err = server.store.UpdateCodeUrl(ctx, db.UpdateCodeUrlParams{
		ID:       url.ID,
		Code:     sql.NullString{String: code, Valid: true},
		ShortUrl: sql.NullString{String: shortUrl, Valid: true},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := urlsResponse{
		Code:        code,
		ShortURL:    shortUrl,
		OriginalURL: url.OriginalUrl,
		Title:       url.Title,
		Clicks:      url.Clicks,
		CreatedAt:   url.CreatedAt,
		ExpiresAt:   url.ExpiresAt,
	}

	ctx.JSON(http.StatusOK, response)
}
