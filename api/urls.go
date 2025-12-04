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

	existingUrl, err := server.store.GetUrlByOriginalUrl(ctx, req.OriginalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			payload := db.CreateUrlTxParams{
				CreateUrlParams: db.CreateUrlParams{
					OriginalUrl: req.OriginalURL,
					Title:       req.Title,
					ExpiresAt:   time.Now().Add(config.URL_EXPIRE_DURATION),
				},
				AfterCreate: func(q *db.Queries, url *db.Url) (db.Url, error) {
					code := utils.EncodeBase62(url.ID)
					shortUrl := fmt.Sprintf("%s/%s", config.DOMAIN_NAME, code)

					result, err := q.UpdateCodeUrl(ctx, db.UpdateCodeUrlParams{
						ID:       url.ID,
						Code:     sql.NullString{String: code, Valid: true},
						ShortUrl: sql.NullString{String: shortUrl, Valid: true},
					})
					if err != nil {
						return db.Url{}, err
					}
					return result, nil
				},
			}

			url, err := server.store.CreateUrlTx(ctx, payload)
			fmt.Println("DEBUG URL: ", url)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}

			response := urlsResponse{
				Code:        url.Url.Code.String,
				ShortURL:    url.Url.ShortUrl.String,
				OriginalURL: url.Url.OriginalUrl,
				Title:       url.Url.Title,
				Clicks:      url.Url.Clicks,
				CreatedAt:   url.Url.CreatedAt,
				ExpiresAt:   url.Url.ExpiresAt,
			}
			ctx.JSON(http.StatusOK, response)
			return
		}
	}

	if time.Now().After(existingUrl.ExpiresAt) {

		newExpire := time.Now().Add(config.URL_EXPIRE_DURATION)

		updated, err := server.store.UpdateExpireUrl(ctx, db.UpdateExpireUrlParams{
			ID:        existingUrl.ID,
			ExpiresAt: newExpire,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusOK, urlsResponse{
			Code:        updated.Code.String,
			ShortURL:    updated.ShortUrl.String,
			OriginalURL: updated.OriginalUrl,
			Title:       updated.Title,
			Clicks:      updated.Clicks,
			CreatedAt:   updated.CreatedAt,
			ExpiresAt:   updated.ExpiresAt,
		})
		return
	}

	ctx.JSON(http.StatusOK, urlsResponse{
		Code:        existingUrl.Code.String,
		ShortURL:    existingUrl.ShortUrl.String,
		OriginalURL: existingUrl.OriginalUrl,
		Title:       existingUrl.Title,
		Clicks:      existingUrl.Clicks,
		CreatedAt:   existingUrl.CreatedAt,
		ExpiresAt:   existingUrl.ExpiresAt,
	})

}

type getUrlByCodeRedirect struct {
	Code string `uri:"code" binding:"required,alphanum"`
}

func (server *Server) redirectUrl(ctx *gin.Context) {
	var req getUrlByCodeRedirect
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	code := req.Code

	url, err := server.store.GetUrlByCode(ctx, sql.NullString{String: code, Valid: true})
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}
	if time.Now().After(url.ExpiresAt) {

		newExpire := time.Now().Add(config.URL_EXPIRE_DURATION)

		updated, err := server.store.UpdateExpireUrl(ctx, db.UpdateExpireUrlParams{
			ID:        url.ID,
			ExpiresAt: newExpire,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		_, err = server.store.UpdateClicks(ctx, url.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		ctx.Redirect(http.StatusFound, updated.OriginalUrl)
		return
	}
	_, err = server.store.UpdateClicks(ctx, url.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.Redirect(http.StatusFound, url.OriginalUrl)
}
