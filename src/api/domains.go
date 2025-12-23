package api

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

// GetWhiteListDomains godoc
//
//	@Summary		Get whitelisted domains
//	@Description	get all whitelisted domains
//	@Tags			whitelist
//	@Produce		json
//	@Success		200	{object}	GetDomainsResp
//	@Router			/v1/whitelist/domains [get]
func GetWhiteListDomains(whitelist *types.DomainList, mu *sync.Mutex) func(*gin.Context) {
	return domainListGetAll(whitelist, mu)
}

// GetBlackListDomains godoc
//
//	@Summary		Get blacklisted domains
//	@Description	get all blacklisted domains
//	@Tags			blacklist
//	@Produce		json
//	@Success		200	{object}	GetDomainsResp
//	@Router			/v1/blacklist/domains [get]
func GetBlackListDomains(blacklist *types.DomainList, mu *sync.Mutex) func(*gin.Context) {
	return domainListGetAll(blacklist, mu)
}

type GetDomainsResp struct {
	Domains []string `json:"domains" example:"bad.com,dead.com"`
}

func domainListGetAll(list *types.DomainList, mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, GetDomainsResp{Domains: list.GetAll()})
	}
}

// CheckWhiteListDomain godoc
//
//	@Summary		Check whitelisted domain
//	@Description	check if a domain is whitelisted
//	@Tags			whitelist
//	@Produce		json
//	@Param			domain	path		string	true	"domain to check"
//	@Success		200		{object}	CheckDomainResp
//	@Router			/v1/whitelist/domains/{domain} [get]
func CheckWhiteListDomain(whitelist *types.DomainList, mu *sync.Mutex) func(*gin.Context) {
	return domainListLookup(whitelist, mu)
}

// CheckBlackListDomain godoc
//
//	@Summary		Check blacklisted domain
//	@Description	check if a domain is blacklisted
//	@Tags			blacklist
//	@Produce		json
//	@Param			domain	path		string	true	"domain to check"
//	@Success		200		{object}	CheckDomainResp
//	@Router			/v1/blacklist/domains/{domain} [get]
func CheckBlackListDomain(blacklist *types.DomainList, mu *sync.Mutex) func(*gin.Context) {
	return domainListLookup(blacklist, mu)
}

type CheckDomainResp struct {
	Found bool `json:"found"`
}

func domainListLookup(list *types.DomainList, mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		domain := c.Param("domain")

		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, CheckDomainResp{
			Found: list.Lookup(domain),
		})
	}
}

// UpsertWhiteListDomains godoc
//
//	@Summary		Upsert whitelisted domains
//	@Description	upsert domains to whitelist
//	@Tags			whitelist
//	@Accept			json
//	@Param			body	body	UpsertDomainsReq	true	"domains to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/whitelist/domains [post]
func UpsertWhiteListDomains(whitelist *types.DomainList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return domainListUpsert(whitelist, mu, db, db.Q.UpsertWhiteListDomain)
}

// UpsertBlackListDomains godoc
//
//	@Summary		Upsert blacklisted domains
//	@Description	upsert domains to blacklist
//	@Tags			blacklist
//	@Accept			json
//	@Param			body	body	UpsertDomainsReq	true	"domains to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/blacklist/domains [post]
func UpsertBlackListDomains(blacklist *types.DomainList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return domainListUpsert(blacklist, mu, db, db.Q.UpsertBlackListDomain)
}

type UpsertDomainsReq struct {
	Domains []string `json:"domains" example:"bad.com,dead.com"`
}

func domainListUpsert(
	list *types.DomainList,
	mu *sync.Mutex,
	db *database.Database,
	upsertFn func(ctx context.Context, db database.DBTX, domain string) error,
) func(*gin.Context) {
	return func(c *gin.Context) {
		var req UpsertDomainsReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if err := list.Upsert(req.Domains); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, domain := range req.Domains {
			if err := upsertFn(c, db.DB, domain); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}

// RemoveWhiteListDomains godoc
//
//	@Summary		Remove whitelisted domains
//	@Description	remove domains from whitelist
//	@Tags			whitelist
//	@Accept			json
//	@Param			body	body	RemoveDomainsReq	true	"domains to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/whitelist/domains [delete]
func RemoveWhiteListDomains(whitelist *types.DomainList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return domainListRemove(whitelist, mu, db, db.Q.RemoveWhiteListDomain)
}

// RemoveBlackListDomains godoc
//
//	@Summary		Remove blacklisted domains
//	@Description	remove domains from blacklist
//	@Tags			blacklist
//	@Accept			json
//	@Param			body	body	RemoveDomainsReq	true	"domains to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/blacklist/domains [delete]
func RemoveBlackListDomains(blacklist *types.DomainList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return domainListRemove(blacklist, mu, db, db.Q.RemoveBlackListDomain)
}

type RemoveDomainsReq struct {
	Domains []string `json:"domains" example:"bad.com,dead.com"`
}

func domainListRemove(
	list *types.DomainList,
	mu *sync.Mutex,
	db *database.Database,
	removeFn func(ctx context.Context, db database.DBTX, domain string) error,
) func(*gin.Context) {
	return func(c *gin.Context) {
		var req RemoveDomainsReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if err := list.Remove(req.Domains); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, domain := range req.Domains {
			if err := removeFn(c, db.DB, domain); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}
