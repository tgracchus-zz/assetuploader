package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/tgracchus/assetuploader/pkg/auerr"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestPostAsset(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/asset", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	putURL, err := url.Parse("http://ok")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("TestCreateAssetIDOK", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset")
		assetManager := &mockAssetManager{postURL: putURL, postErr: nil}
		post := newPostAssetEndpoint(assetManager, "testBucket")
		// Assertions
		if assert.NoError(t, post(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			//TODOassert.Equal(t, http.StatusCeated, rec.Body)
		}
	})
	t.Run("TestCreateAssetIDError", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset")
		assetManager := &mockAssetManager{postURL: nil, postErr: auerr.SError(auerr.ErrorInternalError, "ErrorInternalError")}
		post := newPostAssetEndpoint(assetManager, "testBucket")
		// Assertions
		err := post(c)
		if assert.Error(t, err) {
			AssetUploaderHTTPErrorHandler(err, c)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestPutAsset(t *testing.T) {
	// Setup
	e := echo.New()
	t.Run("TestPutOK", func(t *testing.T) {
		body, err := json.Marshal(&putAssetBody{Status: "uploaded"})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPut, "/asset", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset/:assetID")
		c.SetParamNames("assetID")
		c.SetParamValues(uuid.New().String())
		assetManager := &mockAssetManager{}
		put := newPutAssetEndpoint(assetManager, "testBucket")
		// Assertions
		if assert.NoError(t, put(c)) {
			assert.Equal(t, http.StatusAccepted, rec.Code)
		}
	})
	t.Run("TestPutNotFound", func(t *testing.T) {
		body, err := json.Marshal(&putAssetBody{Status: "uploaded"})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPut, "/asset/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset/:assetID")
		c.SetParamNames("assetID")
		c.SetParamValues(uuid.New().String())
		assetManager := &mockAssetManager{putErr: auerr.SError(auerr.ErrorNotFound, "ErrorNotFound")}
		put := newPutAssetEndpoint(assetManager, "testBucket")
		// Assertions
		err = put(c)
		if assert.Error(t, err) {
			AssetUploaderHTTPErrorHandler(err, c)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
	t.Run("TestPutAssetIDInvalida", func(t *testing.T) {
		body, err := json.Marshal(&putAssetBody{Status: "pa"})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPut, "/asset/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset/:assetID")
		c.SetParamNames("assetID")
		c.SetParamValues("invalidAssetID")
		assetManager := &mockAssetManager{}
		put := newPutAssetEndpoint(assetManager, "testBucket")
		// Assertions
		err = put(c)
		if assert.Error(t, err) {
			AssetUploaderHTTPErrorHandler(err, c)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("TestPutWrongStatus", func(t *testing.T) {
		body, err := json.Marshal(&putAssetBody{Status: "pa"})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPut, "/asset/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset/:assetID")
		c.SetParamNames("assetID")
		c.SetParamValues(uuid.New().String())
		assetManager := &mockAssetManager{}
		put := newPutAssetEndpoint(assetManager, "testBucket")
		// Assertions
		err = put(c)
		if assert.Error(t, err) {
			AssetUploaderHTTPErrorHandler(err, c)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("TestNoBody", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/asset/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset/:assetID")
		c.SetParamNames("assetID")
		c.SetParamValues(uuid.New().String())
		assetManager := &mockAssetManager{}
		put := newPutAssetEndpoint(assetManager, "testBucket")
		// Assertions
		err := put(c)
		if assert.Error(t, err) {
			AssetUploaderHTTPErrorHandler(err, c)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("TestError", func(t *testing.T) {
		body, err := json.Marshal(&putAssetBody{Status: "uploaded"})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPut, "/asset", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset/:assetID")
		c.SetParamNames("assetID")
		c.SetParamValues(uuid.New().String())
		assetManager := &mockAssetManager{putErr: auerr.SError(auerr.ErrorInternalError, "ErrorInternalError")}
		put := newPutAssetEndpoint(assetManager, "testBucket")
		// Assertions
		err = put(c)
		if assert.Error(t, err) {
			AssetUploaderHTTPErrorHandler(err, c)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestGetAsset(t *testing.T) {
	// Setup
	e := echo.New()
	t.Run("TestGetOK", func(t *testing.T) {
		getURL, err := url.Parse("http://ok")
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodGet, "/asset", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset/:assetID")
		c.SetParamNames("assetID")
		c.SetParamValues(uuid.New().String())
		assetManager := &mockAssetManager{getURL: getURL}
		get := newGetAssetEndpoint(assetManager, "testBucket")
		// Assertions
		if assert.NoError(t, get(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})
	t.Run("TestAssetIDNotCorrect", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/asset/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset/:assetID")
		c.SetParamNames("assetID")
		c.SetParamValues("nonValidUUID")
		assetManager := &mockAssetManager{}
		get := newGetAssetEndpoint(assetManager, "testBucket")
		// Assertions
		err := get(c)
		if assert.Error(t, err) {
			AssetUploaderHTTPErrorHandler(err, c)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("TestNotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/asset/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset/:assetID")
		c.SetParamNames("assetID")
		c.SetParamValues(uuid.New().String())
		assetManager := &mockAssetManager{getErr: auerr.SError(auerr.ErrorNotFound, "ErrorNotFound")}
		get := newGetAssetEndpoint(assetManager, "testBucket")
		// Assertions
		err := get(c)
		if assert.Error(t, err) {
			AssetUploaderHTTPErrorHandler(err, c)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
	t.Run("TestError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/asset/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/asset/:assetID")
		c.SetParamNames("assetID")
		c.SetParamValues(uuid.New().String())
		assetManager := &mockAssetManager{getErr: auerr.SError(auerr.ErrorInternalError, "ErrorInternalError")}
		get := newGetAssetEndpoint(assetManager, "testBucket")
		// Assertions
		err := get(c)
		if assert.Error(t, err) {
			AssetUploaderHTTPErrorHandler(err, c)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		}
	})
}

type mockAssetManager struct {
	postURL *url.URL
	postErr error
	putErr  error
	getURL  *url.URL
	getErr  error
}

func (mock *mockAssetManager) PutURL(ctx context.Context, bucket string, assetID uuid.UUID) (*url.URL, error) {
	return mock.postURL, mock.postErr
}
func (mock *mockAssetManager) Uploaded(ctx context.Context, bucket string, assetID uuid.UUID) error {
	return mock.putErr
}
func (mock *mockAssetManager) GetURL(ctx context.Context, bucket string, assetID uuid.UUID, timeout int64) (*url.URL, error) {
	return mock.getURL, mock.getErr
}
