package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/soerenstahlmann/go-auth/ent"
	"go.uber.org/zap"
)

type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
	Status int         `json:"status,omitempty"`
}

type serverConfig func(s *server) error

type server struct {
	router    *gin.Engine
	jwtSecret []byte
	client    *ent.Client
	logger    *zap.SugaredLogger
}

func NewServer(cfgs ...serverConfig) (*server, error) {
	s := &server{
		router: gin.Default(),
	}

	for _, cfg := range cfgs {
		err := cfg(s)
		if err != nil {
			return nil, err
		}
	}

	s.routes()

	return s, nil
}

func (s *server) Run() error {
	fmt.Println("Starting server...")
	err := http.ListenAndServe(":8080", s)
	if err != nil {
		return fmt.Errorf("could not start server: %w", err)
	}

	return nil
}

func WithJWTSecret(jwtSecret []byte) serverConfig {
	return func(s *server) error {
		s.jwtSecret = jwtSecret
		return nil
	}
}

func WithClient(client *ent.Client) serverConfig {
	return func(s *server) error {
		s.client = client
		return nil
	}
}

func WithLogger(logger *zap.SugaredLogger) serverConfig {
	return func(s *server) error {
		s.logger = logger
		return nil
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
