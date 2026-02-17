package api

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

// ASNIncludeListGet godoc
//
//	@Summary		Get included ASNs
//	@Description	get all included ASNs
//	@Tags			ASNs
//	@Produce		json
//	@Success		200	{object}	GetASNsResp
//	@Router			/v1/asns/blocklist/include [get]
func ASNIncludeListGet(include *types.MapList[uint32], mu *sync.Mutex) func(*gin.Context) {
	return asnListGetAll(include, mu)
}

// ASNExcludeListGet godoc
//
//	@Summary		Get excluded ASNs
//	@Description	get all excluded ASNs
//	@Tags			ASNs
//	@Produce		json
//	@Success		200	{object}	GetASNsResp
//	@Router			/v1/asns/blocklist/exclude [get]
func ASNExcludeListGet(exclude *types.MapList[uint32], mu *sync.Mutex) func(*gin.Context) {
	return asnListGetAll(exclude, mu)
}

type GetASNsResp struct {
	ASNs []uint32 `json:"asns" example:"206831,400328"`
}

func asnListGetAll(list *types.MapList[uint32], mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, GetASNsResp{ASNs: list.GetAll()})
	}
}

// ASNIncludeListCheck godoc
//
//	@Summary		Check included ASN
//	@Description	check if an ASN is included
//	@Tags			ASNs
//	@Produce		json
//	@Param			asn	path		string	true	"ASN to check"
//	@Success		200	{object}	CheckASNResp
//	@Failure		400
//	@Router			/v1/asns/blocklist/include/{asn} [get]
func ASNIncludeListCheck(include *types.MapList[uint32], mu *sync.Mutex) func(*gin.Context) {
	return asnListLookup(include, mu)
}

// ASNExcludeListCheck godoc
//
//	@Summary		Check excluded ASN
//	@Description	check if an ASN is excluded
//	@Tags			ASNs
//	@Produce		json
//	@Param			asn	path		string	true	"ASN to check"
//	@Success		200	{object}	CheckASNResp
//	@Failure		400
//	@Router			/v1/asns/blocklist/exclude/{asn} [get]
func ASNExcludeListCheck(exclude *types.MapList[uint32], mu *sync.Mutex) func(*gin.Context) {
	return asnListLookup(exclude, mu)
}

type CheckASNResp struct {
	Found bool `json:"found"`
}

func asnListLookup(list *types.MapList[uint32], mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		asn, err := strconv.Atoi(c.Param("asn"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, CheckASNResp{
			Found: list.Lookup(uint32(asn)),
		})
	}
}

// ASNIncludeListUpsert godoc
//
//	@Summary		Upsert included ASNs
//	@Description	upsert ASNs to includelist
//	@Tags			ASNs
//	@Accept			json
//	@Param			body	body	UpsertASNsReq	true	"ASNs to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/asns/blocklist/include [post]
func ASNIncludeListUpsert(include *types.MapList[uint32], mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return asnListUpsert(include, mu, db, db.Q.UpsertASNIncludeList)
}

// ASNExcludeListUpsert godoc
//
//	@Summary		Upsert excluded ASNs
//	@Description	upsert ASNs to excludelist
//	@Tags			ASNs
//	@Accept			json
//	@Param			body	body	UpsertASNsReq	true	"ASNs to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/asns/blocklist/exclude [post]
func ASNExcludeListUpsert(include *types.MapList[uint32], mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return asnListUpsert(include, mu, db, db.Q.UpsertASNExcludeList)
}

type UpsertASNsReq struct {
	ASNs []uint32 `json:"asns" example:"206831,400328"`
}

func asnListUpsert(
	list *types.MapList[uint32],
	mu *sync.Mutex,
	db *database.Database,
	upsertFn func(ctx context.Context, db database.DBTX, asn int64) error,
) func(*gin.Context) {
	return func(c *gin.Context) {
		var req UpsertASNsReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if err := list.Upsert(req.ASNs); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, asn := range req.ASNs {
			if err := upsertFn(c, db.DB, int64(asn)); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}

// ASNIncludeListRemove godoc
//
//	@Summary		Remove included ASNs
//	@Description	remove ASNs from includelist
//	@Tags			ASNs
//	@Accept			json
//	@Param			body	body	RemoveASNsReq	true	"ASNs to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/asns/blocklist/include [delete]
func ASNIncludeListRemove(include *types.MapList[uint32], mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return asnListRemove(include, mu, db, db.Q.RemoveASNIncludeList)
}

// ASNExcludeListRemove godoc
//
//	@Summary		Remove excluded ASNs
//	@Description	remove ASNs from excludelist
//	@Tags			ASNs
//	@Accept			json
//	@Param			body	body	RemoveASNsReq	true	"ASNs to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/asns/blocklist/exclude [delete]
func ASNExcludeListRemove(include *types.MapList[uint32], mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return asnListRemove(include, mu, db, db.Q.RemoveASNExcludeList)
}

type RemoveASNsReq struct {
	ASNs []uint32 `json:"asns" example:"206831,400328"`
}

func asnListRemove(
	list *types.MapList[uint32],
	mu *sync.Mutex,
	db *database.Database,
	removeFn func(ctx context.Context, db database.DBTX, asn int64) error,
) func(*gin.Context) {
	return func(c *gin.Context) {
		var req RemoveASNsReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if err := list.Remove(req.ASNs); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, asn := range req.ASNs {
			if err := removeFn(c, db.DB, int64(asn)); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}
