package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/types"
)

func TestJA3IncludeListHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testHashes := []string{
		"a50a861119aceb0ccc74902e8fddb618",
		"534ce2dbc413c68e908363b5df0ae5e0",
	}

	// create db mock
	qMock := database.NewMockQuerier(ctrl)
	qMock.EXPECT().UpsertJA3IncludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testHashes) * 2)
	qMock.EXPECT().RemoveJA3IncludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testHashes))
	dbMock := &database.Database{
		Q: qMock,
	}

	// create router
	r := gin.Default()
	Register(r, dbMock, nil, nil, nil, nil, nil, nil, types.NewMapList[string](), nil, nil, nil)

	// get all hashes
	code, getRes, err := makeRequest[any, GetJA3Resp](r, http.MethodGet, "/v1/ja3/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Hashes)

	// check hash
	code, checkRes, err := makeRequest[any, CheckJA3Resp](r, http.MethodGet, fmt.Sprintf("/v1/ja3/blocklist/include/%s", testHashes[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)

	// upsert hashes
	code, _, err = makeRequest[UpsertJA3Req, any](r, http.MethodPost, "/v1/ja3/blocklist/include", &UpsertJA3Req{Hashes: testHashes})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// upsert hashes (second time)
	code, _, err = makeRequest[UpsertJA3Req, any](r, http.MethodPost, "/v1/ja3/blocklist/include", &UpsertJA3Req{Hashes: testHashes})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all hashes (after upsert)
	code, getRes, err = makeRequest[any, GetJA3Resp](r, http.MethodGet, "/v1/ja3/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.ElementsMatch(t, testHashes, getRes.Hashes)

	// check hash (after upsert)
	code, checkRes, err = makeRequest[any, CheckJA3Resp](r, http.MethodGet, fmt.Sprintf("/v1/ja3/blocklist/include/%s", testHashes[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, checkRes.Found)

	// remove hashes
	code, _, err = makeRequest[RemoveJA3Req, any](r, http.MethodDelete, "/v1/ja3/blocklist/include", &RemoveJA3Req{Hashes: testHashes})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all hashes (after remove)
	code, getRes, err = makeRequest[any, GetJA3Resp](r, http.MethodGet, "/v1/ja3/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Hashes)

	// check hash (after remove)
	code, checkRes, err = makeRequest[any, CheckJA3Resp](r, http.MethodGet, fmt.Sprintf("/v1/ja3/blocklist/include/%s", testHashes[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)
}

func TestJA3ExcludeListHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testHashes := []string{
		"a50a861119aceb0ccc74902e8fddb618",
		"534ce2dbc413c68e908363b5df0ae5e0",
	}

	// create db mock
	qMock := database.NewMockQuerier(ctrl)
	qMock.EXPECT().UpsertJA3ExcludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testHashes) * 2)
	qMock.EXPECT().RemoveJA3ExcludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testHashes))
	dbMock := &database.Database{
		Q: qMock,
	}

	// create router
	r := gin.Default()
	Register(r, dbMock, nil, nil, nil, nil, nil, nil, nil, types.NewMapList[string](), nil, nil)

	// get all hashes
	code, getRes, err := makeRequest[any, GetJA3Resp](r, http.MethodGet, "/v1/ja3/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Hashes)

	// check hash
	code, checkRes, err := makeRequest[any, CheckJA3Resp](r, http.MethodGet, fmt.Sprintf("/v1/ja3/blocklist/exclude/%s", testHashes[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)

	// upsert hashes
	code, _, err = makeRequest[UpsertJA3Req, any](r, http.MethodPost, "/v1/ja3/blocklist/exclude", &UpsertJA3Req{Hashes: testHashes})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// upsert hashes (second time)
	code, _, err = makeRequest[UpsertJA3Req, any](r, http.MethodPost, "/v1/ja3/blocklist/exclude", &UpsertJA3Req{Hashes: testHashes})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all hashes (after upsert)
	code, getRes, err = makeRequest[any, GetJA3Resp](r, http.MethodGet, "/v1/ja3/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.ElementsMatch(t, testHashes, getRes.Hashes)

	// check hash (after upsert)
	code, checkRes, err = makeRequest[any, CheckJA3Resp](r, http.MethodGet, fmt.Sprintf("/v1/ja3/blocklist/exclude/%s", testHashes[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, checkRes.Found)

	// remove hashes
	code, _, err = makeRequest[RemoveJA3Req, any](r, http.MethodDelete, "/v1/ja3/blocklist/exclude", &RemoveJA3Req{Hashes: testHashes})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all hashes (after remove)
	code, getRes, err = makeRequest[any, GetJA3Resp](r, http.MethodGet, "/v1/ja3/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Hashes)

	// check hash (after remove)
	code, checkRes, err = makeRequest[any, CheckJA3Resp](r, http.MethodGet, fmt.Sprintf("/v1/ja3/blocklist/exclude/%s", testHashes[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)
}
