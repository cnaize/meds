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

func TestCountryBlockListHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCountries := []string{
		"fr",
		"de",
	}

	// create db mock
	qMock := database.NewMockQuerier(ctrl)
	qMock.EXPECT().UpsertCountryBlockList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testCountries) * 2)
	qMock.EXPECT().RemoveCountryBlockList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(testCountries))
	dbMock := &database.Database{
		Q: qMock,
	}

	// create router
	r := gin.Default()
	Register(r, dbMock, nil, types.NewCountryList(), nil, nil, nil, nil, nil, nil, nil, nil)

	// get all countries
	code, getRes, err := makeRequest[any, GetCountriesResp](r, http.MethodGet, "/v1/countries/blocklist", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Countries)

	// check country
	code, checkRes, err := makeRequest[any, CheckCountryResp](r, http.MethodGet, fmt.Sprintf("/v1/countries/blocklist/%s", testCountries[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)

	// upsert countries
	code, _, err = makeRequest[UpsertCountriesReq, any](r, http.MethodPost, "/v1/countries/blocklist", &UpsertCountriesReq{Countries: testCountries})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// upsert countries (second time)
	code, _, err = makeRequest[UpsertCountriesReq, any](r, http.MethodPost, "/v1/countries/blocklist", &UpsertCountriesReq{Countries: testCountries})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all countries (after upsert)
	code, getRes, err = makeRequest[any, GetCountriesResp](r, http.MethodGet, "/v1/countries/blocklist", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.ElementsMatch(t, testCountries, getRes.Countries)

	// check country (after upsert)
	code, checkRes, err = makeRequest[any, CheckCountryResp](r, http.MethodGet, fmt.Sprintf("/v1/countries/blocklist/%s", testCountries[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, checkRes.Found)

	// remove countries
	code, _, err = makeRequest[RemoveCountriesReq, any](r, http.MethodDelete, "/v1/countries/blocklist", &RemoveCountriesReq{Countries: testCountries})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, code)

	// get all countries (after remove)
	code, getRes, err = makeRequest[any, GetCountriesResp](r, http.MethodGet, "/v1/countries/blocklist", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, getRes.Countries)

	// check country (after remove)
	code, checkRes, err = makeRequest[any, CheckCountryResp](r, http.MethodGet, fmt.Sprintf("/v1/countries/blocklist/%s", testCountries[1]), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, checkRes.Found)
}
