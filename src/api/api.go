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

	// register api endpoints
	root := r.Group("/v1")
	root.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	// register whitelist api
	whitelist := root.Group("/whitelist")
	// register subnet whitelist
	snWhiteList := whitelist.Group("/subnets")
	snWhiteList.GET("", subnetListGetAll(subnetWhiteList))
	snWhiteList.GET("/:subnet", subnetListLookup(subnetWhiteList))
	snWhiteList.POST("", subnetListUpsert(subnetWhiteList, db, db.Q.UpsertWhiteListSubnet))
	snWhiteList.DELETE("", subnetListRemove(subnetWhiteList, db, db.Q.RemoveWhiteListSubnet))
	// register domain whitelist
	dmWhiteList := whitelist.Group("/domains")
	dmWhiteList.GET("", domainListGetAll(domainWhiteList))
	dmWhiteList.GET("/:domain", domainListLookup(domainWhiteList))
	dmWhiteList.POST("", domainListUpsert(domainWhiteList, db, db.Q.UpsertWhiteListDomain))
	dmWhiteList.DELETE("", domainListRemove(domainWhiteList, db, db.Q.RemoveWhiteListDomain))

	// register blacklist api
	blacklist := root.Group("/blacklist")
	// register subnet blacklist
	snBlackList := blacklist.Group("/subnets")
	snBlackList.GET("", subnetListGetAll(subnetBlackList))
	snBlackList.GET("/:subnet", subnetListLookup(subnetBlackList))
	snBlackList.POST("", subnetListUpsert(subnetBlackList, db, db.Q.UpsertBlackListSubnet))
	snBlackList.DELETE("", subnetListRemove(subnetBlackList, db, db.Q.RemoveBlackListSubnet))
	// register domain blacklist
	dmBlackList := blacklist.Group("/domains")
	dmBlackList.GET("", domainListGetAll(domainBlackList))
	dmBlackList.GET("/:domain", domainListLookup(domainBlackList))
	dmBlackList.POST("", domainListUpsert(domainBlackList, db, db.Q.UpsertBlackListDomain))
	dmBlackList.DELETE("", domainListRemove(domainBlackList, db, db.Q.RemoveBlackListDomain))
}
