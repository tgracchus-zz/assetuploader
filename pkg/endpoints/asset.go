package endpoints

import (
	"net/http"
	"strconv"

	"github.com/tgracchus/assetuploader/pkg/auerr"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/tgracchus/assetuploader/pkg/assets"
)

const assetIDParam = "assetID"
const timeoutQueryParam = "timeout"

//RegisterAssetsEndpoints register to echo engine the assets endpoints.
func RegisterAssetsEndpoints(e *echo.Echo, assetManager assets.AssetManager, bucket string) {
	e.POST("/asset", newPostAssetEndpoint(assetManager, bucket))
	e.PUT("/asset/:"+assetIDParam, newPutAssetEndpoint(assetManager, bucket))
	e.GET("/asset/:"+assetIDParam, newGetAssetEndpoint(assetManager, bucket))
}

func newPostAssetEndpoint(assetManager assets.AssetManager, bucket string) func(c echo.Context) error {
	return func(c echo.Context) error {
		assetID := uuid.New()
		url, err := assetManager.PutURL(bucket, assetID)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, &postAssetResponse{UploadURL: url.String(), AssetID: assetID.String()})
	}
}

type postAssetResponse struct {
	UploadURL string `json:"upload_url"`
	AssetID   string `json:"id"`
}

func newPutAssetEndpoint(assetManager assets.AssetManager, bucket string) func(c echo.Context) error {
	return func(c echo.Context) error {
		assetID, err := uuid.Parse(c.Param(assetIDParam))
		if err != nil {
			return err
		}
		statusUpdate := new(putAssetBody)
		err = c.Bind(statusUpdate)
		if err != nil {
			return auerr.CError(auerr.ErrorBadInput, err)
		}

		if statusUpdate.Status != "uploaded" {
			return auerr.FError(auerr.ErrorBadInput, "Expected status uploaded, not %s", statusUpdate.Status)
		}

		err = assetManager.Uploaded(bucket, assetID)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusAccepted, &putAssetResponse{Status: "Accepted"})
	}
}

type putAssetBody struct {
	Status string `json:"Status"`
}

type putAssetResponse struct {
	Status string `json:"Status"`
}

func newGetAssetEndpoint(assetManager assets.AssetManager, bucket string) func(c echo.Context) error {
	return func(c echo.Context) error {
		assetID, err := uuid.Parse(c.Param(assetIDParam))
		if err != nil {
			return err
		}
		timeoutParam := c.QueryParam(timeoutQueryParam)
		if timeoutParam == "" {
			timeoutParam = "60"
		}
		timeout, err := strconv.ParseInt(timeoutParam, 10, 64)
		if err != nil {
			return err
		}
		url, err := assetManager.GetURL(bucket, assetID, timeout)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, &getAssetResponse{DownloadURL: url.String()})
	}
}

type getAssetResponse struct {
	DownloadURL string `json:"Download_url"`
}
