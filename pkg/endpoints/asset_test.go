package endpoints_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	h "github.com/golang/dep/hack/licenseok"
	"github.com/labstack/echo"
	"github.com/tgracchus/assetuploader/pkg/endpoints"
)

func TestPutAsset(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/asset", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	endpoints.RegisterAssetsEndpoints(e,,"testbucket")

	// Assertions
	if assert.NoError(t, h.createUser(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, userJSON, rec.Body.String())
	}
}

type mockAssetManager struct{}
