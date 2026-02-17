package api

import (
	"context"
	"net/http"
	"net/netip"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

// SubnetAllowListGet godoc
//
//	@Summary		Get allowed subnets
//	@Description	get all allowed subnets
//	@Tags			Subnets
//	@Produce		json
//	@Success		200	{object}	GetSubnetsResp
//	@Router			/v1/subnets/allowlist [get]
func SubnetAllowListGet(allowlist *types.IPList, mu *sync.Mutex) func(*gin.Context) {
	return subnetListGetAll(allowlist, mu)
}

// SubnetIncludeListGet godoc
//
//	@Summary		Get included subnets
//	@Description	get all included subnets
//	@Tags			Subnets
//	@Produce		json
//	@Success		200	{object}	GetSubnetsResp
//	@Router			/v1/subnets/blocklist/include [get]
func SubnetIncludeListGet(include *types.IPList, mu *sync.Mutex) func(*gin.Context) {
	return subnetListGetAll(include, mu)
}

// SubnetExcludeListGet godoc
//
//	@Summary		Get excluded subnets
//	@Description	get all excluded subnets
//	@Tags			Subnets
//	@Produce		json
//	@Success		200	{object}	GetSubnetsResp
//	@Router			/v1/subnets/blocklist/exclude [get]
func SubnetExcludeListGet(exclude *types.IPList, mu *sync.Mutex) func(*gin.Context) {
	return subnetListGetAll(exclude, mu)
}

type GetSubnetsResp struct {
	Subnets []string `json:"subnets" example:"100.100.100.100/32,200.200.200.0/24"`
}

func subnetListGetAll(list *types.IPList, mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		all := list.GetAll()
		subnets := make([]string, len(all))
		for i, subnet := range all {
			subnets[i] = subnet.String()
		}

		c.JSON(http.StatusOK, GetSubnetsResp{Subnets: subnets})
	}
}

// SubnetAllowListCheck godoc
//
//	@Summary		Check allowed ip address
//	@Description	check if an ip address is allowed
//	@Tags			Subnets
//	@Produce		json
//	@Param			ip	path		string	true	"ip address to check"
//	@Success		200	{object}	CheckSubnetResp
//	@Failure		400
//	@Router			/v1/subnets/allowlist/{ip} [get]
func SubnetAllowListCheck(allowlist *types.IPList, mu *sync.Mutex) func(*gin.Context) {
	return subnetListLookup(allowlist, mu)
}

// SubnetIncludeListCheck godoc
//
//	@Summary		Check included ip address
//	@Description	check if an ip address is included
//	@Tags			Subnets
//	@Produce		json
//	@Param			ip	path		string	true	"ip address to check"
//	@Success		200	{object}	CheckSubnetResp
//	@Failure		400
//	@Router			/v1/subnets/blocklist/include/{ip} [get]
func SubnetIncludeListCheck(include *types.IPList, mu *sync.Mutex) func(*gin.Context) {
	return subnetListLookup(include, mu)
}

// SubnetExcludeListCheck godoc
//
//	@Summary		Check excluded ip address
//	@Description	check if an ip address is excluded
//	@Tags			Subnets
//	@Produce		json
//	@Param			ip	path		string	true	"ip address to check"
//	@Success		200	{object}	CheckSubnetResp
//	@Failure		400
//	@Router			/v1/subnets/blocklist/exclude/{ip} [get]
func SubnetExcludeListCheck(exclude *types.IPList, mu *sync.Mutex) func(*gin.Context) {
	return subnetListLookup(exclude, mu)
}

type CheckSubnetResp struct {
	Found bool `json:"found"`
}

func subnetListLookup(list *types.IPList, mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		ip := c.Param("ip")
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, CheckSubnetResp{
			Found: list.Lookup(addr),
		})
	}
}

