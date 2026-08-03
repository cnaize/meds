package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/src/api"
	"github.com/cnaize/meds/src/config"
	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

type Server struct {
	router *gin.Engine
	server *http.Server
}

func NewServer(
	cfg *config.Config,
	db *database.Database,
	ipAllowList *types.IPList,
	countryBlockList *types.CountryList,
	ipIncludeList *types.IPList,
	ipExcludeList *types.IPList,
	asnIncludeList *types.MapList[uint32],
	asnExcludeList *types.MapList[uint32],
	ja3IncludeList *types.MapList[string],
	ja3ExcludeList *types.MapList[string],
	domainIncludeList *types.DomainList,
	domainExcludeList *types.DomainList,
) *Server {
	r := gin.New()
	r.Use(gin.BasicAuth(gin.Accounts{cfg.Username: cfg.Password}), gin.Recovery())

	api.Register(
		r,
		db,
		ipAllowList,
		countryBlockList,
		ipIncludeList,
		ipExcludeList,
		asnIncludeList,
		asnExcludeList,
		ja3IncludeList,
		ja3ExcludeList,
		domainIncludeList,
		domainExcludeList,
	)

	return &Server{
		router: r,
		server: &http.Server{
			Addr:    cfg.APIServerAddr,
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

func (s *Server) Close(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
