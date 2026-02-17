package api

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

// JA3IncludeListGet godoc
//
//	@Summary		Get included ja3 hashes
//	@Description	get all included ja3 hashes
//	@Tags			JA3
//	@Produce		json
//	@Success		200	{object}	GetJA3Resp
//	@Router			/v1/ja3/blocklist/include [get]
func JA3IncludeListGet(include *types.MapList[string], mu *sync.Mutex) func(*gin.Context) {
	return ja3ListGetAll(include, mu)
}

// JA3ExcludeListGet godoc
//
//	@Summary		Get excluded ja3 hashes
//	@Description	get all excluded ja3 hashes
//	@Tags			JA3
//	@Produce		json
//	@Success		200	{object}	GetJA3Resp
//	@Router			/v1/ja3/blocklist/exclude [get]
func JA3ExcludeListGet(exclude *types.MapList[string], mu *sync.Mutex) func(*gin.Context) {
	return ja3ListGetAll(exclude, mu)
}

type GetJA3Resp struct {
	Hashes []string `json:"hashes" example:"a50a861119aceb0ccc74902e8fddb618,534ce2dbc413c68e908363b5df0ae5e0"`
}

func ja3ListGetAll(list *types.MapList[string], mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, GetJA3Resp{Hashes: list.GetAll()})
	}
}

// JA3IncludeListCheck godoc
//
//	@Summary		Check included ja3 hash
//	@Description	check if a ja3 hash is included
//	@Tags			JA3
//	@Produce		json
//	@Param			hash	path		string	true	"hash to check"
//	@Success		200		{object}	CheckJA3Resp
//	@Router			/v1/ja3/blocklist/include/{hash} [get]
func JA3IncludeListCheck(include *types.MapList[string], mu *sync.Mutex) func(*gin.Context) {
	return ja3ListLookup(include, mu)
}

// JA3ExcludeListCheck godoc
//
//	@Summary		Check excluded ja3 hash
//	@Description	check if a ja3 hash is excluded
//	@Tags			JA3
//	@Produce		json
//	@Param			hash	path		string	true	"hash to check"
//	@Success		200		{object}	CheckJA3Resp
//	@Router			/v1/ja3/blocklist/exclude/{hash} [get]
func JA3ExcludeListCheck(exclude *types.MapList[string], mu *sync.Mutex) func(*gin.Context) {
	return ja3ListLookup(exclude, mu)
}

type CheckJA3Resp struct {
	Found bool `json:"found"`
}

func ja3ListLookup(list *types.MapList[string], mu *sync.Mutex) func(*gin.Context) {
	return func(c *gin.Context) {
		hash := c.Param("hash")

		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, CheckJA3Resp{
			Found: list.Lookup(hash),
		})
	}
}

// JA3IncludeListUpsert godoc
//
//	@Summary		Upsert included ja3 hashes
//	@Description	upsert ja3 hashes to includelist
//	@Tags			JA3
//	@Accept			json
//	@Param			body	body	UpsertJA3Req	true	"hashes to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/ja3/blocklist/include [post]
func JA3IncludeListUpsert(include *types.MapList[string], mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return ja3ListUpsert(include, mu, db, db.Q.UpsertJA3IncludeList)
}

// JA3ExcludeListUpsert godoc
//
//	@Summary		Upsert excluded ja3 hashes
//	@Description	upsert ja3 hashes to excludelist
//	@Tags			JA3
//	@Accept			json
//	@Param			body	body	UpsertJA3Req	true	"hashes to add"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/ja3/blocklist/exclude [post]
func JA3ExcludeListUpsert(exclude *types.MapList[string], mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return ja3ListUpsert(exclude, mu, db, db.Q.UpsertJA3ExcludeList)
}

type UpsertJA3Req struct {
	Hashes []string `json:"hashes" example:"a50a861119aceb0ccc74902e8fddb618,534ce2dbc413c68e908363b5df0ae5e0"`
}

func ja3ListUpsert(
	list *types.MapList[string],
	mu *sync.Mutex,
	db *database.Database,
	upsertFn func(ctx context.Context, db database.DBTX, hash string) error,
) func(*gin.Context) {
	return func(c *gin.Context) {
		var req UpsertJA3Req
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if err := list.Upsert(req.Hashes); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, hash := range req.Hashes {
			if err := upsertFn(c, db.DB, hash); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}

// JA3IncludeListRemove godoc
//
//	@Summary		Remove included ja3 hashes
//	@Description	remove ja3 hashes from includelist
//	@Tags			JA3
//	@Accept			json
//	@Param			body	body	RemoveJA3Req	true	"hashes to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/ja3/blocklist/include [delete]
func JA3IncludeListRemove(include *types.MapList[string], mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return ja3ListRemove(include, mu, db, db.Q.RemoveJA3IncludeList)
}

// JA3ExcludeListRemove godoc
//
//	@Summary		Remove excluded ja3 hashes
//	@Description	remove ja3 hashes from excludelist
//	@Tags			JA3
//	@Accept			json
//	@Param			body	body	RemoveJA3Req	true	"hashes to remove"
//	@Success		202
//	@Failure		400
//	@Failure		422
//	@Failure		500
//	@Router			/v1/ja3/blocklist/exclude [delete]
func JA3ExcludeListRemove(exclude *types.MapList[string], mu *sync.Mutex, db *database.Database) func(*gin.Context) {
	return ja3ListRemove(exclude, mu, db, db.Q.RemoveJA3ExcludeList)
}

type RemoveJA3Req struct {
	Hashes []string `json:"hashes" example:"a50a861119aceb0ccc74902e8fddb618,534ce2dbc413c68e908363b5df0ae5e0"`
}

func ja3ListRemove(
	list *types.MapList[string],
	mu *sync.Mutex,
	db *database.Database,
	removeFn func(ctx context.Context, db database.DBTX, hash string) error,
) func(*gin.Context) {
	return func(c *gin.Context) {
		var req RemoveJA3Req
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if err := list.Remove(req.Hashes); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		for _, hash := range req.Hashes {
			if err := removeFn(c, db.DB, hash); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusAccepted)
	}
}
