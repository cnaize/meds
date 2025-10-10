package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

func domainListGetAll(list *types.DomainList) func(*gin.Context) {
	type Out struct {
		Domains []string `json:"domains"`
	}

	return func(c *gin.Context) {
		c.JSON(http.StatusOK, Out{Domains: list.GetAll()})
	}
}

func domainListLookup(list *types.DomainList) func(*gin.Context) {
	return func(c *gin.Context) {
		domain := c.Param("domain")

		c.JSON(http.StatusOK, gin.H{
			"found": list.Lookup(domain),
		})
	}
}

func domainListUpsert(
	list *types.DomainList,
	db *database.Database,
	upsertFn func(ctx context.Context, db database.DBTX, domain string) error,
) func(*gin.Context) {
	type In struct {
		Domains []string `json:"domains"`
	}

	return func(c *gin.Context) {
		var in In
		if err := c.ShouldBindJSON(&in); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := list.Upsert(in.Domains); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, domain := range in.Domains {
			if err := upsertFn(c, db.DB, domain); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}

func domainListRemove(
	list *types.DomainList,
	db *database.Database,
	removeFn func(ctx context.Context, db database.DBTX, domain string) error,
) func(*gin.Context) {
	type In struct {
		Domains []string `json:"domains"`
	}

	return func(c *gin.Context) {
		var in In
		if err := c.ShouldBindJSON(&in); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := list.Remove(in.Domains); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, domain := range in.Domains {
			if err := removeFn(c, db.DB, domain); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}
