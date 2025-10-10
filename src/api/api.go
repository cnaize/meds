package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cnaize/meds/src/core/metrics"
	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

func Register(
	r *gin.Engine,
	db *database.Database,
	subnetWhiteList *types.SubnetList,
	subnetBlackList *types.SubnetList,
	domainWhiteList *types.DomainList,
	domainBlackList *types.DomainList,
) {
	// register prometheus metrics
	reg := prometheus.NewRegistry()
	metrics.Get().Register(reg)

	root := r.Group("/v1")
	root.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	// register api endpoints
	whitelist := root.Group("/whitelist")

	wlSubnet := whitelist.Group("/subnets")
	wlSubnet.GET("", subnetWhiteListGetAll(subnetWhiteList))
	wlSubnet.GET("/:subnet", subnetWhiteListLookup(subnetWhiteList))
	wlSubnet.POST("", subnetWhiteListUpsert(db, subnetWhiteList))
	wlSubnet.DELETE("", subnetWhiteListRemove(db, subnetWhiteList))
}
