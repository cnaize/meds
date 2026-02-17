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

func TestASNIncludeListHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testASNs := []uint32{
		123,
		456,
		789,
	}

	// create db mock
	qMock := database.NewMockQuerier(ctrl)
	qMock.EXPECT().UpsertASNIncludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testASNs) * 2)
	qMock.EXPECT().RemoveASNIncludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testASNs))
	dbMock := &database.Database{
		Q: qMock,
	}

	// create router
	r := gin.Default()
	Register(r, dbMock, nil, nil, nil, nil, types.NewMapList[uint32](), nil, nil, nil, nil, nil)

	// get all asns
	code, getRes, err := makeRequest[any, GetASNsResp](r, http.MethodGet, "/v1/asns/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.ASNs)

	// check ans
	code, checkRes, err := makeRequest[any, CheckASNResp](r, http.MethodGet, fmt.Sprintf("/v1/asns/blocklist/include/%d", testASNs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)

	// upsert asns
	code, _, err = makeRequest[UpsertASNsReq, any](r, http.MethodPost, "/v1/asns/blocklist/include", &UpsertASNsReq{ASNs: testASNs})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// upsert asns (second time)
	code, _, err = makeRequest[UpsertASNsReq, any](r, http.MethodPost, "/v1/asns/blocklist/include", &UpsertASNsReq{ASNs: testASNs})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all asns (after upsert)
	code, getRes, err = makeRequest[any, GetASNsResp](r, http.MethodGet, "/v1/asns/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.ElementsMatch(t, testASNs, getRes.ASNs)

	// check ans (after upsert)
	code, checkRes, err = makeRequest[any, CheckASNResp](r, http.MethodGet, fmt.Sprintf("/v1/asns/blocklist/include/%d", testASNs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, checkRes.Found)

	// remove asns
	code, _, err = makeRequest[RemoveASNsReq, any](r, http.MethodDelete, "/v1/asns/blocklist/include", &RemoveASNsReq{ASNs: testASNs})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all asns (after remove)
	code, getRes, err = makeRequest[any, GetASNsResp](r, http.MethodGet, "/v1/asns/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.ASNs)

	// check ans (after remove)
	code, checkRes, err = makeRequest[any, CheckASNResp](r, http.MethodGet, fmt.Sprintf("/v1/asns/blocklist/include/%d", testASNs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)
}

func TestASNExcludeListHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testASNs := []uint32{
		123,
		456,
		789,
	}

	// create db mock
	qMock := database.NewMockQuerier(ctrl)
	qMock.EXPECT().UpsertASNExcludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testASNs) * 2)
	qMock.EXPECT().RemoveASNExcludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testASNs))
	dbMock := &database.Database{
		Q: qMock,
	}

	// create router
	r := gin.Default()
	Register(r, dbMock, nil, nil, nil, nil, nil, types.NewMapList[uint32](), nil, nil, nil, nil)

	// get all asns
	code, getRes, err := makeRequest[any, GetASNsResp](r, http.MethodGet, "/v1/asns/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.ASNs)

	// check ans
	code, checkRes, err := makeRequest[any, CheckASNResp](r, http.MethodGet, fmt.Sprintf("/v1/asns/blocklist/exclude/%d", testASNs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)

	// upsert asns
	code, _, err = makeRequest[UpsertASNsReq, any](r, http.MethodPost, "/v1/asns/blocklist/exclude", &UpsertASNsReq{ASNs: testASNs})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// upsert asns (second time)
	code, _, err = makeRequest[UpsertASNsReq, any](r, http.MethodPost, "/v1/asns/blocklist/exclude", &UpsertASNsReq{ASNs: testASNs})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all asns (after upsert)
	code, getRes, err = makeRequest[any, GetASNsResp](r, http.MethodGet, "/v1/asns/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.ElementsMatch(t, testASNs, getRes.ASNs)

	// check ans (after upsert)
	code, checkRes, err = makeRequest[any, CheckASNResp](r, http.MethodGet, fmt.Sprintf("/v1/asns/blocklist/exclude/%d", testASNs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, checkRes.Found)

	// remove asns
	code, _, err = makeRequest[RemoveASNsReq, any](r, http.MethodDelete, "/v1/asns/blocklist/exclude", &RemoveASNsReq{ASNs: testASNs})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all asns (after remove)
	code, getRes, err = makeRequest[any, GetASNsResp](r, http.MethodGet, "/v1/asns/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.ASNs)

	// check ans (after remove)
	code, checkRes, err = makeRequest[any, CheckASNResp](r, http.MethodGet, fmt.Sprintf("/v1/asns/blocklist/exclude/%d", testASNs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)
}
