package api

//	@title			Meds: net healing
//	@version		v1.2.1
//	@description	Hybrid firewall using public blocklists
//
//	@contact.name	cnaize
//	@contact.url	https://github.com/cnaize
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

//go:generate go tool swag init -g api.go -o ../../docs
//go:generate go tool swag fmt

// TODO: move out of global
var (
	ipAllowListMu       sync.Mutex
	countryBlockListMu  sync.Mutex
	ipIncludeListMu     sync.Mutex
	ipExcludeListMu     sync.Mutex
	asnIncludeListMu    sync.Mutex
	asnExcludeListMu    sync.Mutex
	ja3IncludeListMu    sync.Mutex
	ja3ExcludeListMu    sync.Mutex
	domainIncludeListMu sync.Mutex
	domainExcludeListMu sync.Mutex
)

func Register(
	r *gin.Engine,
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
) {
	// register prometheus metrics
	reg := prometheus.NewRegistry()
	metrics.Get().Register(reg)

	// register api endpoints
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	root := r.Group("/v1")

	// register subnets api
	snList := root.Group("/subnets")
	// register subnets allowlist
	snAllowList := snList.Group("/allowlist")
	snAllowList.GET("", SubnetAllowListGet(ipAllowList, &ipAllowListMu))
	snAllowList.GET("/:ip", SubnetAllowListCheck(ipAllowList, &ipAllowListMu))
	snAllowList.POST("", SubnetAllowListUpsert(ipAllowList, &ipAllowListMu, db))
	snAllowList.DELETE("", SubnetAllowListRemove(ipAllowList, &ipAllowListMu, db))
	// register subnets blocklist
	snBlockList := snList.Group("/blocklist")
	// register subnets includelist
	snIncludeList := snBlockList.Group("/include")
	snIncludeList.GET("", SubnetIncludeListGet(ipIncludeList, &ipIncludeListMu))
	snIncludeList.GET("/:ip", SubnetIncludeListCheck(ipIncludeList, &ipIncludeListMu))
	snIncludeList.POST("", SubnetIncludeListUpsert(ipIncludeList, &ipIncludeListMu, db))
	snIncludeList.DELETE("", SubnetIncludeListRemove(ipIncludeList, &ipIncludeListMu, db))
	// register subnets excludelist
	snExcludeList := snBlockList.Group("/exclude")
	snExcludeList.GET("", SubnetExcludeListGet(ipExcludeList, &ipExcludeListMu))
	snExcludeList.GET("/:ip", SubnetExcludeListCheck(ipExcludeList, &ipExcludeListMu))
	snExcludeList.POST("", SubnetExcludeListUpsert(ipExcludeList, &ipExcludeListMu, db))
	snExcludeList.DELETE("", SubnetExcludeListRemove(ipExcludeList, &ipExcludeListMu, db))

	// register asns api
	anList := root.Group("/asns")
	// register asns blocklist
	anBlockList := anList.Group("/blocklist")
	// register asns includelist
	anIncludeList := anBlockList.Group("/include")
	anIncludeList.GET("", ASNIncludeListGet(asnIncludeList, &asnIncludeListMu))
	anIncludeList.GET("/:asn", ASNIncludeListCheck(asnIncludeList, &asnIncludeListMu))
	anIncludeList.POST("", ASNIncludeListUpsert(asnIncludeList, &asnIncludeListMu, db))
	anIncludeList.DELETE("", ASNIncludeListRemove(asnIncludeList, &asnIncludeListMu, db))
	// register asns excludelist
	anExcludeList := anBlockList.Group("/exclude")
	anExcludeList.GET("", ASNExcludeListGet(asnExcludeList, &asnExcludeListMu))
	anExcludeList.GET("/:asn", ASNExcludeListCheck(asnExcludeList, &asnExcludeListMu))
	anExcludeList.POST("", ASNExcludeListUpsert(asnExcludeList, &asnExcludeListMu, db))
	anExcludeList.DELETE("", ASNExcludeListRemove(asnExcludeList, &asnExcludeListMu, db))

	// register ja3 api
	j3List := root.Group("/ja3")
	// register ja3 blocklist
	j3BlockList := j3List.Group("/blocklist")
	// register ja3 includelist
	j3IncludeList := j3BlockList.Group("/include")
	j3IncludeList.GET("", JA3IncludeListGet(ja3IncludeList, &ja3IncludeListMu))
	j3IncludeList.GET("/:hash", JA3IncludeListCheck(ja3IncludeList, &ja3IncludeListMu))
	j3IncludeList.POST("", JA3IncludeListUpsert(ja3IncludeList, &ja3IncludeListMu, db))
	j3IncludeList.DELETE("", JA3IncludeListRemove(ja3IncludeList, &ja3IncludeListMu, db))
	// register ja3 excludelist
	j3ExcludeList := j3BlockList.Group("/exclude")
	j3ExcludeList.GET("", JA3ExcludeListGet(ja3ExcludeList, &ja3ExcludeListMu))
	j3ExcludeList.GET("/:hash", JA3ExcludeListCheck(ja3ExcludeList, &ja3ExcludeListMu))
	j3ExcludeList.POST("", JA3ExcludeListUpsert(ja3ExcludeList, &ja3ExcludeListMu, db))
	j3ExcludeList.DELETE("", JA3ExcludeListRemove(ja3ExcludeList, &ja3ExcludeListMu, db))

	// register domains api
	dmList := root.Group("/domains")
	// register domains blocklist
	dmBlockList := dmList.Group("/blocklist")
	// register domains includelist
	dmIncludeList := dmBlockList.Group("/include")
	dmIncludeList.GET("", DomainIncludeListGet(domainIncludeList, &domainIncludeListMu))
	dmIncludeList.GET("/:domain", DomainIncludeListCheck(domainIncludeList, &domainIncludeListMu))
	dmIncludeList.POST("", DomainIncludeListUpsert(domainIncludeList, &domainIncludeListMu, db))
	dmIncludeList.DELETE("", DomainIncludeListRemove(domainIncludeList, &domainIncludeListMu, db))
	// register domains
	dmExcludeList := dmBlockList.Group("/exclude")
	dmExcludeList.GET("", DomainExcludeListGet(domainExcludeList, &domainExcludeListMu))
	dmExcludeList.GET("/:domain", DomainExcludeListCheck(domainExcludeList, &domainExcludeListMu))
	dmExcludeList.POST("", DomainExcludeListUpsert(domainExcludeList, &domainExcludeListMu, db))
	dmExcludeList.DELETE("", DomainExcludeListRemove(domainExcludeList, &domainExcludeListMu, db))

	// register countries api
	crList := root.Group("/countries")
	// register countries blocklist
	crBlockList := crList.Group("/blocklist")
	crBlockList.GET("", CountryBlockListGet(countryBlockList, &countryBlockListMu))
	crBlockList.GET("/:country", CountryBlockListCheck(countryBlockList, &countryBlockListMu))
	crBlockList.POST("", CountryBlockListUpsert(countryBlockList, &countryBlockListMu, db))
	crBlockList.DELETE("", CountryBlockListRemove(countryBlockList, &countryBlockListMu, db))
}
