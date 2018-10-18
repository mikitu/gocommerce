package main

import (
	"fmt"
	"github.com/labstack/echo"
	md "github.com/labstack/echo/middleware"
	"github.com/mikitu/gocommerce/api/src/route"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config/")
	viper.SetConfigType("yml")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	e := echo.New()
	e.Pre(md.AddTrailingSlash())
	e.Use(md.CORS());

	route.ImportBackendRoutes(e)
	route.ImportFrontendRoutes(e)
	e.Logger.Fatal(e.Start(":1323"))
}

