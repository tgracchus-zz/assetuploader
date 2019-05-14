package endpoints

import (
	"net/http"

	"github.com/tgracchus/assertuploader/pkg/auerr"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

func AssetUploaderHTTPErrorHandler(err error, c echo.Context) {
	if err, ok := err.(*echo.HTTPError); ok {
		c.JSON(err.Code, err.Error())
	}
	switch code := errors.Cause(err).Error(); code {
	case auerr.ErrorBadInput:
		c.JSON(http.StatusBadRequest, &httpError{err.Error()})
	case auerr.ErrorConflict:
		c.JSON(http.StatusConflict, &httpError{err.Error()})
	case auerr.ErrorNotFound:
		c.JSON(http.StatusNotFound, &httpError{err.Error()})
	case auerr.ErrorInternalError:
		c.JSON(http.StatusInternalServerError, &httpError{err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, &httpError{err.Error()})
	}
	c.Logger().Errorf("%+v", err)
}

type httpError struct {
	Error string `json:"error"`
}
