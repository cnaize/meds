package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/src/api"
	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

type Server struct {
	router *gin.Engine
	server *http.Server
}

func NewServer(
	addr,
	username,
	password string,
	db *database.Database,
	subnetWhiteList *types.SubnetList,
	subnetBlackList *types.SubnetList,
	domainWhiteList *types.DomainList,
	domainBlackList *types.DomainList,
) *Server {
	r := gin.New()
	r.Use(gin.BasicAuth(gin.Accounts{username: password}), gin.Recovery())

	api.Register(r, db, subnetWhiteList, subnetBlackList, domainWhiteList, domainBlackList)

	return &Server{
		router: r,
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}
