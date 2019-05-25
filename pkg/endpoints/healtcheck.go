package endpoints

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/labstack/echo"
)

//RegisterHealthCheck register to echo engine a healthcheck endpoint.
func RegisterHealthCheck(e *echo.Echo, sess *session.Session) {
	e.GET("/healthcheck", newHealthCheck(sess))
}

func newHealthCheck(sess *session.Session) func(c echo.Context) error {
	return func(c echo.Context) error {
		status := "DOWN"
		statusCode := http.StatusServiceUnavailable
		ticker := time.NewTicker(5 * time.Second)
		go func() {
			for range ticker.C {
				session.Session()
			}
		}()

		return c.JSON(statusCode, &healthcheck{Status: status})
	}
}

type healthcheck struct {
	Status string `json:"Status"`
}
