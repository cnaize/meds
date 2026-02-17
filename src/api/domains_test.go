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

func TestDomainIncludeListHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDomains := []string{
		"good.com",
		"bad.com",
	}

	// create db mock
	qMock := database.NewMockQuerier(ctrl)
	qMock.EXPECT().UpsertDomainIncludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testDomains) * 2)
	qMock.EXPECT().RemoveDomainIncludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testDomains))
	dbMock := &database.Database{
		Q: qMock,
	}

	// create router
	r := gin.Default()
	Register(r, dbMock, nil, nil, nil, nil, nil, nil, nil, nil, types.NewDomainList(), nil)

	// get all domains
	code, getRes, err := makeRequest[any, GetDomainsResp](r, http.MethodGet, "/v1/domains/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Domains)

	// check domain
	code, checkRes, err := makeRequest[any, CheckDomainResp](r, http.MethodGet, fmt.Sprintf("/v1/domains/blocklist/include/%s", testDomains[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)

	// upsert domains
	code, _, err = makeRequest[UpsertDomainsReq, any](r, http.MethodPost, "/v1/domains/blocklist/include", &UpsertDomainsReq{Domains: testDomains})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// upsert domains (second time)
	code, _, err = makeRequest[UpsertDomainsReq, any](r, http.MethodPost, "/v1/domains/blocklist/include", &UpsertDomainsReq{Domains: testDomains})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all domains (after upsert)
	code, getRes, err = makeRequest[any, GetDomainsResp](r, http.MethodGet, "/v1/domains/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.ElementsMatch(t, testDomains, getRes.Domains)

	// check domain (after upsert)
	code, checkRes, err = makeRequest[any, CheckDomainResp](r, http.MethodGet, fmt.Sprintf("/v1/domains/blocklist/include/%s", testDomains[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, checkRes.Found)

	// remove domains
	code, _, err = makeRequest[RemoveDomainsReq, any](r, http.MethodDelete, "/v1/domains/blocklist/include", &RemoveDomainsReq{Domains: testDomains})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all domains (after remove)
	code, getRes, err = makeRequest[any, GetDomainsResp](r, http.MethodGet, "/v1/domains/blocklist/include", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Domains)

	// check domain (after remove)
	code, checkRes, err = makeRequest[any, CheckDomainResp](r, http.MethodGet, fmt.Sprintf("/v1/domains/blocklist/include/%s", testDomains[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)
}

func TestDomainExcludeListHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDomains := []string{
		"good.com",
		"bad.com",
	}

	// create db mock
	qMock := database.NewMockQuerier(ctrl)
	qMock.EXPECT().UpsertDomainExcludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testDomains) * 2)
	qMock.EXPECT().RemoveDomainExcludeList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testDomains))
	dbMock := &database.Database{
		Q: qMock,
	}

	// create router
	r := gin.Default()
	Register(r, dbMock, nil, nil, nil, nil, nil, nil, nil, nil, nil, types.NewDomainList())

	// get all domains
	code, getRes, err := makeRequest[any, GetDomainsResp](r, http.MethodGet, "/v1/domains/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Domains)

	// check domain
	code, checkRes, err := makeRequest[any, CheckDomainResp](r, http.MethodGet, fmt.Sprintf("/v1/domains/blocklist/exclude/%s", testDomains[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)

	// upsert domains
	code, _, err = makeRequest[UpsertDomainsReq, any](r, http.MethodPost, "/v1/domains/blocklist/exclude", &UpsertDomainsReq{Domains: testDomains})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// upsert domains (second time)
	code, _, err = makeRequest[UpsertDomainsReq, any](r, http.MethodPost, "/v1/domains/blocklist/exclude", &UpsertDomainsReq{Domains: testDomains})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all domains (after upsert)
	code, getRes, err = makeRequest[any, GetDomainsResp](r, http.MethodGet, "/v1/domains/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.ElementsMatch(t, testDomains, getRes.Domains)

	// check domain (after upsert)
	code, checkRes, err = makeRequest[any, CheckDomainResp](r, http.MethodGet, fmt.Sprintf("/v1/domains/blocklist/exclude/%s", testDomains[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, checkRes.Found)

	// remove domains
	code, _, err = makeRequest[RemoveDomainsReq, any](r, http.MethodDelete, "/v1/domains/blocklist/exclude", &RemoveDomainsReq{Domains: testDomains})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all domains (after remove)
	code, getRes, err = makeRequest[any, GetDomainsResp](r, http.MethodGet, "/v1/domains/blocklist/exclude", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Domains)

	// check domain (after remove)
	code, checkRes, err = makeRequest[any, CheckDomainResp](r, http.MethodGet, fmt.Sprintf("/v1/domains/blocklist/exclude/%s", testDomains[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)
}
