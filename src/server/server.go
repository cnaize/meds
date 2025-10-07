package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/src/api"
)

type Server struct {
	router *gin.Engine
	server *http.Server
}

func NewServer(addr, username, password string) *Server {
	r := gin.New()
	r.Use(gin.BasicAuth(gin.Accounts{username: password}), gin.Recovery())

	api.Register(r)

	return &Server{
		router: r,
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	return s.server.ListenAndServe()
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}
