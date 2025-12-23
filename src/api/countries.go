package api

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

// GetBlackListCountries godoc
//
//	@Summary		Get blacklisted countries
//	@Description	get all blacklisted countries
//	@Tags			blacklist
//	@Produce		json
//	@Success		200	{object}	GetCountriesResp
//	@Router			/v1/blacklist/countries [get]
func GetBlackListCountries(blacklist *types.CountryList, mu *sync.Mutex) func(*gin.Context) {
	return countryListGetAll(blacklist, mu)
}

type GetCountriesResp struct {
	Countries []string `json:"countries" example:"fr,de"`
}

func countryListGetAll(list *types.CountryList, mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, GetCountriesResp{Countries: list.GetAll()})
	}
}

// CheckBlackListCountry godoc
//
//	@Summary		Check blacklisted country
//	@Description	check if a country is blacklisted
//	@Tags			blacklist
//	@Produce		json
//	@Param			country	path		string	true	"country to check"
//	@Success		200		{object}	CheckCountryResp
//	@Router			/v1/blacklist/countries/{country} [get]
func CheckBlackListCountry(blacklist *types.CountryList, mu *sync.Mutex) func(*gin.Context) {
	return countryListLookup(blacklist, mu)
}

type CheckCountryResp struct {
	Found bool `json:"found"`
}

func countryListLookup(list *types.CountryList, mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		country := c.Param("country")

		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, CheckCountryResp{
			Found: list.Lookup(country),
		})
	}
}

// UpsertCountriesReq godoc
//
//	@Summary		Upsert blacklisted countries
//	@Description	upsert countries to blacklist
//	@Tags			blacklist
//	@Accept			json
//	@Param			body	body	UpsertCountriesReq	true	"countries to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/blacklist/countries [post]
func UpsertBlackListCountries(blacklist *types.CountryList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return countryListUpsert(blacklist, mu, db, db.Q.UpsertBlackListCountry)
}

type UpsertCountriesReq struct {
	Countries []string `json:"countries" example:"fr,de"`
}

func countryListUpsert(
	list *types.CountryList,
	mu *sync.Mutex,
	db *database.Database,
	upsertFn func(ctx context.Context, db database.DBTX, country string) error,
) func(*gin.Context) {
	return func(c *gin.Context) {
		var req UpsertCountriesReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if err := list.Upsert(req.Countries); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, country := range req.Countries {
			if err := upsertFn(c, db.DB, country); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}

// RemoveBlackListCountries godoc
//
//	@Summary		Remove blacklisted countries
//	@Description	remove countries from blacklist
//	@Tags			blacklist
//	@Accept			json
//	@Param			body	body	RemoveCountriesReq	true	"countries to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/blacklist/countries [delete]
func RemoveBlackListCountries(blacklist *types.CountryList, mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return countryListRemove(blacklist, mu, db, db.Q.RemoveBlackListCountry)
}

type RemoveCountriesReq struct {
	Countries []string `json:"countries" example:"fr,de"`
}

func countryListRemove(
	list *types.CountryList,
	mu *sync.Mutex,
	db *database.Database,
	removeFn func(ctx context.Context, db database.DBTX, country string) error,
) func(*gin.Context) {
	return func(c *gin.Context) {
		var req RemoveCountriesReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if err := list.Remove(req.Countries); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, country := range req.Countries {
			if err := removeFn(c, db.DB, country); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}
