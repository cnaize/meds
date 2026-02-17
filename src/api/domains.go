package api

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

// DomainIncludeListGet godoc
//
//	@Summary		Get included domains
//	@Description	get all included domains
//	@Tags			Domains
//	@Produce		json
//	@Success		200	{object}	GetDomainsResp
//	@Router			/v1/domains/blocklist/include [get]
func DomainIncludeListGet(include *types.DomainList, mu *sync.Mutex) func(*gin.Context) {
	return domainListGetAll(include, mu)
}

// DomainExcludeListGet godoc
//
//	@Summary		Get excluded domains
//	@Description	get all excluded domains
//	@Tags			Domains
//	@Produce		json
//	@Success		200	{object}	GetDomainsResp
//	@Router			/v1/domains/blocklist/exclude [get]
func DomainExcludeListGet(exclude *types.DomainList, mu *sync.Mutex) func(*gin.Context) {
	return domainListGetAll(exclude, mu)
}

type GetDomainsResp struct {
	Domains []string `json:"domains" example:"good.com,bad.com"`
}

func domainListGetAll(list *types.DomainList, mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, GetDomainsResp{Domains: list.GetAll()})
	}
}

// DomainIncludeListCheck godoc
//
//	@Summary		Check included domains
//	@Description	check if a domain is included
//	@Tags			Domains
//	@Produce		json
//	@Param			domain	path		string	true	"domain to check"
//	@Success		200		{object}	CheckDomainResp
//	@Router			/v1/domains/blocklist/include/{domain} [get]
func DomainIncludeListCheck(include *types.DomainList, mu *sync.Mutex) func(*gin.Context) {
	return domainListLookup(include, mu)
}

// DomainExcludeListCheck godoc
//
//	@Summary		Check excluded domains
//	@Description	check if a domain is excluded
//	@Tags			Domains
//	@Produce		json
//	@Param			domain	path		string	true	"domain to check"
//	@Success		200		{object}	CheckDomainResp
//	@Router			/v1/domains/blocklist/exclude/{domain} [get]
func DomainExcludeListCheck(exclude *types.DomainList, mu *sync.Mutex) func(*gin.Context) {
	return domainListLookup(exclude, mu)
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

// DomainIncludeListUpsert godoc
//
//	@Summary		Upsert included domains
//	@Description	upsert domains to includelist
//	@Tags			Domains
//	@Accept			json
//	@Param			body	body	UpsertDomainsReq	true	"domains to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/domains/blocklist/include [post]
func DomainIncludeListUpsert(include *types.DomainList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return domainListUpsert(include, mu, db, db.Q.UpsertDomainIncludeList)
}

// DomainExcludeListUpsert godoc
//
//	@Summary		Upsert excluded domains
//	@Description	upsert domains to excludelist
//	@Tags			Domains
//	@Accept			json
//	@Param			body	body	UpsertDomainsReq	true	"domains to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/domains/blocklist/exclude [post]
func DomainExcludeListUpsert(exclude *types.DomainList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return domainListUpsert(exclude, mu, db, db.Q.UpsertDomainExcludeList)
}

type UpsertDomainsReq struct {
	Domains []string `json:"domains" example:"good.com,bad.com"`
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

// DomainIncludeListRemove godoc
//
//	@Summary		Remove included domains
//	@Description	remove domains from includelist
//	@Tags			Domains
//	@Accept			json
//	@Param			body	body	RemoveDomainsReq	true	"domains to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/domains/blocklist/include [delete]
func DomainIncludeListRemove(include *types.DomainList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return domainListRemove(include, mu, db, db.Q.RemoveDomainIncludeList)
}

// DomainExcludeListRemove godoc
//
//	@Summary		Remove excluded domains
//	@Description	remove domains from excludelist
//	@Tags			Domains
//	@Accept			json
//	@Param			body	body	RemoveDomainsReq	true	"domains to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/domains/blocklist/exclude [delete]
func DomainExcludeListRemove(exclude *types.DomainList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return domainListRemove(exclude, mu, db, db.Q.RemoveDomainExcludeList)
}

type RemoveDomainsReq struct {
	Domains []string `json:"domains" example:"good.com,bad.com"`
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
