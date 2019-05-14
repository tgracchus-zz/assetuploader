package endpoints

import (
	"net/http"

	"github.com/tgracchus/assertuploader/pkg/auerr"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

func AssetUploaderHTTPErrorHandler(err error, c echo.Context) {
	c.Logger().Error(err)

	switch code := errors.Cause(err).Error(); code {
	case auerr.ErrorBadInput:
		c.JSON(http.StatusBadRequest, err.Error())
	case auerr.ErrorConflict:
		c.JSON(http.StatusConflict, err.Error())
	case auerr.ErrorNotFound:
		c.JSON(http.StatusNotFound, err.Error())
	case auerr.ErrorInternalError:
		c.JSON(http.StatusInternalServerError, "Internal Server Error")
	default:
		c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

}

type httpError struct {
	Error string `json:"error"`
}
