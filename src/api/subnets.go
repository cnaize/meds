package api

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

// GetWhiteListSubnets godoc
//
//	@Summary		Get whitelisted subnets
//	@Description	get all whitelisted subnets
//	@Tags			whitelist
//	@Produce		json
//	@Success		200	{object}	GetSubnetsResp
//	@Router			/v1/whitelist/subnets [get]
func GetWhiteListSubnets(whitelist *types.SubnetList, mu *sync.Mutex) func(*gin.Context) {
	return subnetListGetAll(whitelist, mu)
}

// GetBlackListSubnets godoc
//
//	@Summary		Get blacklisted subnets
//	@Description	get all blacklisted subnets
//	@Tags			blacklist
//	@Produce		json
//	@Success		200	{object}	GetSubnetsResp
//	@Router			/v1/blacklist/subnets [get]
func GetBlackListSubnets(blacklist *types.SubnetList, mu *sync.Mutex) func(*gin.Context) {
	return subnetListGetAll(blacklist, mu)
}

type GetSubnetsResp struct {
	Subnets []string `json:"subnets"`
}

func subnetListGetAll(list *types.SubnetList, mu *sync.Mutex) func(*gin.Context) {
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

// CheckWhiteListSubnet godoc
//
//	@Summary		Check whitelisted subnet
//	@Description	check if a subnet is whitelisted
//	@Tags			whitelist
//	@Produce		json
//	@Param			subnet	path		string	true	"subnet to check"
//	@Success		200		{object}	CheckSubnetResp
//	@Failure		400
//	@Router			/v1/whitelist/subnets/{subnet} [get]
func CheckWhiteListSubnet(whitelist *types.SubnetList, mu *sync.Mutex) func(*gin.Context) {
	return subnetListLookup(whitelist, mu)
}

// CheckBlackListSubnet godoc
//
//	@Summary		Check blacklisted subnet
//	@Description	check if a subnet is blacklisted
//	@Tags			blacklist
//	@Produce		json
//	@Param			subnet	path		string	true	"subnet to check"
//	@Success		200		{object}	CheckSubnetResp
//	@Failure		400
//	@Router			/v1/blacklist/subnets/{subnet} [get]
func CheckBlackListSubnet(blacklist *types.SubnetList, mu *sync.Mutex) func(*gin.Context) {
	return subnetListLookup(blacklist, mu)
}

type CheckSubnetResp struct {
	Found bool `json:"found"`
}

func subnetListLookup(list *types.SubnetList, mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		subnet := c.Param("subnet")
		prefix, ok := get.Subnet(subnet)
		if !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, CheckSubnetResp{
			Found: list.Lookup(prefix),
		})
	}
}

// UpsertWhiteListSubnets godoc
//
//	@Summary		Upsert whitelisted subnets
//	@Description	upsert subnets to whitelist
//	@Tags			whitelist
//	@Accept			json
//	@Param			body	body	UpsertSubnetsReq	true	"subnets to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/whitelist/subnets [post]
func UpsertWhiteListSubnets(whitelist *types.SubnetList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return subnetListUpsert(whitelist, mu, db, db.Q.UpsertWhiteListSubnet)
}

// UpsertBlackListSubnets godoc
//
//	@Summary		Upsert blacklisted subnets
//	@Description	upsert subnets to blacklist
//	@Tags			blacklist
//	@Accept			json
//	@Param			body	body	UpsertSubnetsReq	true	"subnets to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/blacklist/subnets [post]
func UpsertBlackListSubnets(blacklist *types.SubnetList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return subnetListUpsert(blacklist, mu, db, db.Q.UpsertBlackListSubnet)
}

type UpsertSubnetsReq struct {
	Subnets []string `json:"subnets"`
}

func subnetListUpsert(
	list *types.SubnetList,
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

// RemoveWhiteListSubnets godoc
//
//	@Summary		Remove whitelisted subnets
//	@Description	remove subnets from whitelist
//	@Tags			whitelist
//	@Accept			json
//	@Param			body	body	RemoveSubnetsReq	true	"subnets to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/whitelist/subnets [delete]
func RemoveWhiteListSubnets(whitelist *types.SubnetList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return subnetListRemove(whitelist, mu, db, db.Q.RemoveWhiteListSubnet)
}

// RemoveBlackListSubnets godoc
//
//	@Summary		Remove blacklisted subnets
//	@Description	remove subnets from blacklist
//	@Tags			blacklist
//	@Accept			json
//	@Param			body	body	RemoveSubnetsReq	true	"subnets to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/blacklist/subnets [delete]
func RemoveBlackListSubnets(blacklist *types.SubnetList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return subnetListRemove(blacklist, mu, db, db.Q.RemoveBlackListSubnet)
}

type RemoveSubnetsReq struct {
	Subnets []string `json:"subnets"`
}

func subnetListRemove(
	list *types.SubnetList,
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
