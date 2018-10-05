package route

import (
	"github.com/labstack/echo"
	apihttp "github.com/mikitu/gocommerce/api/src/http"
	"github.com/mikitu/gocommerce/api/src/middleware"
	"net/http"
)

func ImportBackendRoutes(e *echo.Echo) {
	userPoolId := "us-east-1_cPPkedcYY" //os.Getenv("COGNITO_USER_POOL_ID")
	auth := middleware.CognitoAuthWithConfig(middleware.CognitoAuthConfig{
		AwsRegion: "us-east-1",
		UserPoolId: userPoolId,
	})

	admin := e.Group("/admin", auth)
	admin.GET("/", func(c echo.Context) error {
		data := map[string]string{"message": "Hello Admin"}
		return c.JSON(http.StatusOK, &apihttp.ResponseFormatter{Status: http.StatusOK, Data: data, Errors: nil})
	})

}
