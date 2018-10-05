package route

import (
	"github.com/labstack/echo"
	apihttp "github.com/mikitu/gocommerce/api/src/http"
	"net/http"
)

func ImportFrontendRoutes(e *echo.Echo) {
	e.GET("/", func(c echo.Context) error {
		data := map[string]string{"message": "Hello World"}
		return c.JSON(http.StatusOK, &apihttp.ResponseFormatter{Status: http.StatusOK, Data: data, Errors: nil})
	})
}
