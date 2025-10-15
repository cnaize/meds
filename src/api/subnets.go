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

func subnetListGetAll(list *types.SubnetList, mu *sync.Mutex) func(*gin.Context) {
	type Out struct {
		Subnets []string `json:"subnets"`
	}

	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		all := list.GetAll()
		subnets := make([]string, len(all))
		for i, subnet := range all {
			subnets[i] = subnet.String()
		}

		c.JSON(http.StatusOK, Out{Subnets: subnets})
	}
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

		c.JSON(http.StatusOK, gin.H{
			"found": list.Lookup(prefix),
		})
	}
}

func subnetListUpsert(
	list *types.SubnetList,
	mu *sync.Mutex,
	db *database.Database,
	upsertFn func(ctx context.Context, db database.DBTX, subnet string) error,
) func(*gin.Context) {
	type In struct {
		Subnets []string `json:"subnets"`
	}

	return func(c *gin.Context) {
		var in In
		if err := c.ShouldBindJSON(&in); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		subnets, err := get.Subnets(in.Subnets)
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

func subnetListRemove(
	list *types.SubnetList,
	mu *sync.Mutex,
	db *database.Database,
	removeFn func(ctx context.Context, db database.DBTX, subnet string) error,
) func(*gin.Context) {
	type In struct {
		Subnets []string `json:"subnets"`
	}

	return func(c *gin.Context) {
		var in In
		if err := c.ShouldBindJSON(&in); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		subnets, err := get.Subnets(in.Subnets)
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
