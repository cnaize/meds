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

func TestSubnetAllowListHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testIPs := []string{
		"100.100.100.100",
		"200.200.200.0",
	}
	testSubnets := []string{
		"100.100.100.100/32",
		"200.200.200.0/24",
	}

	// create db mock
	qMock := database.NewMockQuerier(ctrl)
	qMock.EXPECT().UpsertIPAllowList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testSubnets) * 2)
	qMock.EXPECT().RemoveIPAllowList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testSubnets))
	dbMock := &database.Database{
		Q: qMock,
	}

	// create router
	r := gin.Default()
	Register(r, dbMock, types.NewIPList(), nil, nil, nil, nil, nil, nil, nil, nil, nil)

	// get all subnets
	code, getRes, err := makeRequest[any, GetSubnetsResp](r, http.MethodGet, "/v1/subnets/allowlist", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Subnets)

	// check ip address
	code, checkRes, err := makeRequest[any, CheckSubnetResp](r, http.MethodGet, fmt.Sprintf("/v1/subnets/allowlist/%s", testIPs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)

	// upsert subnets
	code, _, err = makeRequest[UpsertSubnetsReq, any](r, http.MethodPost, "/v1/subnets/allowlist", &UpsertSubnetsReq{Subnets: testSubnets})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// upsert subnets (second time)
	code, _, err = makeRequest[UpsertSubnetsReq, any](r, http.MethodPost, "/v1/subnets/allowlist", &UpsertSubnetsReq{Subnets: testSubnets})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all subnets (after upsert)
	code, getRes, err = makeRequest[any, GetSubnetsResp](r, http.MethodGet, "/v1/subnets/allowlist", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.ElementsMatch(t, testSubnets, getRes.Subnets)

	// check ip address (after upsert)
	code, checkRes, err = makeRequest[any, CheckSubnetResp](r, http.MethodGet, fmt.Sprintf("/v1/subnets/allowlist/%s", testIPs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, checkRes.Found)

	// remove subnets
	code, _, err = makeRequest[RemoveSubnetsReq, any](r, http.MethodDelete, "/v1/subnets/allowlist", &RemoveSubnetsReq{Subnets: testSubnets})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all subnets (after remove)
	code, getRes, err = makeRequest[any, GetSubnetsResp](r, http.MethodGet, "/v1/subnets/allowlist", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Subnets)

	// check ip address (after remove)
	code, checkRes, err = makeRequest[any, CheckSubnetResp](r, http.MethodGet, fmt.Sprintf("/v1/subnets/allowlist/%s", testIPs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)
}

func TestSubnetIncludeListHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testIPs := []string{
		"100.100.100.100",
		"200.200.200.0",
	}
	testSubnets := []string{
		"100.100.100.100/32",
		"200.200.200.0/24",
	}

	// create db mock
	qMock := database.NewMockQuerier(ctrl)
	qMock.EXPECT().UpsertIPIncludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testSubnets) * 2)
	qMock.EXPECT().RemoveIPIncludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testSubnets))
	dbMock := &database.Database{
		Q: qMock,
	}

	// create router
	r := gin.Default()
	Register(r, dbMock, nil, nil, types.NewIPList(), nil, nil, nil, nil, nil, nil, nil)

	// get all subnets
	code, getRes, err := makeRequest[any, GetSubnetsResp](r, http.MethodGet, "/v1/subnets/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Subnets)

	// check ip address
	code, checkRes, err := makeRequest[any, CheckSubnetResp](r, http.MethodGet, fmt.Sprintf("/v1/subnets/blocklist/include/%s", testIPs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)

	// upsert subnets
	code, _, err = makeRequest[UpsertSubnetsReq, any](r, http.MethodPost, "/v1/subnets/blocklist/include", &UpsertSubnetsReq{Subnets: testSubnets})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// upsert subnets (second time)
	code, _, err = makeRequest[UpsertSubnetsReq, any](r, http.MethodPost, "/v1/subnets/blocklist/include", &UpsertSubnetsReq{Subnets: testSubnets})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all subnets (after upsert)
	code, getRes, err = makeRequest[any, GetSubnetsResp](r, http.MethodGet, "/v1/subnets/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.ElementsMatch(t, testSubnets, getRes.Subnets)

	// check ip address (after upsert)
	code, checkRes, err = makeRequest[any, CheckSubnetResp](r, http.MethodGet, fmt.Sprintf("/v1/subnets/blocklist/include/%s", testIPs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, checkRes.Found)

	// remove subnets
	code, _, err = makeRequest[RemoveSubnetsReq, any](r, http.MethodDelete, "/v1/subnets/blocklist/include", &RemoveSubnetsReq{Subnets: testSubnets})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all subnets (after remove)
	code, getRes, err = makeRequest[any, GetSubnetsResp](r, http.MethodGet, "/v1/subnets/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Subnets)

	// check ip address (after remove)
	code, checkRes, err = makeRequest[any, CheckSubnetResp](r, http.MethodGet, fmt.Sprintf("/v1/subnets/blocklist/include/%s", testIPs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)
}

func TestSubnetExcludeListHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testIPs := []string{
		"100.100.100.100",
		"200.200.200.0",
	}
	testSubnets := []string{
		"100.100.100.100/32",
		"200.200.200.0/24",
	}

	// create db mock
	qMock := database.NewMockQuerier(ctrl)
	qMock.EXPECT().UpsertIPExcludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testSubnets) * 2)
	qMock.EXPECT().RemoveIPExcludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testSubnets))
	dbMock := &database.Database{
		Q: qMock,
	}

	// create router
	r := gin.Default()
	Register(r, dbMock, nil, nil, nil, types.NewIPList(), nil, nil, nil, nil, nil, nil)

	// get all subnets
	code, getRes, err := makeRequest[any, GetSubnetsResp](r, http.MethodGet, "/v1/subnets/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Subnets)

	// check ip address
	code, checkRes, err := makeRequest[any, CheckSubnetResp](r, http.MethodGet, fmt.Sprintf("/v1/subnets/blocklist/exclude/%s", testIPs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)

	// upsert subnets
	code, _, err = makeRequest[UpsertSubnetsReq, any](r, http.MethodPost, "/v1/subnets/blocklist/exclude", &UpsertSubnetsReq{Subnets: testSubnets})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// upsert subnets (second time)
	code, _, err = makeRequest[UpsertSubnetsReq, any](r, http.MethodPost, "/v1/subnets/blocklist/exclude", &UpsertSubnetsReq{Subnets: testSubnets})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all subnets (after upsert)
	code, getRes, err = makeRequest[any, GetSubnetsResp](r, http.MethodGet, "/v1/subnets/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.ElementsMatch(t, testSubnets, getRes.Subnets)

	// check ip address (after upsert)
	code, checkRes, err = makeRequest[any, CheckSubnetResp](r, http.MethodGet, fmt.Sprintf("/v1/subnets/blocklist/exclude/%s", testIPs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, checkRes.Found)

	// remove subnets
	code, _, err = makeRequest[RemoveSubnetsReq, any](r, http.MethodDelete, "/v1/subnets/blocklist/exclude", &RemoveSubnetsReq{Subnets: testSubnets})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all subnets (after remove)
	code, getRes, err = makeRequest[any, GetSubnetsResp](r, http.MethodGet, "/v1/subnets/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Subnets)

	// check ip address (after remove)
	code, checkRes, err = makeRequest[any, CheckSubnetResp](r, http.MethodGet, fmt.Sprintf("/v1/subnets/blocklist/exclude/%s", testIPs[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)
}
