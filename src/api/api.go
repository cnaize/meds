package api

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cnaize/meds/src/core/metrics"
	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

// TODO: move out of global
var (
	subnetWhiteListMu sync.Mutex
	subnetBlackListMu sync.Mutex
	domainWhiteListMu sync.Mutex
	domainBlackListMu sync.Mutex
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

	// register api endpoints
	root := r.Group("/v1")
	root.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	// register whitelist api
	whitelist := root.Group("/whitelist")
	// register subnet whitelist
	snWhiteList := whitelist.Group("/subnets")
	snWhiteList.GET("", subnetListGetAll(subnetWhiteList, &subnetWhiteListMu))
	snWhiteList.GET("/:subnet", subnetListLookup(subnetWhiteList, &subnetWhiteListMu))
	snWhiteList.POST("", subnetListUpsert(subnetWhiteList, &subnetWhiteListMu, db, db.Q.UpsertWhiteListSubnet))
	snWhiteList.DELETE("", subnetListRemove(subnetWhiteList, &subnetWhiteListMu, db, db.Q.RemoveWhiteListSubnet))
	// register domain whitelist
	dmWhiteList := whitelist.Group("/domains")
	dmWhiteList.GET("", domainListGetAll(domainWhiteList, &domainWhiteListMu))
	dmWhiteList.GET("/:domain", domainListLookup(domainWhiteList, &domainWhiteListMu))
	dmWhiteList.POST("", domainListUpsert(domainWhiteList, &domainWhiteListMu, db, db.Q.UpsertWhiteListDomain))
	dmWhiteList.DELETE("", domainListRemove(domainWhiteList, &domainWhiteListMu, db, db.Q.RemoveWhiteListDomain))

	// register blacklist api
	blacklist := root.Group("/blacklist")
	// register subnet blacklist
	snBlackList := blacklist.Group("/subnets")
	snBlackList.GET("", subnetListGetAll(subnetBlackList, &subnetBlackListMu))
	snBlackList.GET("/:subnet", subnetListLookup(subnetBlackList, &subnetBlackListMu))
	snBlackList.POST("", subnetListUpsert(subnetBlackList, &subnetBlackListMu, db, db.Q.UpsertBlackListSubnet))
	snBlackList.DELETE("", subnetListRemove(subnetBlackList, &subnetBlackListMu, db, db.Q.RemoveBlackListSubnet))
	// register domain blacklist
	dmBlackList := blacklist.Group("/domains")
	dmBlackList.GET("", domainListGetAll(domainBlackList, &domainBlackListMu))
	dmBlackList.GET("/:domain", domainListLookup(domainBlackList, &domainBlackListMu))
	dmBlackList.POST("", domainListUpsert(domainBlackList, &domainBlackListMu, db, db.Q.UpsertBlackListDomain))
	dmBlackList.DELETE("", domainListRemove(domainBlackList, &domainBlackListMu, db, db.Q.RemoveBlackListDomain))
}
