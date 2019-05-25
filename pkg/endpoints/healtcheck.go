package endpoints

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/labstack/echo"
)

//RegisterHealthCheck register to echo engine a healthcheck endpoint.
func RegisterHealthCheck(e *echo.Echo, svc *s3.S3, bucket string) {
	e.GET("/healthcheck", newHealthCheck(svc, bucket))
}

func newHealthCheck(svc *s3.S3, bucket string) func(c echo.Context) error {
	queries := make(chan bool)
	status := make(chan healthcheck)
	go func() {
		defer close(status)
		defer close(queries)
		ticker := time.NewTicker(5 * time.Second)
		check := healthcheck{
			Status:     "DOWN",
			StatusCode: http.StatusServiceUnavailable,
		}
		for {
			select {
			case <-ticker.C:
				_, err := svc.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: aws.String(bucket)})
				if err == nil {
					check = healthcheck{
						Status:     "UP",
						StatusCode: http.StatusOK,
					}
				} else {
					check = healthcheck{
						Status:     "DOWN",
						StatusCode: http.StatusServiceUnavailable,
					}
				}
			case <-queries:
				status <- check
			}
			if status == nil || queries == nil {
				panic("status or queries closed")
			}
		}
	}()

	return func(c echo.Context) error {
		queries <- true
		select {
		case <-c.Request().Context().Done():
			return c.Request().Context().Err()
		case check := <-status:
			return c.JSON(check.StatusCode, &healthcheck{Status: check.Status})
		}
	}
}

type healthcheck struct {
	Status     string `json:"Status"`
	StatusCode int    `json:"-"`
}
