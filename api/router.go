package api

import "github.com/gin-gonic/gin"

func (server *Server) setupRouter() *gin.Engine {
	router := gin.Default()

	router.POST("/urls", server.newUrlsResponse)
	router.GET("/:code", server.redirectUrl)

	server.router = router
	return router
}
