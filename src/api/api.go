package api

//	@title			Meds: net healing
//	@version		v0.9.0
//	@description	NFQUEUE firewall written in Go
//
//	@contact.name	cnaize
//	@contact.url	https://github.com/cnaize/meds
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/cnaize/meds/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/cnaize/meds/src/core/metrics"
	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

// TODO: move out of global
var (
	subnetWhiteListMu  sync.Mutex
	subnetBlackListMu  sync.Mutex
	domainWhiteListMu  sync.Mutex
	domainBlackListMu  sync.Mutex
	countryBlackListMu sync.Mutex
)

func Register(
	r *gin.Engine,
	db *database.Database,
	subnetWhiteList *types.SubnetList,
	subnetBlackList *types.SubnetList,
	domainWhiteList *types.DomainList,
	domainBlackList *types.DomainList,
	countryBlackList *types.CountryList,
) {
	// register prometheus metrics
	reg := prometheus.NewRegistry()
	metrics.Get().Register(reg)

	// register api endpoints
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	root := r.Group("/v1")

	// register whitelist api
	whitelist := root.Group("/whitelist")
	// register subnet whitelist
	snWhiteList := whitelist.Group("/subnets")
	snWhiteList.GET("", GetWhiteListSubnets(subnetWhiteList, &subnetWhiteListMu))
	snWhiteList.GET("/:subnet", CheckWhiteListSubnet(subnetWhiteList, &subnetWhiteListMu))
	snWhiteList.POST("", UpsertWhiteListSubnets(subnetWhiteList, &subnetWhiteListMu, db))
	snWhiteList.DELETE("", RemoveWhiteListSubnets(subnetWhiteList, &subnetWhiteListMu, db))
	// register domain whitelist
	dmWhiteList := whitelist.Group("/domains")
	dmWhiteList.GET("", GetWhiteListDomains(domainWhiteList, &domainWhiteListMu))
	dmWhiteList.GET("/:domain", CheckWhiteListDomain(domainWhiteList, &domainWhiteListMu))
	dmWhiteList.POST("", UpsertWhiteListDomains(domainWhiteList, &domainWhiteListMu, db))
	dmWhiteList.DELETE("", RemoveWhiteListDomains(domainWhiteList, &domainWhiteListMu, db))

	// register blacklist api
	blacklist := root.Group("/blacklist")
	// register subnet blacklist
	snBlackList := blacklist.Group("/subnets")
	snBlackList.GET("", GetBlackListSubnets(subnetBlackList, &subnetBlackListMu))
	snBlackList.GET("/:subnet", CheckBlackListSubnet(subnetBlackList, &subnetBlackListMu))
	snBlackList.POST("", UpsertBlackListSubnets(subnetBlackList, &subnetBlackListMu, db))
	snBlackList.DELETE("", RemoveBlackListSubnets(subnetBlackList, &subnetBlackListMu, db))
	// register domain blacklist
	dmBlackList := blacklist.Group("/domains")
	dmBlackList.GET("", GetBlackListDomains(domainBlackList, &domainBlackListMu))
	dmBlackList.GET("/:domain", CheckBlackListDomain(domainBlackList, &domainBlackListMu))
	dmBlackList.POST("", UpsertBlackListDomains(domainBlackList, &domainBlackListMu, db))
	dmBlackList.DELETE("", RemoveBlackListDomains(domainBlackList, &domainBlackListMu, db))
	// register country blacklist
	crBlackList := blacklist.Group("/countries")
	crBlackList.GET("", GetBlackListCountries(countryBlackList, &countryBlackListMu))
	crBlackList.GET("/:country", CheckBlackListCountry(countryBlackList, &countryBlackListMu))
	crBlackList.POST("", UpsertBlackListCountries(countryBlackList, &countryBlackListMu, db))
	crBlackList.DELETE("", RemoveBlackListCountries(countryBlackList, &countryBlackListMu, db))
}