// SubnetAllowListUpsert godoc
//
//	@Summary		Upsert allowed subnets
//	@Description	upsert subnets to allowlist
//	@Tags			Subnets
//	@Accept			json
//	@Param			body	body	UpsertSubnetsReq	true	"subnets to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/subnets/allowlist [post]
func SubnetAllowListUpsert(allowlist *types.IPList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return subnetListUpsert(allowlist, mu, db, db.Q.UpsertIPAllowList)
}

// SubnetIncludeListUpsert godoc
//
//	@Summary		Upsert included subnets
//	@Description	upsert subnets to includelist
//	@Tags			Subnets
//	@Accept			json
//	@Param			body	body	UpsertSubnetsReq	true	"subnets to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/subnets/blocklist/include [post]
func SubnetIncludeListUpsert(include *types.IPList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return subnetListUpsert(include, mu, db, db.Q.UpsertIPIncludeList)
}

// SubnetExcludeListUpsert godoc
//
//	@Summary		Upsert excluded subnets
//	@Description	upsert subnets to excludelist
//	@Tags			Subnets
//	@Accept			json
//	@Param			body	body	UpsertSubnetsReq	true	"subnets to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/subnets/blocklist/exclude [post]
func SubnetExcludeListUpsert(exclude *types.IPList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return subnetListUpsert(exclude, mu, db, db.Q.UpsertIPExcludeList)
}

type UpsertSubnetsReq struct {
	Subnets []string `json:"subnets" example:"100.100.100.100,200.200.200.0/24"`
}

func subnetListUpsert(
	list *types.IPList,
	mu *sync.Mutex,
	db *database.Database,
	upsertFn func(ctx context.Context, db database.DBTX, subnet string) error,
) func(*gin.Context) {
	return func(c *gin.Context) {
		var req UpsertSubnetsReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		subnets, err := get.Subnets(req.Subnets)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if err := list.Upsert(subnets); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, subnet := range subnets {
			if err := upsertFn(c, db.DB, subnet.String()); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}

// SubnetAllowListRemove godoc
//
//	@Summary		Remove allowed subnets
//	@Description	remove subnets from allowlist
//	@Tags			Subnets
//	@Accept			json
//	@Param			body	body	RemoveSubnetsReq	true	"subnets to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/subnets/allowlist [delete]
func SubnetAllowListRemove(allowlist *types.IPList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return subnetListRemove(allowlist, mu, db, db.Q.RemoveIPAllowList)
}

// SubnetIncludeListRemove godoc
//
//	@Summary		Remove included subnets
//	@Description	remove subnets from includelist
//	@Tags			Subnets
//	@Accept			json
//	@Param			body	body	RemoveSubnetsReq	true	"subnets to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/subnets/blocklist/include [delete]
func SubnetIncludeListRemove(include *types.IPList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return subnetListRemove(include, mu, db, db.Q.RemoveIPIncludeList)
}

// SubnetExcludeListRemove godoc
//
//	@Summary		Remove excluded subnets
//	@Description	remove subnets from excludelist
//	@Tags			Subnets
//	@Accept			json
//	@Param			body	body	RemoveSubnetsReq	true	"subnets to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/subnets/blocklist/exclude [delete]
func SubnetExcludeListRemove(exclude *types.IPList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return subnetListRemove(exclude, mu, db, db.Q.RemoveIPExcludeList)
}

type RemoveSubnetsReq struct {
	Subnets []string `json:"subnets" example:"100.100.100.100,200.200.200.0/24"`
}

func subnetListRemove(
	list *types.IPList,
	mu *sync.Mutex,
	db *database.Database,
	removeFn func(ctx context.Context, db database.DBTX, subnet string) error,
) func(*gin.Context) {
	return func(c *gin.Context) {
		var req RemoveSubnetsReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		subnets, err := get.Subnets(req.Subnets)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if err := list.Remove(subnets); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, subnet := range subnets {
			if err := removeFn(c, db.DB, subnet.String()); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}
