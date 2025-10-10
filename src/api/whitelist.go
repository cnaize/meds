package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

func subnetWhiteListGetAll(whiteList *types.SubnetList) func(*gin.Context) {
	type Out struct {
		Subnets []string `json:"subnets"`
	}

	return func(c *gin.Context) {
		all := whiteList.GetAll()
		subnets := make([]string, len(all))
		for i, subnet := range all {
			subnets[i] = subnet.String()
		}

		c.JSON(http.StatusOK, Out{Subnets: subnets})
	}
}

func subnetWhiteListLookup(whiteList *types.SubnetList) func(*gin.Context) {
	return func(c *gin.Context) {
		subnet := c.Param("subnet")
		prefix, ok := get.Subnet(subnet)
		if !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"found": whiteList.Lookup(prefix),
		})
	}
}

func subnetWhiteListUpsert(db *database.Database, whiteList *types.SubnetList) func(*gin.Context) {
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

		if err := whiteList.Upsert(subnets); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, subnet := range subnets {
			if err := db.Q.UpsertWhiteListSubnet(c, db.DB, subnet.String()); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}

func subnetWhiteListRemove(db *database.Database, whiteList *types.SubnetList) func(*gin.Context) {
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

		if err := whiteList.Remove(subnets); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, subnet := range subnets {
			if err := db.Q.RemoveWhiteListSubnet(c, db.DB, subnet.String()); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}
