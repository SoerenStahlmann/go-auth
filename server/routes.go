package server

import (
	"github.com/gin-gonic/gin"
	"github.com/soerenstahlmann/go-auth/middleware"
)

func (s *server) routes() {
	s.router.GET("/health", s.health)

	auth := s.router.Group("/auth")
	auth.POST("/signup", s.register)
	auth.POST("/login", s.login)
	auth.POST("/refresh", middleware.Auth(s.jwtSecret), s.refresh)
}

func (s *server) health(c *gin.Context) {
	c.JSON(200, map[string]string{"status": "ok"})
}
