package route

import (
	"github.com/labstack/echo"
	"github.com/mikitu/gocommerce/api/src/handlers"
	apihttp "github.com/mikitu/gocommerce/api/src/http"
	"github.com/mikitu/gocommerce/api/src/middleware"
	"github.com/spf13/viper"
	"net/http"
)

func ImportBackendRoutes(e *echo.Echo) {
	//userPoolId := viper.GetString("userPoolId") //os.Getenv("COGNITO_USER_POOL_ID")
	auth := middleware.CognitoAuthWithConfig(middleware.CognitoAuthConfig{
		AwsRegion: viper.GetString("aws_region"),
		UserPoolId: viper.GetString("userPoolId"),
		Skipper: func(c echo.Context) bool {
			// Skip authentication for login requests
			if c.Path() == "/admin/login/" {
				return true
			}
			return false
		},
	})

	admin := e.Group("/admin", auth)
	admin.GET("/", func(c echo.Context) error {
		data := map[string]string{"message": "Hello Admin"}
		return c.JSON(http.StatusOK, &apihttp.ResponseFormatter{Status: http.StatusOK, Data: data, Errors: nil})
	})

	admin.POST("/login/", handlers.AdminLoginHandler)

}
